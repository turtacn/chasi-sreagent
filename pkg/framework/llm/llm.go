package llm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/turtacn/chasi-sreagent/pkg/common/errors"
	"github.com/turtacn/chasi-sreagent/pkg/common/log"
	"github.com/turtacn/chasi-sreagent/pkg/common/types"
)

// Package llm defines the interface for LLM providers and a registry.
// 包 llm 定义了 LLM 提供商的接口和一个注册表。

// LLM is the interface that all LLM providers must implement.
// LLM 是所有 LLM 提供商必须实现的接口。
// It provides methods for interacting with language models.
// 它提供了与语言模型交互的方法。
type LLM interface {
	// Name returns the unique name of the LLM provider.
	// Name 返回 LLM 提供商的唯一名称。
	Name() string

	// Description returns a brief description of the LLM provider.
	// Description 返回 LLM 提供商的简要描述。
	Description() string

	// GenerateText sends a prompt to the LLM and returns the generated text.
	// GenerateText 将提示发送给 LLM 并返回生成的文本。
	// The prompt can be a simple string or a more structured representation.
	// 提示可以是简单字符串或更结构化的表示。
	// Options can include parameters like temperature, max tokens, etc.
	// Options 可以包含温度、最大 token 等参数。
	GenerateText(ctx context.Context, prompt string, options map[string]interface{}) (string, error)

	// // Optional: Add a method for generating embeddings
	// // 可选: 添加一个用于生成 embedding 的方法
	// EmbedText(ctx context.Context, text string) ([]float32, error)

	// // Optional: Add a method to configure the LLM provider
	// // 可选: 添加一个方法来配置 LLM 提供商
	// Configure(config types.LLMProviderConfig) error
}

// LLMRegistry is a global registry for managing LLM implementations.
// LLMRegistry 是一个用于管理 LLM 实现的全局注册表。
type LLMRegistry struct {
	providers map[string]LLM
	mu        sync.RWMutex
}

// Global registry instance.
// 全局注册表实例。
var globalLLMRegistry = &LLMRegistry{
	providers: make(map[string]LLM),
}

// RegisterLLMProvider registers an LLM provider with the global registry.
// RegisterLLMProvider 在全局注册表中注册一个 LLM 提供商。
// It panics if a provider with the same name is already registered.
// 如果同名的提供商已被注册，则会 panic。
func RegisterLLMProvider(provider LLM) {
	globalLLMRegistry.mu.Lock()
	defer globalLLMRegistry.mu.Unlock()

	name := provider.Name()
	if _, exists := globalLLMRegistry.providers[name]; exists {
		panic(fmt.Sprintf("LLM provider with name '%s' already registered", name))
	}
	globalLLMRegistry.providers[name] = provider
	log.L().Info("Registered LLM provider", zap.String("name", name), zap.String("description", provider.Description()))
}

// GetLLMProvider retrieves an LLM provider from the global registry by name.
// GetLLMProvider 按名称从全局注册表中检索一个 LLM 提供商。
// It returns the LLM provider and true if found, otherwise nil and false.
// 如果找到，返回 LLM 提供商和 true，否则返回 nil 和 false。
func GetLLMProvider(name string) (LLM, bool) {
	globalLLMRegistry.mu.RLock()
	defer globalLLMRegistry.mu.RUnlock()

	provider, found := globalLLMRegistry.providers[name]
	return provider, found
}

// GetEnabledLLMProvider retrieves the enabled LLM provider based on configuration.
// GetEnabledLLMProvider 根据配置检索启用的 LLM 提供商。
func GetEnabledLLMProvider(cfg *types.LLMConfig) (LLM, error) {
	provider, found := GetLLMProvider(cfg.Provider)
	if !found {
		return nil, errors.New(errors.ErrorCodeNotFound, "LLM provider not found", fmt.Sprintf("LLM provider '%s' is enabled in config but not registered", cfg.Provider))
	}
	return provider, nil
}

// WithTimeout adds a timeout to the context for LLM calls.
// WithTimeout 为 LLM 调用向 context 添加超时。
func WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout > 0 {
		return context.WithTimeout(ctx, timeout)
	}
	return context.WithCancel(ctx) // No timeout, but still provide a cancel func
}
