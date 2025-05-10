package businesssdk

import (
	"time"

	"github.com/turtacn/chasi-sreagent/pkg/common/enum"
)

// Package businesssdk defines the data types used by the Business SDK interface.
// 包 businesssdk 定义了业务 SDK 接口使用的数据类型。
// These types represent the structured operational data exchanged between
// business systems and the chasi-sreagent agent.
// 这些类型表示业务系统与 chasi-sreagent 代理之间交换的结构化运营数据。

// BusinessStatus represents the current health and key status indicators of a business system or component.
// BusinessStatus 表示业务系统或组件的当前健康状况和关键状态指标。
type BusinessStatus struct {
	Timestamp time.Time              `json:"timestamp"` // Time when the status was reported / 报告状态的时间
	Status    string                 `json:"status"`    // Overall status (e.g., "Healthy", "Degraded", "Unhealthy") / 总体状态 (例如, "Healthy", "Degraded", "Unhealthy")
	Details   map[string]interface{} `json:"details"`   // Specific status details (e.g., dependency health, queue sizes) / 具体状态详情 (例如, 依赖健康状况, 队列大小)
	Metrics   map[string]float64     `json:"metrics"`   // Key business metrics (e.g., request latency, error rate, active users) / 关键业务指标 (例如, 请求延迟, 错误率, 活跃用户)
}

// LogEntry represents a single business log entry.
// LogEntry 表示单个业务日志条目。
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"` // Log timestamp / 日志时间戳
	Level     string                 `json:"level"`     // Log level (e.g., "INFO", "WARN", "ERROR") / 日志级别 (例如, "INFO", "WARN", "ERROR")
	Message   string                 `json:"message"`   // Log message content / 日志消息内容
	Fields    map[string]interface{} `json:"fields"`    // Structured log fields / 结构化日志字段
	ServiceID string                 `json:"serviceId"` // Identifier for the service or component that generated the log / 生成日志的服务或组件标识符
	TraceID   string                 `json:"traceId"`   // Optional: trace ID / 可选: trace ID
}

// BusinessEvent represents a significant event occurring within a business system.
// BusinessEvent 表示业务系统中发生的重大事件。
type BusinessEvent struct {
	Timestamp time.Time              `json:"timestamp"` // Event timestamp / 事件时间戳
	Type      string                 `json:"type"`      // Type of event (e.g., "UserLoginFailed", "OrderProcessingError", "FeatureToggleUpdated") / 事件类型 (例如, "UserLoginFailed", "OrderProcessingError", "FeatureToggleUpdated")
	Severity  enum.IssueSeverity     `json:"severity"`  // Severity of the event / 事件的严重性
	Message   string                 `json:"message"`   // Event message / 事件消息
	Details   map[string]interface{} `json:"details"`   // Event details / 事件详情
	Resource  *types.IssueResource   `json:"resource"`  // Optional: Business resource related to the event / 可选: 与事件相关的业务资源
}

// BusinessConfig represents the critical configuration settings of a business system.
// BusinessConfig 表示业务系统的关键配置设置。
// Exposing relevant configuration can help in diagnosing configuration-related issues.
// 暴露相关配置有助于诊断与配置相关的问题。
type BusinessConfig struct {
	Timestamp     time.Time              `json:"timestamp"`     // Time when the config was reported / 报告配置的时间
	ConfigVersion string                 `json:"configVersion"` // Version or identifier of the configuration / 配置的版本或标识符
	Settings      map[string]interface{} `json:"settings"`      // Key configuration settings (exclude sensitive info) / 关键配置设置 (排除敏感信息)
}

// Runbook represents an automated or semi-automated procedure within the business system.
// Runbook 表示业务系统内的自动化或半自动化过程。
// Exposing Runbooks allows the SRE agent to suggest or potentially trigger them.
// 暴露 Runbook 允许 SRE 代理建议或可能触发它们。
type Runbook struct {
	ID          string            `json:"id"`          // Unique identifier for the runbook / Runbook 的唯一标识符
	Name        string            `json:"name"`        // Name of the runbook / Runbook 名称
	Description string            `json:"description"` // Description of what the runbook does / Runbook 功能描述
	Parameters  map[string]string `json:"parameters"`  // Expected parameters for execution / 执行所需的参数
	Automated   bool              `json:"automated"`   // Can this runbook be automated by the agent? / 此 Runbook 是否可由代理自动化执行?
	Steps       []string          `json:"steps"`       // (Optional) Manual steps if not automated / (可选) 如果非自动化，则为手动步骤
}

// RunbookExecutionResult represents the result of executing a runbook via the SDK.
// RunbookExecutionResult 表示通过 SDK 执行 Runbook 的结果。
type RunbookExecutionResult struct {
	ExecutionID string    `json:"executionId"` // Unique ID for this execution instance / 此执行实例的唯一 ID
	Timestamp   time.Time `json:"timestamp"`   // Time of execution / 执行时间
	Status      string    `json:"status"`      // Status of execution (e.g., "Succeeded", "Failed", "Running") / 执行状态 (例如, "Succeeded", "Failed", "Running")
	Output      string    `json:"output"`      // Output or result message / 输出或结果消息
	Error       string    `json:"error"`       // Error message if failed / 如果失败的错误信息
}
