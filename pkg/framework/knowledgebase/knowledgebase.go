package knowledgebase

import (
	"context"
	"fmt"
	"github.com/uber-go/zap"
	"sync"

	"github.com/turtacn/chasi-sreagent/pkg/common/errors"
	"github.com/turtacn/chasi-sreagent/pkg/common/log"
	"github.com/turtacn/chasi-sreagent/pkg/common/types"
)

// Package knowledgebase defines the interface for knowledge base components and RAG logic.
// 包 knowledgebase 定义了知识库组件和 RAG 逻辑的接口。

// KnowledgeBase is the interface that all knowledge base providers must implement.
// KnowledgeBase 是所有知识库提供商必须实现的接口。
// It is used to store and retrieve SRE-related knowledge for RAG.
// 它用于存储和检索 SRE 相关知识以支持 RAG。
type KnowledgeBase interface {
	// Name returns the unique name of the knowledge base provider.
	// Name 返回知识库提供商的唯一名称。
	Name() string

	// Description returns a brief description of the knowledge base.
	// Description 返回知识库的简要描述。
	Description() string

	// Store adds knowledge data to the knowledge base.
	// Store 向知识库添加知识数据。
	// The data can be structured (e.g., document snippets, Q&A) or unstructured.
	// 数据可以是结构化的 (例如, 文档片段, 问答) 或非结构化的。
	Store(ctx context.Context, data []types.KnowledgeBaseHit) error // Reusing KnowledgeBaseHit for data structure? Or define new type?
	// Note: Rethink the data structure for storing - maybe a dedicated struct for source data.
	// 注意: 需要重新思考存储的数据结构 - 可能需要一个专门的结构体来表示源数据。
	// For now, using KnowledgeBaseHit for simplicity, but will refine later.
	// 目前为简化起见使用 KnowledgeBaseHit，后续会进行改进。

	// Retrieve searches the knowledge base for information relevant to the query.
	// Retrieve 在知识库中搜索与查询相关的信息。
	// It returns a slice of relevant knowledge snippets (hits) or an error.
	// 它返回一个相关的知识片段 (命中) 切片或一个错误。
	Retrieve(ctx context.Context, query string, options map[string]interface{}) ([]types.KnowledgeBaseHit, error)

	// // Optional: Add a method to configure the knowledge base
	// // 可选: 添加一个方法来配置知识库
	// Configure(config types.KnowledgeBaseConfig) error
}

// KnowledgeBaseRegistry is a global registry for managing KnowledgeBase implementations.
// KnowledgeBaseRegistry 是一个用于管理 KnowledgeBase 实现的全局注册表。
type KnowledgeBaseRegistry struct {
	kbs map[string]KnowledgeBase
	mu  sync.RWMutex
}

// Global registry instance.
// 全局注册表实例。
var globalKnowledgeBaseRegistry = &KnowledgeBaseRegistry{
	kbs: make(map[string]KnowledgeBase),
}

// RegisterKnowledgeBase registers a KnowledgeBase with the global registry.
// RegisterKnowledgeBase 在全局注册表中注册一个 KnowledgeBase。
// It panics if a provider with the same name is already registered.
// 如果同名的提供商已被注册，则会 panic。
func RegisterKnowledgeBase(kb KnowledgeBase) {
	globalKnowledgeBaseRegistry.mu.Lock()
	defer globalKnowledgeBaseRegistry.mu.Unlock()

	name := kb.Name()
	if _, exists := globalKnowledgeBaseRegistry.kbs[name]; exists {
		panic(fmt.Sprintf("knowledge base provider with name '%s' already registered", name))
	}
	globalKnowledgeBaseRegistry.kbs[name] = kb
	log.L().Info("Registered knowledge base provider", zap.String("name", name), zap.String("description", kb.Description()))
}

// GetKnowledgeBase retrieves a KnowledgeBase from the global registry by name.
// GetKnowledgeBase 按名称从全局注册表中检索一个 KnowledgeBase。
// It returns the KnowledgeBase and true if found, otherwise nil and false.
// 如果找到，返回 KnowledgeBase 和 true，否则返回 nil 和 false。
func GetKnowledgeBase(name string) (KnowledgeBase, bool) {
	globalKnowledgeBaseRegistry.mu.RLock()
	defer globalKnowledgeBaseRegistry.mu.RUnlock()

	kb, found := globalKnowledgeBaseRegistry.kbs[name]
	return kb, found
}

// GetEnabledKnowledgeBase retrieves the enabled KnowledgeBase provider based on configuration.
// GetEnabledKnowledgeBase 根据配置检索启用的知识库提供商。
func GetEnabledKnowledgeBase(cfg *types.KnowledgeBaseConfig) (KnowledgeBase, error) {
	if !cfg.Enabled {
		return nil, errors.New(errors.ErrorCodeInvalidInput, "knowledge base is disabled", "")
	}
	kb, found := GetKnowledgeBase(cfg.Provider)
	if !found {
		return nil, errors.New(errors.ErrorCodeNotFound, "knowledge base provider not found", fmt.Sprintf("knowledge base provider '%s' is enabled in config but not registered", cfg.Provider))
	}
	return kb, nil
}
