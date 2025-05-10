package business

import (
	"context"
	"fmt"
	"github.com/turtacn/chasi-sreagent/pkg/adaptors/businesssdk"
	"net/url"
	"sync"

	"github.com/turtacn/chasi-sreagent/pkg/common/enum"
	"github.com/turtacn/chasi-sreagent/pkg/common/errors"
	"github.com/turtacn/chasi-sreagent/pkg/common/log"
	"github.com/turtacn/chasi-sreagent/pkg/common/types"
	"github.com/turtacn/chasi-sreagent/pkg/framework/datacollector"
	"go.uber.org/zap"
	// Add imports for gRPC or HTTP clients based on SDK implementation
	// 根据 SDK 实现添加 gRPC 或 HTTP 客户端的导入
	// "google.golang.org/grpc"
	// "net/http"
)

// Package business provides a data collector for business systems using the Business SDK.
// 包 business 提供一个使用业务 SDK 收集业务系统数据的数据采集器。

// BusinessDataCollector collects data from business systems via the Business SDK interface.
// BusinessDataCollector 通过业务 SDK 接口从业务系统收集数据。
type BusinessDataCollector struct {
	config *types.BusinessSDKConfig
	// clientCache caches connections/clients to business SDK endpoints.
	// clientCache 缓存与业务 SDK 终点的连接/客户端。
	// Map endpoint URL to client instance (e.g., gRPC client, HTTP client).
	// 将终点 URL 映射到客户端实例 (例如, gRPC 客户端, HTTP 客户端)。
	clientCache map[string]businesssdk.BusinessAdaptorService // Placeholder interface / 占位符接口
	mu          sync.RWMutex
	// Add dependencies for service discovery (e.g., K8s client if discoveryMethod is kubernetes-service)
	// 添加服务发现的依赖项 (例如, 如果 discoveryMethod 是 kubernetes-service，则添加 K8s 客户端)
	// k8sClient kubernetes.Interface // Placeholder
}

// Ensure BusinessDataCollector implements the datacollector.DataCollector interface.
// 确保 BusinessDataCollector 实现了 datacollector.DataCollector 接口。
var _ datacollector.DataCollector = &BusinessDataCollector{}

// NewBusinessDataCollector creates a new BusinessDataCollector instance.
// NewBusinessDataCollector 创建一个新的 BusinessDataCollector 实例。
// It takes the configuration and initializes endpoint discovery.
// 它接收配置并初始化终点发现。
func NewBusinessDataCollector(cfg *types.BusinessSDKConfig) (*BusinessDataCollector, error) {
	if cfg == nil {
		return nil, errors.New(errors.ErrorCodeInvalidInput, "config cannot be nil", "")
	}

	collector := &BusinessDataCollector{
		config:      cfg,
		clientCache: make(map[string]businesssdk.BusinessAdaptorService),
		// Initialize service discovery helper if needed
		// 如果需要，初始化服务发现助手
	}

	// Perform initial endpoint discovery based on config
	// 根据配置执行初始终点发现
	err := collector.discoverEndpoints(context.Background()) // Use a background context for initialization
	if err != nil {
		log.L().Error("Initial business SDK endpoint discovery failed", zap.Error(err))
		// Decide if discovery failure is fatal
		// 决定发现失败是否致命
		// For now, just log and continue, maybe some endpoints are available
		// 目前，只记录并继续，可能有些终点是可用的
	}

	// TODO: Implement periodic endpoint discovery/refresh
	// TODO: 实现周期性终点发现/刷新

	log.L().Info("Business Data Collector initialized", zap.String("discoveryMethod", cfg.DiscoveryMethod))

	return collector, nil
}

