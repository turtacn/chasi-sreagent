package deepseek

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

// Package deepseek provides an LLM implementation for the DeepSeek API.
// 包 deepseek 提供一个用于 DeepSeek API 的 LLM 实现。

// DeepSeekProvider implements the llm.LLM interface for the DeepSeek API.
// DeepSeekProvider 为 DeepSeek API 实现 llm.LLM 接口。
type DeepSeekProvider struct {
	config *types.LLMProviderConfig
	client *http.Client
}

// Ensure DeepSeekProvider implements the llm.LLM interface.
// 确保 DeepSeekProvider 实现了 llm.LLM 接口。
var _ llm.LLM = &DeepSeekProvider{}

// NewDeepSeekProvider creates a new DeepSeekProvider instance.
// NewDeepSeekProvider 创建一个新的 DeepSeekProvider 实例。
func NewDeepSeekProvider(cfg *types.LLMProviderConfig, timeout time.Duration) (*DeepSeekProvider, error) {
	if cfg == nil || cfg.URL == "" || cfg.Model == "" || cfg.APIKey == "" {
		return nil, fmt.Errorf("deepseek provider configuration is incomplete (URL, Model, or APIKey missing)")
	}

	// Basic HTTP client
	// 基本 HTTP 客户端
	httpClient := &http.Client{
		Timeout: timeout, // Use the global LLM timeout / 使用全局 LLM 超时时间
	}

	log.L().Info("Initialized DeepSeek LLM Provider", zap.String("url", cfg.URL), zap.String("model", cfg.Model))

	return &DeepSeekProvider{
		config: cfg,
		client: httpClient,
	}, nil
}

// Name returns the name of the LLM provider.
// Name 返回 LLM 提供商的名称。
func (p *DeepSeekProvider) Name() string {
	return "deepseek" // Using the constant defined in common.types
}

// Description returns a brief description.
// Description 返回简要描述。
func (p *DeepSeekProvider) Description() string {
	return "Provides access to the DeepSeek API."
}

// GenerateText sends a chat completions request to the DeepSeek API.
// GenerateText 向 DeepSeek API 发送一个 chat completions 请求。
// DeepSeek primarily uses the chat completions endpoint.
// DeepSeek 主要使用 chat completions 终点。
// Assumes the prompt is a simple string and sends it as a user message.
// 假设 prompt 是一个简单字符串，并将其作为用户消息发送。
// Options map can be used to pass additional parameters like temperature, max_tokens.
// Options map 可用于传递其他参数，如 temperature, max_tokens。
func (p *DeepSeekProvider) GenerateText(ctx context.Context, prompt string, options map[string]interface{}) (string, error) {
	logger := log.LWithContext(ctx).With(zap.String("llmProvider", p.Name()), zap.String("model", p.config.Model))
	logger.Debug("Generating text with DeepSeek")

	if p.client == nil {
		return "", fmt.Errorf("deepseek client is not initialized")
	}

	// Construct the request body based on DeepSeek's API spec (OpenAI compatible chat completions)
	// 根据 DeepSeek 的 API 规范构建请求体 (与 OpenAI 兼容的 chat completions)
	requestBody := map[string]interface{}{
		"model": p.config.Model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":  500, // Default max tokens, can be overridden by options
		"temperature": 0.7, // Default temperature, can be overridden by options
		// Add other parameters from options
		// 添加 options 中的其他参数
	}

	// Override defaults with options
	// 使用 options 覆盖默认值
	if options != nil {
		for key, value := range options {
			// Special handling for 'messages' if needed, but generally options apply to top level
			// 如果需要特殊处理 'messages'，但通常 options 应用于顶级
			if key != "messages" {
				requestBody[key] = value
			}
		}
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		logger.Error("Failed to marshal request body", zap.Error(err))
		return "", errors.Wrap(errors.ErrorCodeLLMProviderError, "failed to marshal request body", err, "")
	}

	// DeepSeek chat completions endpoint
	// DeepSeek chat completions 终点
	reqURL := fmt.Sprintf("%s/chat/completions", p.config.URL)

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		logger.Error("Failed to create HTTP request", zap.Error(err))
		return "", errors.Wrap(errors.ErrorCodeLLMProviderError, "failed to create http request", err, "")
	}

	req.Header.Set("Content-Type", "application/json")
	// DeepSeek requires API Key in Authorization header
	// DeepSeek 需要在 Authorization header 中包含 API Key
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

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

	// Parse the response body (OpenAI chat completions format)
	// 解析响应体 (OpenAI chat completions 格式)
	var response map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		logger.Error("Failed to unmarshal response body", zap.Error(err), zap.ByteString("body", bodyBytes))
		return "", errors.Wrap(errors.ErrorCodeLLMProviderError, "failed to unmarshal response body", err, string(bodyBytes))
	}

	// Extract the generated text: choices[0].message.content
	// 提取生成的文本: choices[0].message.content
	choices, ok := response["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", errors.New(errors.ErrorCodeLLMProviderError, "no choices found in LLM response", string(bodyBytes))
	}
	firstChoice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", errors.New(errors.ErrorCodeLLMProviderError, "invalid choice format in LLM response", string(bodyBytes))
	}
	message, msgOK := firstChoice["message"].(map[string]interface{})
	if !msgOK {
		return "", errors.New(errors.ErrorCodeLLMProviderError, "message field not found in LLM response", string(bodyBytes))
	}
	generatedText, ok := message["content"].(string)
	if !ok {
		return "", errors.New(errors.ErrorCodeLLMProviderError, "content field not found in message in LLM response", string(bodyBytes))
	}

	logger.Debug("Text generated successfully")
	return generatedText, nil
}

// // Placeholder for EmbedText if needed (DeepSeek has embedding models)
// // 如果需要，EmbedText 的占位符 (DeepSeek 有 embedding 模型)
// func (p *DeepSeekProvider) EmbedText(ctx context.Context, text string) ([]float32, error) {
// 	logger := log.LWithContext(ctx).With(zap.String("llmProvider", p.Name()))
// 	logger.Warn("DeepSeekProvider.EmbedText not implemented")
// 	// TODO: Implement embedding call to DeepSeek / TODO: 实现对 DeepSeek 的 embedding 调用
// 	// Needs a separate endpoint /embeddings and request format
// 	// 需要单独的 /embeddings 终点和请求格式
// 	return nil, fmt.Errorf("embedding not implemented for DeepSeek provider")
// }

// Register the LLM provider with the global registry.
// 在全局注册表中注册 LLM 提供商。
func init() {
	// This stateful provider needs configuration and a timeout
	// before registration. We will rely on the engine initialization to create and register it.
	// 这个有状态的提供商在注册之前需要配置和超时时间。
	// 我们将依赖于引擎初始化来创建和注册它。

	// llm.RegisterLLMProvider(&DeepSeekProvider{}) // Cannot register like this
	log.L().Debug("DeepSeek LLM provider init() called, will be registered during engine setup.")
}

// Global instance placeholder, will be initialized in main.
// 全局实例占位符，将在 main 中初始化。
var DeepSeekProviderInstance *DeepSeekProvider

// RegisterDeepSeekProvider registers the initialized DeepSeekProvider instance.
// RegisterDeepSeekProvider 注册已初始化的 DeepSeekProvider 实例。
// This should be called after NewDeepSeekProvider is successful.
// 应在 NewDeepSeekProvider 成功后调用此函数。
func RegisterDeepSeekProvider(provider *DeepSeekProvider) {
	llm.RegisterLLMProvider(provider)
	DeepSeekProviderInstance = provider // Keep a global reference if needed elsewhere
}
