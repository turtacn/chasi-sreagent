package localai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/turtacn/chasi-sreagent/pkg/common/errors"
	"github.com/turtacn/chasi-sreagent/pkg/common/log"
	"github.com/turtacn/chasi-sreagent/pkg/common/types"
	"github.com/turtacn/chasi-sreagent/pkg/framework/llm"
	"go.uber.org/zap"
)

// Package localai provides an LLM implementation for LocalAI.
// 包 localai 提供一个用于 LocalAI 的 LLM 实现。

// LocalAIProvider implements the llm.LLM interface for the LocalAI API.
// LocalAIProvider 为 LocalAI API 实现 llm.LLM 接口。
type LocalAIProvider struct {
	config *types.LLMProviderConfig
	client *http.Client
}

// Ensure LocalAIProvider implements the llm.LLM interface.
// 确保 LocalAIProvider 实现了 llm.LLM 接口。
var _ llm.LLM = &LocalAIProvider{}

// NewLocalAIProvider creates a new LocalAIProvider instance.
// NewLocalAIProvider 创建一个新的 LocalAIProvider 实例。
func NewLocalAIProvider(cfg *types.LLMProviderConfig, timeout time.Duration) (*LocalAIProvider, error) {
	if cfg == nil || cfg.URL == "" || cfg.Model == "" {
		return nil, fmt.Errorf("localai provider configuration is incomplete")
	}

	// Basic HTTP client
	// 基本 HTTP 客户端
	httpClient := &http.Client{
		Timeout: timeout, // Use the global LLM timeout / 使用全局 LLM 超时时间
	}

	log.L().Info("Initialized LocalAI LLM Provider", zap.String("url", cfg.URL), zap.String("model", cfg.Model))

	return &LocalAIProvider{
		config: cfg,
		client: httpClient,
	}, nil
}

// Name returns the name of the LLM provider.
// Name 返回 LLM 提供商的名称。
func (p *LocalAIProvider) Name() string {
	return "localai" // Using the constant defined in common.types
}

// Description returns a brief description.
// Description 返回简要描述。
func (p *LocalAIProvider) Description() string {
	return "Provides access to a LocalAI compatible API."
}