// discoverEndpoints finds and caches business SDK endpoints based on the configured method.
// discoverEndpoints 根据配置的方法查找并缓存业务 SDK 终点。
func (c *BusinessDataCollector) discoverEndpoints(ctx context.Context) error {
	logger := log.LWithContext(ctx).With(zap.String("collector", c.Name()))
	logger.Info("Discovering business SDK endpoints")

	endpoints := []types.BusinessSDKEndpoint{}
	var discoveryErr error

	switch c.config.DiscoveryMethod {
	case types.BusinessSDKDiscoveryKubernetesService:
		// TODO: Implement Kubernetes service discovery using K8s client
		// Search for services with specific labels in configured namespaces
		// TODO: 使用 K8s 客户端实现 Kubernetes 服务发现
		// 在配置的命名空间中搜索具有特定标签的服务
		logger.Warn("Kubernetes service discovery not implemented yet")
		// Example: Use k8sClient to list services
		// serviceList, err := c.k8sClient.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{LabelSelector: ...})
		// Convert Service IPs/Hostnames to endpoint URLs
		// 将 Service IP/主机名转换为终点 URL
		discoveryErr = errors.New(errors.ErrorCodeUnknown, "Kubernetes service discovery not implemented", "") // Placeholder error
	case types.BusinessSDKDiscoveryStaticList:
		endpoints = c.config.StaticEndpoints
		logger.Info("Using static list for business SDK endpoints", zap.Int("count", len(endpoints)))
	default:
		discoveryErr = errors.New(errors.ErrorCodeInvalidInput, "unsupported discovery method", fmt.Sprintf("discovery method '%s' is not supported", c.config.DiscoveryMethod))
		logger.Error("Unsupported business SDK discovery method", zap.String("method", c.config.DiscoveryMethod))
	}

	if discoveryErr != nil {
		return discoveryErr
	}

	// Cache the discovered endpoints (and potentially initialize clients)
	// 缓存发现的终点 (并可能初始化客户端)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.clientCache = make(map[string]businesssdk.BusinessAdaptorService) // Clear existing cache / 清空现有缓存
	for _, ep := range endpoints {
		// TODO: Initialize the appropriate client (gRPC or HTTP) based on URL scheme or config
		// For now, just store a placeholder
		// TODO: 根据 URL scheme 或配置初始化相应的客户端 (gRPC 或 HTTP)
		// 目前，只存储一个占位符
		logger.Debug("Discovered business SDK endpoint", zap.String("name", ep.Name), zap.String("url", ep.URL))
		// Placeholder: In a real implementation, you'd create a gRPC client or HTTP client here
		// 占位符: 在实际实现中，你将在这里创建一个 gRPC 客户端或 HTTP 客户端
		// client, err := newBusinessSDKClient(ctx, ep.URL, c.config.Timeout) // TODO: Implement client creation
		// if err != nil {
		//     logger.Error("Failed to create business SDK client", zap.String("url", ep.URL), zap.Error(err))
		//     continue // Skip this endpoint
		// }
		// c.clientCache[ep.URL] = client
		c.clientCache[ep.URL] = &placeholderBusinessAdaptorService{} // Using a placeholder struct / 使用占位符结构体
	}

	logger.Info("Business SDK endpoint discovery completed", zap.Int("discoveredCount", len(c.clientCache)))
	return nil
}

// Name returns the name of the data collector.
// Name 返回数据采集器的名称。
func (c *BusinessDataCollector) Name() string {
	return "business-data-collector"
}

// Description returns a brief description of the collector.
// Description 返回采集器的简要描述。
func (c *BusinessDataCollector) Description() string {
	return "Collects data from business systems via the Business SDK interface."
}

// Type returns the data source type.
// Type 返回数据源类型。
func (c *BusinessDataCollector) Type() enum.DataSourceType {
	return enum.DataSourceTypeBusinessSDK // This collector provides data *from* BusinessSDK
}

