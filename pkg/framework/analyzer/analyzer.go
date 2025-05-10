package analyzer

import (
	"context"
	"fmt"
	"sync"

	"github.com/turtacn/chasi-sreagent/pkg/common/errors"
	"github.com/turtacn/chasi-sreagent/pkg/common/log"
	"github.com/turtacn/chasi-sreagent/pkg/common/types"
)

// Package analyzer defines the interface for analysis components and a registry for managing them.
// 包 analyzer 定义了分析组件的接口和一个用于管理它们的注册表。

// Analyzer is the interface that all analysis components must implement.
// Analyzer 是所有分析组件必须实现的接口。
// An analyzer is responsible for examining specific data sources or resource types
// and identifying potential issues.
// 分析器负责检查特定的数据源或资源类型，并识别潜在问题。
type Analyzer interface {
	// Name returns the unique name of the analyzer.
	// Name 返回分析器的唯一名称。
	Name() string

	// Description returns a brief description of what the analyzer does.
	// Description 返回分析器功能的简要描述。
	Description() string

	// Analyze performs the analysis based on the provided context and collected data.
	// Analyze 根据提供的上下文和收集的数据执行分析。
	// It returns a slice of detected issues or an error.
	// 它返回一个检测到的问题切片或一个错误。
	Analyze(ctx context.Context) ([]types.Issue, error)

	// RequiredDataSources returns the types of data sources this analyzer needs.
	// RequiredDataSources 返回此分析器所需的数据源类型。
	RequiredDataSources() []enum.DataSourceType

	// // Optional: Add a method to configure the analyzer
	// // 可选: 添加一个方法来配置分析器
	// Configure(config types.AnalyzerConfig) error
}

// AnalyzerRegistry is a global registry for managing Analyzer implementations.
// AnalyzerRegistry 是一个用于管理 Analyzer 实现的全局注册表。
type AnalyzerRegistry struct {
	analyzers map[string]Analyzer
	mu        sync.RWMutex
}

// Global registry instance.
// 全局注册表实例。
var globalAnalyzerRegistry = &AnalyzerRegistry{
	analyzers: make(map[string]Analyzer),
}

// RegisterAnalyzer registers an Analyzer with the global registry.
// RegisterAnalyzer 在全局注册表中注册一个 Analyzer。
// It panics if an analyzer with the same name is already registered.
// 如果同名的分析器已被注册，则会 panic。
func RegisterAnalyzer(analyzer Analyzer) {
	globalAnalyzerRegistry.mu.Lock()
	defer globalAnalyzerRegistry.mu.Unlock()

	name := analyzer.Name()
	if _, exists := globalAnalyzerRegistry.analyzers[name]; exists {
		panic(fmt.Sprintf("analyzer with name '%s' already registered", name))
	}
	globalAnalyzerRegistry.analyzers[name] = analyzer
	log.L().Info("Registered analyzer", zap.String("name", name), zap.String("description", analyzer.Description()))
}

// GetAnalyzer retrieves an Analyzer from the global registry by name.
// GetAnalyzer 按名称从全局注册表中检索一个 Analyzer。
// It returns the Analyzer and true if found, otherwise nil and false.
// 如果找到，返回 Analyzer 和 true，否则返回 nil 和 false。
func GetAnalyzer(name string) (Analyzer, bool) {
	globalAnalyzerRegistry.mu.RLock()
	defer globalAnalyzerRegistry.mu.RUnlock()

	analyzer, found := globalAnalyzerRegistry.analyzers[name]
	return analyzer, found
}

// ListAnalyzers returns a list of all registered Analyzer names.
// ListAnalyzers 返回所有已注册的 Analyzer 名称列表。
func ListAnalyzers() []string {
	globalAnalyzerRegistry.mu.RLock()
	defer globalAnalyzerRegistry.mu.RUnlock()

	names := make([]string, 0, len(globalAnalyzerRegistry.analyzers))
	for name := range globalAnalyzerRegistry.analyzers {
		names = append(names, name)
	}
	return names
}

// GetEnabledAnalyzers retrieves enabled Analyzers from the global registry based on configuration.
// GetEnabledAnalyzers 根据配置从全局注册表中检索启用的分析器。
// If cfg.EnabledAnalyzers is empty, all registered analyzers are returned.
// 如果 cfg.EnabledAnalyzers 为空，则返回所有已注册的分析器。
func GetEnabledAnalyzers(cfg *types.AnalysisConfig) ([]Analyzer, error) {
	globalAnalyzerRegistry.mu.RLock()
	defer globalAnalyzerRegistry.mu.RUnlock()

	var enabledAnalyzers []Analyzer
	if len(cfg.EnabledAnalyzers) == 0 {
		// If no specific analyzers are enabled in config, run all registered ones
		// 如果配置中未启用特定的分析器，则运行所有已注册的分析器
		for _, analyzer := range globalAnalyzerRegistry.analyzers {
			enabledAnalyzers = append(enabledAnalyzers, analyzer)
		}
	} else {
		// Run only specified analyzers
		// 只运行指定的分析器
		for _, name := range cfg.EnabledAnalyzers {
			analyzer, found := globalAnalyzerRegistry.analyzers[name]
			if !found {
				return nil, errors.New(errors.ErrorCodeInvalidInput, "analyzer not found", fmt.Sprintf("analyzer '%s' is enabled in config but not registered", name))
			}
			enabledAnalyzers = append(enabledAnalyzers, analyzer)
		}
	}
	return enabledAnalyzers, nil
}