// GenerateText sends a completion request to the LocalAI API.
// GenerateText 向 LocalAI API 发送一个 completions 请求。
// Assumes the prompt is a simple string for the 'prompt' field in the request body.
// 假设 prompt 是请求体中 'prompt' 字段的简单字符串。
// Options map can be used to pass additional parameters like temperature, max_tokens.
// Options map 可用于传递其他参数，如 temperature, max_tokens。
func (p *LocalAIProvider) GenerateText(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
	logger := log.LWithContext(ctx).With(zap.String("llmProvider", p.Name()), zap.String("model", p.config.Model))
	logger.Debug("Generating text with LocalAI")

	if p.client == nil {
		return "", fmt.Errorf("localai client is not initialized")
	}

	// Construct the request body based on LocalAI's API spec (often OpenAI compatible)
	// 根据 LocalAI 的 API 规范构建请求体 (通常与 OpenAI 兼容)
	requestBody := map[string]interface{}{
		"model":       p.config.Model,
		"prompt":      prompt,
		"max_tokens":  500, // Default max tokens, can be overridden by options
		"temperature": 0.7, // Default temperature, can be overridden by options
		// Add other parameters from options
		// 添加 options 中的其他参数
	}

	// Override defaults with options
	// 使用 options 覆盖默认值
	if options != nil {
		for key, value := range options {
			requestBody[key] = value
		}
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		logger.Error("Failed to marshal request body", zap.Error(err))
		return "", errors.Wrap(errors.ErrorCodeLLMProviderError, "failed to marshal request body", err, "")
	}

	// Construct the request URL (assuming /v1/completions or /v1/chat/completions)
	// 构建请求 URL (假设 /v1/completions 或 /v1/chat/completions)
	// LocalAI often supports /v1/chat/completions for chat models
	// LocalAI 通常支持 /v1/chat/completions 用于 chat 模型
	// For simplicity, let's use /v1/completions for now, adjust if using chat models.
	// 为简化起见，我们暂时使用 /v1/completions，如果使用 chat 模型请调整。
	reqURL := fmt.Sprintf("%s/completions", p.config.URL)

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		logger.Error("Failed to create HTTP request", zap.Error(err))
		return "", errors.Wrap(errors.ErrorCodeLLMProviderError, "failed to create http request", err, "")
	}

	req.Header.Set("Content-Type", "application/json")
	if p.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		logger.Error("HTTP request failed", zap.Error(err))
		return "", errors.Wrap(errors.ErrorCodeLLMProviderError, "http request failed", err, "")
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response body", zap.Error(err))
		return "", errors.Wrap(errors.ErrorCodeLLMProviderError, "failed to read response body", err, "")
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("API returned non-200 status", zap.Int("statusCode", resp.StatusCode), zap.ByteString("body", bodyBytes))
		return "", errors.New(errors.ErrorCodeLLMProviderError, fmt.Sprintf("API returned status %d", resp.StatusCode), string(bodyBytes))
	}

	// Parse the response body (assuming OpenAI completions format)
	// 解析响应体 (假设 OpenAI completions 格式)
	var response map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		logger.Error("Failed to unmarshal response body", zap.Error(err), zap.ByteString("body", bodyBytes))
		return "", errors.Wrap(errors.ErrorCodeLLMProviderError, "failed to unmarshal response body", err, string(bodyBytes))
	}

	// Extract the generated text. Path depends on API version/endpoint (/v1/completions vs /v1/chat/completions)
	// 提取生成的文本。路径取决于 API 版本/终点 (/v1/completions vs /v1/chat/completions)
	// Assuming /v1/completions -> choices[0].text
	// 假设 /v1/completions -> choices[0].text
	choices, ok := response["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", errors.New(errors.ErrorCodeLLMProviderError, "no choices found in LLM response", string(bodyBytes))
	}
	firstChoice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", errors.New(errors.ErrorCodeLLMProviderError, "invalid choice format in LLM response", string(bodyBytes))
	}
	generatedText, ok := firstChoice["text"].(string) // For completions
	// For chat completions: content, ok := firstChoice["message"].(map[string]interface{})["content"].(string)
	if !ok {
		// Try chat completions format
		// 尝试 chat completions 格式
		message, msgOK := firstChoice["message"].(map[string]interface{})
		if msgOK {
			generatedText, ok = message["content"].(string)
		}
		if !ok {
			return "", errors.New(errors.ErrorCodeLLMProviderError, "generated text or content not found in LLM response", string(bodyBytes))
		}
	}

	logger.Debug("Text generated successfully")
	return generatedText, nil
}

// // Placeholder for EmbedText if needed
// // 如果需要，EmbedText 的占位符
// func (p *LocalAIProvider) EmbedText(ctx context.Context, text string) ([]float32, error) {
// 	logger := log.LWithContext(ctx).With(zap.String("llmProvider", p.Name()))
// 	logger.Warn("LocalAIProvider.EmbedText not implemented")
// 	// TODO: Implement embedding call to LocalAI / TODO: 实现对 LocalAI 的 embedding 调用
// 	return nil, fmt.Errorf("embedding not implemented for LocalAI provider")
// }

// Register the LLM provider with the global registry.
// 在全局注册表中注册 LLM 提供商。
func init() {
	// This stateful provider needs configuration and potentially a timeout
	// before registration. We will rely on the engine initialization to create and register it.
	// 这个有状态的提供商在注册之前需要配置和可能的超时时间。
	// 我们将依赖于引擎初始化来创建和注册它。

	// llm.RegisterLLMProvider(&LocalAIProvider{}) // Cannot register like this
	log.L().Debug("LocalAI LLM provider init() called, will be registered during engine setup.")
}

// Global instance placeholder, will be initialized in main.
// 全局实例占位符，将在 main 中初始化。
var LocalAIProviderInstance *LocalAIProvider

// RegisterLocalAIProvider registers the initialized LocalAIProvider instance.
// RegisterLocalAIProvider 注册已初始化的 LocalAIProvider 实例。
// This should be called after NewLocalAIProvider is successful.
// 应在 NewLocalAIProvider 成功后调用此函数。
func RegisterLocalAIProvider(provider *LocalAIProvider) {
	llm.RegisterLLMProvider(provider)
	LocalAIProviderInstance = provider // Keep a global reference if needed elsewhere
}