// Collect gathers data from business systems.
// Collect 从业务系统收集数据。
// Options should include "dataType" (e.g., enum.DataSourceTypeLog, enum.DataSourceTypeStatus)
// and potentially filters like "serviceId", "timeRange", "keywords", etc.
// Options 应包含 "dataType" (例如, enum.DataSourceTypeLog, enum.DataSourceTypeStatus)
// 并可选包含 "serviceId", "timeRange", "keywords" 等过滤器。
// Returns a slice of businesssdk types (e.g., []businesssdk.LogEntry, businesssdk.BusinessStatus) or an error.
// 返回 businesssdk 类型的切片 (例如, []businesssdk.LogEntry, businesssdk.BusinessStatus) 或一个错误。
func (c *BusinessDataCollector) Collect(ctx context.Context, options map[string]interface{}) (interface{}, error) {
	logger := log.LWithContext(ctx).With(zap.String("collector", c.Name()))
	logger.Debug("Collecting business data", zap.Any("options", options))

	dataTypeVal, ok := options["dataType"]
	if !ok {
		return nil, errors.New(errors.ErrorCodeInvalidInput, "missing 'dataType' in options", "")
	}
	dataType, ok := dataTypeVal.(enum.DataSourceType)
	if !ok {
		return nil, errors.New(errors.ErrorCodeInvalidInput, "invalid 'dataType' type in options", fmt.Sprintf("expected %T, got %T", enum.DataSourceTypeUnknown, dataTypeVal))
	}

	// TODO: Iterate through discovered endpoints and collect data based on dataType
	// This might involve calling the appropriate method on the businesssdk.BusinessAdaptorService interface
	// based on the dataType requested.
	// TODO: 遍历发现的终点并根据 dataType 收集数据
	// 这可能涉及根据请求的 dataType 调用 businesssdk.BusinessAdaptorService 接口上的适当方法。

	c.mu.RLock()
	clientsToUse := c.clientCache
	c.mu.RUnlock()

	var allCollectedData []interface{}
	var collectionErrors []error

	for url, client := range clientsToUse {
		endpointLogger := logger.With(zap.String("endpoint", url))
		endpointLogger.Debug("Collecting from endpoint")

		var collectedData interface{}
		var err error

		// --- Collection Logic based on DataType ---
		// Call the corresponding method on the client based on dataType.
		// Needs type assertion on the client if newBusinessSDKClient returns a concrete type.
		// 根据 dataType 调用客户端上相应的方法。
		// 如果 newBusinessSDKClient 返回具体类型，需要对客户端进行类型断言。

		// This requires mapping enum.DataSourceType to BusinessAdaptorService methods.
		// 这需要将 enum.DataSourceType 映射到 BusinessAdaptorService 方法。
		// Example mapping:
		switch dataType {
		case enum.DataSourceTypeStatus:
			status, statusErr := client.GetStatus(ctx)
			if statusErr != nil {
				err = fmt.Errorf("failed to get status: %w", statusErr)
			} else {
				collectedData = status
			}
		case enum.DataSourceTypeLog:
			// Need to pass relevant options to QueryLogs
			// 需要将相关 options 传递给 QueryLogs
			logs, logsErr := client.QueryLogs(ctx, options) // Assuming options map is compatible
			if logsErr != nil {
				err = fmt.Errorf("failed to query logs: %w", logsErr)
			} else {
				collectedData = logs
			}
		case enum.DataSourceTypeEvent:
			// Need to pass relevant options to GetEvents
			// 需要将相关 options 传递给 GetEvents
			events, eventsErr := client.GetEvents(ctx, options) // Assuming options map is compatible
			if eventsErr != nil {
				err = fmt.Errorf("failed to get events: %w", eventsErr)
			} else {
				collectedData = events
			}
		case enum.DataSourceTypeMetric:
			// BusinessAdaptorService interface doesn't have a generic GetMetrics yet.
			// Need to add it or specific methods.
			// BusinessAdaptorService 接口目前没有通用的 GetMetrics 方法。
			// 需要添加它或特定的方法。
			err = errors.New(errors.ErrorCodeInvalidInput, "metric collection not implemented", "")
			endpointLogger.Warn("Metric collection not implemented in Business SDK interface")

		default:
			err = errors.New(errors.ErrorCodeInvalidInput, "unsupported business data type", fmt.Sprintf("business data type '%s' is not supported by collector", dataType.String()))
			endpointLogger.Error("Unsupported business data type requested")
		}

		if err != nil {
			endpointLogger.Error("Failed to collect business data for type", zap.Error(err), zap.String("dataType", dataType.String()))
			collectionErrors = append(collectionErrors, err)
			continue // Continue collecting from other endpoints
		}

		if collectedData != nil {
			// Append data. Combining results from multiple endpoints for the same dataType
			// requires combining logic similar to K8s collector.
			// 附加数据。合并来自多个终点相同 dataType 的结果
			// 需要类似于 K8s 采集器的合并逻辑。
			// For simplicity of the stub, let's just append. Downstream analyzers
			// will need to handle processing a slice of slices or a mixed slice.
			// 为了占位符的简单性，我们暂时只附加。下游分析器
			// 需要处理处理切片组成的切片或混合切片。
			allCollectedData = append(allCollectedData, collectedData)
		}
	}

	if len(collectionErrors) > 0 && len(allCollectedData) == 0 {
		// If all endpoints failed and no data was collected, return an error.
		// 如果所有终点都失败且没有收集到数据，则返回错误。
		// Consider aggregating errors.
		// 考虑聚合错误。
		return nil, fmt.Errorf("failed to collect business data from all endpoints. First error: %w", collectionErrors[0])
	}

	logger.Debug("Business data collection finished", zap.Int("numEndpoints", len(clientsToUse)), zap.String("dataType", dataType.String()), zap.Int("itemsCollected", len(allCollectedData)))

	// Return the collected data. Structure depends on how Analyzers expect it.
	// Returning a map keyed by service name or endpoint URL might be better.
	// 返回收集到的数据。结构取决于分析器如何期望它。
	// 返回按服务名称或终点 URL 为键的映射可能更好。
	// For now, return the raw slice of collected items.
	// 目前，返回收集到的项目的原始切片。
	return allCollectedData, nil // This needs proper combining based on dataType / 这需要根据 dataType 进行适当的合并
}

