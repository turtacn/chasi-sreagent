package enum

// Package enum defines various enumeration types used across the chasi-sreagent project.
// 包 enum 定义了 chasi-sreagent 项目中使用的各种枚举类型。

// LogLevel represents the severity level of a log entry.
// LogLevel 表示日志条目的严重级别。
type LogLevel int

const (
	// LogLevelDebug indicates debug level logs.
	// LogLevelDebug 表示调试级别日志。
	LogLevelDebug LogLevel = iota
	// LogLevelInfo indicates informational logs.
	// LogLevelInfo 表示信息级别日志。
	LogLevelInfo
	// LogLevelWarn indicates warning level logs.
	// LogLevelWarn 表示警告级别日志。
	LogLevelWarn
	// LogLevelError indicates error level logs.
	// LogLevelError 表示错误级别日志。
	LogLevelError
	// LogLevelFatal indicates fatal error logs, usually followed by program termination.
	// LogLevelFatal 表示致命错误日志，通常随后程序终止。
	LogLevelFatal
	// LogLevelPanic indicates panic level logs, usually followed by a panic.
	// LogLevelPanic 表示恐慌级别日志，通常随后发生 panic。
	LogLevelPanic
)

// String returns the string representation of a LogLevel.
// String 返回 LogLevel 的字符串表示。
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "debug"
	case LogLevelInfo:
		return "info"
	case LogLevelWarn:
		return "warn"
	case LogLevelError:
		return "error"
	case LogLevelFatal:
		return "fatal"
	case LogLevelPanic:
		return "panic"
	default:
		return "unknown"
	}
}

// IssueSeverity represents the severity of a detected issue.
// IssueSeverity 表示检测到问题的严重性。
type IssueSeverity int

const (
	// IssueSeverityUnknown indicates unknown severity.
	// IssueSeverityUnknown 表示未知严重性。
	IssueSeverityUnknown IssueSeverity = iota
	// IssueSeverityInfo indicates informational severity.
	// IssueSeverityInfo 表示信息级别严重性。
	IssueSeverityInfo
	// IssueSeverityWarning indicates warning severity.
	// IssueSeverityWarning 表示警告级别严重性。
	IssueSeverityWarning
	// IssueSeverityError indicates error severity.
	// IssueSeverityError 表示错误级别严重性。
	IssueSeverityError
	// IssueSeverityCritical indicates critical severity.
	// IssueSeverityCritical 表示紧急级别严重性。
	IssueSeverityCritical
)

// String returns the string representation of an IssueSeverity.
// String 返回 IssueSeverity 的字符串表示。
func (s IssueSeverity) String() string {
	switch s {
	case IssueSeverityInfo:
		return "Info"
	case IssueSeverityWarning:
		return "Warning"
	case IssueSeverityError:
		return "Error"
	case IssueSeverityCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

// AnalysisStatus represents the status of an analysis task.
// AnalysisStatus 表示分析任务的状态。
type AnalysisStatus int

const (
	// AnalysisStatusPending indicates the analysis is pending.
	// AnalysisStatusPending 表示分析正在等待中。
	AnalysisStatusPending AnalysisStatus = iota
	// AnalysisStatusRunning indicates the analysis is currently running.
	// AnalysisStatusRunning 表示分析正在进行中。
	AnalysisStatusRunning
	// AnalysisStatusCompleted indicates the analysis completed successfully.
	// AnalysisStatusCompleted 表示分析成功完成。
	AnalysisStatusCompleted
	// AnalysisStatusFailed indicates the analysis failed.
	// AnalysisStatusFailed 表示分析失败。
	AnalysisStatusFailed
	// AnalysisStatusSkipped indicates the analysis was skipped.
	// AnalysisStatusSkipped 表示分析被跳过。
	AnalysisStatusSkipped
)

// String returns the string representation of an AnalysisStatus.
// String 返回 AnalysisStatus 的字符串表示。
func (s AnalysisStatus) String() string {
	switch s {
	case AnalysisStatusPending:
		return "Pending"
	case AnalysisStatusRunning:
		return "Running"
	case AnalysisStatusCompleted:
		return "Completed"
	case AnalysisStatusFailed:
		return "Failed"
	case AnalysisStatusSkipped:
		return "Skipped"
	default:
		return "Unknown"
	}
}

// ActionType represents the type of a remediation action.
// ActionType 表示处置动作的类型。
type ActionType int

const (
	// ActionTypeUnknown indicates an unknown action type.
	// ActionTypeUnknown 表示未知动作类型。
	ActionTypeUnknown ActionType = iota
	// ActionTypeSuggestion indicates the action is a suggestion for manual execution.
	// ActionTypeSuggestion 表示该动作是供人工执行的建议。
	ActionTypeSuggestion
	// ActionTypeAutomated indicates the action can be automated and executed by the agent.
	// ActionTypeAutomated 表示该动作可以由代理自动化执行。
	ActionTypeAutomated
)

// String returns the string representation of an ActionType.
// String 返回 ActionType 的字符串表示。
func (t ActionType) String() string {
	switch t {
	case ActionTypeSuggestion:
		return "Suggestion"
	case ActionTypeAutomated:
		return "Automated"
	default:
		return "Unknown"
	}
}

// DataSourceType represents the type of a data source.
// DataSourceType 表示数据源的类型。
type DataSourceType int

const (
	// DataSourceTypeUnknown indicates an unknown data source type.
	// DataSourceTypeUnknown 表示未知数据源类型。
	DataSourceTypeUnknown DataSourceType = iota
	// DataSourceTypeKubernetesAPI indicates data from Kubernetes API.
	// DataSourceTypeKubernetesAPI 表示来自 Kubernetes API 的数据。
	DataSourceTypeKubernetesAPI
	// DataSourceTypeBusinessSDK indicates data from a Business SDK.
	// DataSourceTypeBusinessSDK 表示来自业务 SDK 的数据。
	DataSourceTypeBusinessSDK
	// DataSourceTypeLog indicates data from logs.
	// DataSourceTypeLog 表示来自日志的数据。
	DataSourceTypeLog
	// DataSourceTypeMetric indicates data from metrics.
	// DataSourceTypeMetric 表示来自指标的数据。
	DataSourceTypeMetric
	// DataSourceTypeEvent indicates data from events.
	// DataSourceTypeEvent 表示来自事件的数据。
	DataSourceTypeEvent
)

// String returns the string representation of a DataSourceType.
// String 返回 DataSourceType 的字符串表示。
func (t DataSourceType) String() string {
	switch t {
	case DataSourceTypeKubernetesAPI:
		return "KubernetesAPI"
	case DataSourceTypeBusinessSDK:
		return "BusinessSDK"
	case DataSourceTypeLog:
		return "Log"
	case DataSourceTypeMetric:
		return "Metric"
	case DataSourceTypeEvent:
		return "Event"
	default:
		return "Unknown"
	}
}
