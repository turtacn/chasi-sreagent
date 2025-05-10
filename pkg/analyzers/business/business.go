package business

import (
	"context"
	"fmt"
	"time"

	"github.com/turtacn/chasi-sreagent/pkg/common/enum"
	"github.com/turtacn/chasi-sreagent/pkg/common/log"
	"github.com/turtacn/chasi-sreagent/pkg/common/types"
	"github.com/turtacn/chasi-sreagent/pkg/framework/analyzer"
	"github.com/turtacn/chasi-sreagent/pkg/framework/datacollector" // Will use datacollector to get business data
	"go.uber.org/zap"
)

// Package business provides analysis logic for business systems, utilizing the Business SDK.
// 包 business 提供业务系统的分析逻辑，利用业务 SDK。

// BusinessLogAnalyzer analyzes business logs for specific error patterns or events.
// BusinessLogAnalyzer 分析业务日志中特定的错误模式或事件。
type BusinessLogAnalyzer struct {
	// Add dependencies here, e.g., data collector for business logs
	// 在这里添加依赖项，例如，业务日志的数据采集器
}

// Ensure BusinessLogAnalyzer implements the analyzer.Analyzer interface.
// 确保 BusinessLogAnalyzer 实现了 analyzer.Analyzer 接口。
var _ analyzer.Analyzer = &BusinessLogAnalyzer{}

// Name returns the name of the analyzer.
// Name 返回分析器的名称。
func (a *BusinessLogAnalyzer) Name() string {
	return types.AnalyzerBusinessLog // Using the constant defined in common
}

// Description returns a brief description of the analyzer.
// Description 返回分析器的简要描述。
func (a *BusinessLogAnalyzer) Description() string {
	return "Analyzes business logs for known error patterns and events."
}

// Analyze performs analysis on business logs.
// Analyze 对业务日志执行分析。
func (a *BusinessLogAnalyzer) Analyze(ctx context.Context) ([]types.Issue, error) {
	logger := log.LWithContext(ctx).With(zap.String("analyzer", a.Name()))
	logger.Info("Running Business Log analysis")

	// TODO: Implement actual analysis logic using Business SDK data collector
	// TODO: 使用业务 SDK 数据采集器实现实际的分析逻辑

	// Example: Get Business Log data (assuming a Business data collector is registered)
	// 示例: 获取业务日志数据 (假设业务数据采集器已注册)
	businessCollector, found := datacollector.GetDataCollector("business-data-collector") // Need a constant for business collector name
	if !found {
		logger.Error("Business data collector not found")
		return nil, fmt.Errorf("business data collector not found")
	}

	// Define options for collecting logs, e.g., time range, keywords
	// 定义收集日志的选项，例如，时间范围，关键字
	collectOptions := map[string]interface{}{
		"dataType":  enum.DataSourceTypeLog, // Specify the type of business data needed
		"timeRange": 10 * time.Minute,       // Collect logs from the last 10 minutes
		"keywords":  []string{"Error", "Failed", "Exception"},
	}

	logData, err := businessCollector.Collect(ctx, collectOptions)
	if err != nil {
		logger.Error("Failed to collect business log data", zap.Error(err))
		// Decide if collection failure is fatal
		// 决定收集失败是否致命
		return nil, fmt.Errorf("failed to collect business log data: %w", err)
	}

	// The collector returns interface{}, need to cast to the expected businesssdk type
	// 采集器返回 interface{}，需要转换为期望的 businesssdk 类型
	logEntries, ok := logData.([]businesssdk.LogEntry)
	if !ok {
		logger.Error("Collected data is not []businesssdk.LogEntry", zap.Any("dataType", fmt.Sprintf("%T", logData)))
		return nil, fmt.Errorf("unexpected data type from collector: expected []businesssdk.LogEntry, got %T", logData)
	}

	// Now analyze the 'logEntries' slice
	// 现在分析 'logEntries' 切片
	issues := []types.Issue{}
	for _, entry := range logEntries {
		// Example analysis: Look for specific error messages or patterns
		// 示例分析: 查找特定的错误消息或模式
		if entry.Level == "ERROR" {
			// Create an issue for each error log entry
			// 为每个错误日志条目创建一个问题
			issues = append(issues, types.Issue{
				ID:        uuid.New().String(), // Generate unique ID
				Name:      "BusinessLogError",
				Message:   fmt.Sprintf("Business service '%s' reported error: %s", entry.ServiceID, entry.Message),
				Severity:  enum.IssueSeverityError,
				Timestamp: entry.Timestamp,
				Resource: &types.IssueResource{
					Type:      "BusinessService", // Custom resource type
					Name:      entry.ServiceID,
					Namespace: "N/A", // Namespace might not be applicable for business service
					VCluster:  "",    // Need to determine which vcluster this service is in (e.g., from collector options)
				},
				Context: map[string]interface{}{
					"logEntry": entry, // Include the log entry in context
				},
				Analyzers: []string{a.Name()},
			})
		}
		// Add more complex analysis logic here (pattern matching, anomaly detection)
		// 在这里添加更复杂的分析逻辑 (模式匹配, 异常检测)
	}

	logger.Info("Business Log analysis completed", zap.Int("issuesFound", len(issues)))
	return issues, nil
}

// RequiredDataSources returns the data source types needed by this analyzer.
// RequiredDataSources 返回此分析器所需的数据源类型。
func (a *BusinessLogAnalyzer) RequiredDataSources() []enum.DataSourceType {
	return []enum.DataSourceType{
		enum.DataSourceTypeBusinessSDK, // Needs data from Business SDK (specifically logs)
	}
}

// Register the analyzer with the global registry.
// 在全局注册表中注册分析器。
func init() {
	analyzer.RegisterAnalyzer(&BusinessLogAnalyzer{})
}