// placeholderBusinessAdaptorService is a dummy implementation for the client cache.
// placeholderBusinessAdaptorService 是用于客户端缓存的虚拟实现。
// A real implementation would be a gRPC or HTTP client struct.
// 实际实现将是一个 gRPC 或 HTTP 客户端结构体。
type placeholderBusinessAdaptorService struct{}

func (p *placeholderBusinessAdaptorService) GetStatus(ctx context.Context) (*businesssdk.BusinessStatus, error) {
	log.L().Warn("placeholderBusinessAdaptorService.GetStatus called")
	return &businesssdk.BusinessStatus{Status: "Placeholder"}, nil
}
func (p *placeholderBusinessAdaptorService) QueryLogs(ctx context.Context, options map[string]interface{}) ([]businesssdk.LogEntry, error) {
	log.L().Warn("placeholderBusinessAdaptorService.QueryLogs called")
	return []businesssdk.LogEntry{}, nil
}
func (p *placeholderBusinessAdaptorService) GetEvents(ctx context.Context, options map[string]interface{}) ([]businesssdk.BusinessEvent, error) {
	log.L().Warn("placeholderBusinessAdaptorService.GetEvents called")
	return []businesssdk.BusinessEvent{}, nil
}
func (p *placeholderBusinessAdaptorService) GetConfiguration(ctx context.Context) (*businesssdk.BusinessConfig, error) {
	log.L().Warn("placeholderBusinessAdaptorService.GetConfiguration called")
	return &businesssdk.BusinessConfig{}, nil
}
func (p *placeholderBusinessAdaptorService) ListRunbooks(ctx context.Context) ([]businesssdk.Runbook, error) {
	log.L().Warn("placeholderBusinessAdaptorService.ListRunbooks called")
	return []businesssdk.Runbook{}, nil
}
func (p *placeholderBusinessAdaptorService) ExecuteRunbook(ctx context.Context, runbookID string, parameters map[string]string) (*businesssdk.RunbookExecutionResult, error) {
	log.L().Warn("placeholderBusinessAdaptorService.ExecuteRunbook called")
	return &businesssdk.RunbookExecutionResult{Status: "Placeholder"}, nil
}

// Register the data collector with the global registry.
// 在全局注册表中注册数据采集器。
func init() {
	// Similar to K8s collector, this stateful collector needs configuration
	// and possibly dependencies (like K8s client for discovery) before registration.
	// We will rely on the engine initialization to create and register the collector.
	// 类似于 K8s 采集器，这个有状态的采集器在注册之前需要配置
	// 和可能的依赖项 (例如，用于发现的 K8s 客户端)。
	// 我们将依赖于引擎初始化来创建和注册采集器。

	// datacollector.RegisterDataCollector(&BusinessDataCollector{}) // Cannot register like this as it needs config
	log.L().Debug("Business data collector init() called, will be registered during engine setup.")
}

// Global instance placeholder, will be initialized in main.
// 全局实例占位符，将在 main 中初始化。
var BusinessCollectorInstance *BusinessDataCollector

// RegisterBusinessCollector registers the initialized BusinessDataCollector instance.
// RegisterBusinessCollector 注册已初始化的 BusinessDataCollector 实例。
// This should be called after NewBusinessDataCollector is successful.
// 应在 NewBusinessDataCollector 成功后调用此函数。
func RegisterBusinessCollector(collector *BusinessDataCollector) {
	datacollector.RegisterDataCollector(collector)
	BusinessCollectorInstance = collector // Keep a global reference if needed elsewhere
}
