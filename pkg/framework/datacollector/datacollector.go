package datacollector

import (
	"context"
	"fmt"
	"github.com/uber-go/zap"
	"sync"

	"github.com/turtacn/chasi-sreagent/pkg/common/enum"
	"github.com/turtacn/chasi-sreagent/pkg/common/log"
	"github.com/turtacn/chasi-sreagent/pkg/common/types"
)

// Package datacollector defines the interface for data collection components and a registry.
// 包 datacollector 定义了数据采集组件的接口和一个注册表。

// DataCollector is the interface that all data collection components must implement.
// DataCollector 是所有数据采集组件必须实现的接口。
// A data collector is responsible for gathering raw data from specific sources.
// 数据采集器负责从特定来源收集原始数据。
type DataCollector interface {
	// Name returns the unique name of the data collector.
	// Name 返回数据采集器的唯一名称。
	Name() string

	// Description returns a brief description of what the collector does.
	// Description 返回采集器功能的简要描述。
	Description() string

	// Type returns the type of data source this collector handles.
	// Type 返回此采集器处理的数据源类型。
	Type() enum.DataSourceType

	// Collect gathers data from the source based on the provided context and options.
	// Collect 根据提供的上下文和选项从来源收集数据。
	// The options map can contain parameters like resource names, namespaces, time ranges, etc.
	// options 映射可以包含资源名称、命名空间、时间范围等参数。
	// It returns the collected raw data or an error. The raw data structure is collector-specific.
	// 它返回收集到的原始数据或一个错误。原始数据结构是采集器特定的。
	// Implementations should use specific types for the collected data where possible,
	// or a generic map/slice for unstructured data.
	// 实现应尽可能使用特定类型表示收集到的数据，或使用通用映射/切片表示非结构化数据。
	Collect(ctx context.Context, options map[string]interface{}) (interface{}, error)

	// // Optional: Add a method to configure the collector
	// // 可选: 添加一个方法来配置采集器
	// Configure(config types.DataCollectorConfig) error
}

// DataCollectorRegistry is a global registry for managing DataCollector implementations.
// DataCollectorRegistry 是一个用于管理 DataCollector 实现的全局注册表。
type DataCollectorRegistry struct {
	collectors map[string]DataCollector
	mu         sync.RWMutex
}

// Global registry instance.
// 全局注册表实例。
var globalDataCollectorRegistry = &DataCollectorRegistry{
	collectors: make(map[string]DataCollector),
}

// RegisterDataCollector registers a DataCollector with the global registry.
// RegisterDataCollector 在全局注册表中注册一个 DataCollector。
// It panics if a collector with the same name is already registered.
// 如果同名的采集器已被注册，则会 panic。
func RegisterDataCollector(collector DataCollector) {
	globalDataCollectorRegistry.mu.Lock()
	defer globalDataCollectorRegistry.mu.Unlock()

	name := collector.Name()
	if _, exists := globalDataCollectorRegistry.collectors[name]; exists {
		panic(fmt.Sprintf("data collector with name '%s' already registered", name))
	}
	globalDataCollectorRegistry.collectors[name] = collector
	log.L().Info("Registered data collector", zap.String("name", name), zap.String("type", collector.Type().String()))
}

// GetDataCollector retrieves a DataCollector from the global registry by name.
// GetDataCollector 按名称从全局注册表中检索一个 DataCollector。
// It returns the DataCollector and true if found, otherwise nil and false.
// 如果找到，返回 DataCollector 和 true，否则返回 nil 和 false。
func GetDataCollector(name string) (DataCollector, bool) {
	globalDataCollectorRegistry.mu.RLock()
	defer globalDataCollectorRegistry.mu.RUnlock()

	collector, found := globalDataCollectorRegistry.collectors[name]
	return collector, found
}

// GetDataCollectorsByType retrieves DataCollectors from the global registry by source type.
// GetDataCollectorsByType 按来源类型从全局注册表中检索 DataCollector。
func GetDataCollectorsByType(dataType enum.DataSourceType) []DataCollector {
	globalDataCollectorRegistry.mu.RLock()
	defer globalDataCollectorRegistry.mu.RUnlock()

	var collectors []DataCollector
	for _, collector := range globalDataCollectorRegistry.collectors {
		if collector.Type() == dataType {
			collectors = append(collectors, collector)
		}
	}
	return collectors
}
