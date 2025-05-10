package businesssdk

import (
	"context"
)

// Package businesssdk defines the interface that business systems should implement
// to expose their operational data and capabilities to chasi-sreagent.
// 包 businesssdk 定义了业务系统应实现的接口，以向 chasi-sreagent 暴露其运营数据和能力。

// BusinessAdaptorService is the interface that a business system needs to implement
// and expose (e.g., via gRPC or HTTP) for the chasi-sreagent agent to consume.
// BusinessAdaptorService 是业务系统需要实现并暴露 (例如, 通过 gRPC 或 HTTP) 的接口，供 chasi-sreagent 代理消费。
type BusinessAdaptorService interface {
	// GetStatus retrieves the current health and key status indicators.
	// GetStatus 检索当前的健康状况和关键状态指标。
	GetStatus(ctx context.Context) (*BusinessStatus, error)

	// QueryLogs retrieves business log entries based on criteria.
	// QueryLogs 根据条件检索业务日志条目。
	// Options map can include "timeRange", "keywords", "level", "serviceId", etc.
	// options map 可以包含 "timeRange", "keywords", "level", "serviceId" 等。
	QueryLogs(ctx context.Context, options map[string]interface{}) ([]LogEntry, error)

	// GetEvents retrieves significant business events based on criteria.
	// GetEvents 根据条件检索重要的业务事件。
	// Options map can include "timeRange", "eventTypes", "severity", etc.
	// options map 可以包含 "timeRange", "eventTypes", "severity" 等。
	GetEvents(ctx context.Context, options map[string]interface{}) ([]BusinessEvent, error)

	// GetConfiguration retrieves critical configuration settings.
	// GetConfiguration 检索关键配置设置。
	GetConfiguration(ctx context.Context) (*BusinessConfig, error)

	// ListRunbooks lists the available automated or semi-automated runbooks.
	// ListRunbooks 列出可用的自动化或半自动化 Runbook。
	ListRunbooks(ctx context.Context) ([]Runbook, error)

	// ExecuteRunbook triggers the execution of a specific runbook.
	// ExecuteRunbook 触发特定 Runbook 的执行。
	// This should only be called for runbooks where the 'Automated' field is true.
	// 只应为 'Automated' 字段为 true 的 Runbook 调用此方法。
	ExecuteRunbook(ctx context.Context, runbookID string, parameters map[string]string) (*RunbookExecutionResult, error)

	// // Optional: Add other methods as needed, e.g., for specific metrics or traces
	// // 可选: 根据需要添加其他方法，例如，用于特定指标或追踪数据
	// GetMetrics(ctx context.Context, metricName string, timeRange time.Duration) (map[string]float64, error)
}
