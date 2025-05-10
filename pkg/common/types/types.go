package types

import (
	"github.com/turtacn/chasi-sreagent/pkg/common/types/enum"
	"time"
)

// Package types defines common data structures used across the chasi-sreagent project.
// 包 types 定义了 chasi-sreagent 项目中使用的通用数据结构。

// ContextKey is a type for context keys to avoid collisions.
// ContextKey 是用于 context 键的类型，以避免冲突。
type ContextKey string

const (
	// ContextKeyLogger is the key for the logger in the context.
	// ContextKeyLogger 是 context 中 logger 的键。
	ContextKeyLogger ContextKey = "logger"
	// ContextKeyConfig is the key for the configuration in the context.
	// ContextKeyConfig 是 context 中配置的键。
	ContextKeyConfig ContextKey = "config"
	// ContextKeyVClusterName is the key for the current vcluster name in the context.
	// ContextKeyVClusterName 是 context 中当前 vcluster 名称的键。
	ContextKeyVClusterName ContextKey = "vclusterName"
	// ContextKeyTraceID is the key for the trace ID in the context.
	// ContextKeyTraceID 是 context 中 trace ID 的键。
	ContextKeyTraceID ContextKey = "traceID"
)

// Config represents the overall configuration structure.
// Config 表示整体配置结构。
type Config struct {
	Log           LogConfig           `yaml:"log"`           // Logging configuration / 日志配置
	Kubernetes    KubernetesConfig    `yaml:"kubernetes"`    // Kubernetes connection configuration / Kubernetes 连接配置
	LLM           LLMConfig           `yaml:"llm"`           // LLM configuration / 大模型配置
	KnowledgeBase KnowledgeBaseConfig `yaml:"knowledgeBase"` // Knowledge base (RAG) configuration / 知识库 (RAG) 配置
	BusinessSDK   BusinessSDKConfig   `yaml:"businessSDK"`   // Business Adaptation SDK configuration / 业务适配 SDK 配置
	Analysis      AnalysisConfig      `yaml:"analysis"`      // Analysis configuration / 分析配置
	Actions       ActionsConfig       `yaml:"actions"`       // Action configuration / 动作配置
}

// LogConfig represents logging configuration.
// LogConfig 表示日志配置。
type LogConfig struct {
	Level  string `yaml:"level"`  // Logging level / 日志级别 (e.g., "info", "debug")
	Format string `yaml:"format"` // Logging format / 日志格式 (e.g., "json", "console")
	Output string `yaml:"output"` // Logging output / 日志输出 (e.g., "stdout", "/var/log/agent.log")
}

// KubernetesConfig represents Kubernetes connection configuration.
// KubernetesConfig 表示 Kubernetes 连接配置。
type KubernetesConfig struct {
	KubeconfigPath string           `yaml:"kubeconfigPath"` // Path to kubeconfig file / kubeconfig 文件路径
	Vclusters      []VClusterConfig `yaml:"vclusters"`      // List of vcluster configurations / vcluster 配置列表
}

// VClusterConfig represents configuration for a single vcluster.
// VClusterConfig 表示单个 vcluster 的配置。
type VClusterConfig struct {
	Name       string `yaml:"name"`       // Name of the vcluster / vcluster 名称
	Context    string `yaml:"context"`    // Kubeconfig context name (alternative to inline kubeconfig) / Kubeconfig 上下文名称 (与内联 kubeconfig 互斥)
	Kubeconfig string `yaml:"kubeconfig"` // Inline kubeconfig content / 内联 kubeconfig 内容
	Namespace  string `yaml:"namespace"`  // Namespace in the host cluster where the vcluster runs / vcluster 在宿主机集群中运行的命名空间
}

// LLMConfig represents LLM configuration.
// LLMConfig 表示大模型配置。
type LLMConfig struct {
	Provider string            `yaml:"provider"` // LLM provider name / LLM 提供商名称
	LocalAI  LLMProviderConfig `yaml:"localai"`  // LocalAI specific config / LocalAI 特定配置
	DeepSeek LLMProviderConfig `yaml:"deepseek"` // DeepSeek specific config / DeepSeek 特定配置
	OpenAI   LLMProviderConfig `yaml:"openai"`   // OpenAI specific config / OpenAI 特定配置
	Timeout  time.Duration     `yaml:"timeout"`  // Timeout for LLM API calls / LLM API 调用超时时间
}

// LLMProviderConfig represents configuration for a specific LLM provider.
// LLMProviderConfig 表示特定 LLM 提供商的配置。
type LLMProviderConfig struct {
	URL    string `yaml:"url"`    // API endpoint URL / API 端点 URL
	Model  string `yaml:"model"`  // Model name / 模型名称
	APIKey string `yaml:"apiKey"` // API Key / API Key
	// Add other provider specific fields here
	// 在这里添加其他提供商特定字段
}

// KnowledgeBaseConfig represents knowledge base (RAG) configuration.
// KnowledgeBaseConfig 表示知识库 (RAG) 配置。
type KnowledgeBaseConfig struct {
	Enabled   bool            `yaml:"enabled"`   // Enable RAG / 启用 RAG
	Provider  string          `yaml:"provider"`  // KB provider name / 知识库提供商名称
	VectorDB  VectorDBConfig  `yaml:"vectorDB"`  // Vector DB specific config / 向量数据库特定配置
	Embedding EmbeddingConfig `yaml:"embedding"` // Embedding model config for RAG / RAG 使用的 Embedding 模型配置
}

// VectorDBConfig represents configuration for a vector database.
// VectorDBConfig 表示向量数据库配置。
type VectorDBConfig struct {
	URL        string `yaml:"url"`        // Vector database endpoint URL / 向量数据库终点 URL
	Collection string `yaml:"collection"` // Collection/index name / Collection/索引 名称
	APIKey     string `yaml:"apiKey"`     // API Key / API Key
	// Add other vector DB specific fields
	// 添加其他向量数据库特定字段
}

// EmbeddingConfig represents configuration for an embedding model.
// EmbeddingConfig 表示 Embedding 模型配置。
type EmbeddingConfig struct {
	Provider string `yaml:"provider"` // Embedding model provider name / Embedding 模型提供商名称
	Model    string `yaml:"model"`    // Model name / 模型名称
	URL      string `yaml:"url"`      // API endpoint URL / API 端点 URL
	APIKey   string `yaml:"apiKey"`   // API Key / API Key
}

// BusinessSDKConfig represents business SDK adaptation configuration.
// BusinessSDKConfig 表示业务 SDK 适配配置。
type BusinessSDKConfig struct {
	DiscoveryMethod            string                           `yaml:"discoveryMethod"`            // Method to discover business service endpoints / 发现业务服务终点的方法
	KubernetesServiceDiscovery KubernetesServiceDiscoveryConfig `yaml:"kubernetesServiceDiscovery"` // K8s service discovery config / K8s 服务发现配置
	StaticEndpoints            []BusinessSDKEndpoint            `yaml:"staticEndpoints"`            // Static list of endpoints / 终点静态列表
	Timeout                    time.Duration                    `yaml:"timeout"`                    // Timeout for calling SDK endpoints / 调用 SDK 终点的超时时间
}

// KubernetesServiceDiscoveryConfig represents configuration for discovering business services via Kubernetes.
// KubernetesServiceDiscoveryConfig 表示通过 Kubernetes 发现业务服务的配置。
type KubernetesServiceDiscoveryConfig struct {
	Namespaces []string          `yaml:"namespaces"` // Namespaces to scan / 扫描的命名空间
	Selector   map[string]string `yaml:"selector"`   // Label selector / 标签选择器
}

// BusinessSDKEndpoint represents a single business SDK endpoint.
// BusinessSDKEndpoint 表示单个业务 SDK 终点。
type BusinessSDKEndpoint struct {
	Name string `yaml:"name"` // Name of the business service / 业务服务名称
	URL  string `yaml:"url"`  // Endpoint URL (e.g., "grpc://...", "http://...") / 终点 URL (例如, "grpc://...", "http://...")
}

// AnalysisConfig represents analysis configuration.
// AnalysisConfig 表示分析配置。
type AnalysisConfig struct {
	EnabledAnalyzers []string      `yaml:"enabledAnalyzers"` // List of analyzers to enable / 要启用的分析器列表
	Interval         time.Duration `yaml:"interval"`         // Default analysis interval / 默认分析间隔
	// Add other analysis specific configurations
	// 添加其他分析特定配置
}

// ActionsConfig represents actions configuration.
// ActionsConfig 表示动作配置。
type ActionsConfig struct {
	Enabled bool `yaml:"enabled"` // Enable automated actions / 启用自动化动作
	// Add action specific configurations (e.g., approval workflows, dry run)
	// 添加动作特定配置 (例如, 审批流程, 干运行)
}

// Issue represents a detected issue in the environment.
// Issue 表示环境中检测到的问题。
type Issue struct {
	ID        string                 `json:"id"`        // Unique identifier for the issue / 问题的唯一标识符
	Name      string                 `json:"name"`      // Name or type of the issue / 问题名称或类型
	Message   string                 `json:"message"`   // Detailed message describing the issue / 描述问题的详细信息
	Severity  enum.IssueSeverity     `json:"severity"`  // Severity of the issue / 问题的严重性
	Timestamp time.Time              `json:"timestamp"` // Time when the issue was detected / 检测到问题的时间
	Resource  *IssueResource         `json:"resource"`  // Resource associated with the issue / 与问题相关的资源
	Context   map[string]interface{} `json:"context"`   // Additional context data / 附加上下文数据
	Analyzers []string               `json:"analyzers"` // Analyzers that identified this issue / 识别出此问题的分析器
}

// IssueResource represents a resource associated with an issue (e.g., K8s object, business service).
// IssueResource 表示与问题相关的资源 (例如, K8s 对象, 业务服务)。
type IssueResource struct {
	Type      string `json:"type"`      // Type of the resource (e.g., "Pod", "Deployment", "BusinessService") / 资源类型 (例如, "Pod", "Deployment", "BusinessService")
	Namespace string `json:"namespace"` // Namespace of the resource (if applicable) / 资源的命名空间 (如果适用)
	Name      string `json:"name"`      // Name of the resource / 资源的名称
	UID       string `json:"uid"`       // UID of the resource (if applicable) / 资源的 UID (如果适用)
	VCluster  string `json:"vcluster"`  // Name of the vcluster the resource belongs to (if applicable) / 资源所属的 vcluster 名称 (如果适用)
}

// AnalysisResult represents the result of an analysis run.
// AnalysisResult 表示一次分析运行的结果。
type AnalysisResult struct {
	ID           string              `json:"id"`           // Unique identifier for the analysis run / 分析运行的唯一标识符
	Timestamp    time.Time           `json:"timestamp"`    // Time when the analysis was performed / 执行分析的时间
	Duration     time.Duration       `json:"duration"`     // Duration of the analysis run / 分析运行的持续时间
	Status       enum.AnalysisStatus `json:"status"`       // Status of the analysis run / 分析运行的状态
	Issues       []Issue             `json:"issues"`       // List of issues found / 找到的问题列表
	AnalyzersRun []string            `json:"analyzersRun"` // List of analyzers that were run / 运行的分析器列表
	Error        string              `json:"error"`        // Error message if status is failed / 如果状态为失败的错误信息
}

// RemediationSuggestion represents a suggested action to resolve an issue.
// RemediationSuggestion 表示解决问题的建议动作。
type RemediationSuggestion struct {
	IssueID     string                 `json:"issueId"`     // ID of the issue this suggestion is for / 此建议针对的问题 ID
	Description string                 `json:"description"` // Description of the suggested action / 建议动作的描述
	ActionType  enum.ActionType        `json:"actionType"`  // Type of action (Suggestion or Automated) / 动作类型 (建议 或 自动化)
	Command     string                 `json:"command"`     // (Optional) Command to execute for suggestion type / (可选) 建议类型的执行命令
	Payload     map[string]interface{} `json:"payload"`     // (Optional) Payload for automated action type / (可选) 自动化动作类型的载荷
	Confidence  float64                `json:"confidence"`  // Confidence level of the suggestion (0.0 - 1.0) / 建议的置信度 (0.0 - 1.0)
	Source      string                 `json:"source"`      // Source of the suggestion (e.g., "LLM", "KnowledgeBase", "Rule") / 建议的来源 (例如, "LLM", "KnowledgeBase", "Rule")
}

// DiagnosisResult represents the outcome of the diagnosis process.
// DiagnosisResult 表示诊断过程的结果。
type DiagnosisResult struct {
	AnalysisResultID  string                  `json:"analysisResultId"`  // ID of the analysis result this diagnosis is based on / 此诊断所基于的分析结果 ID
	Timestamp         time.Time               `json:"timestamp"`         // Time when the diagnosis was performed / 执行诊断的时间
	Duration          time.Duration           `json:"duration"`          // Duration of the diagnosis process / 诊断过程的持续时间
	RootCause         string                  `json:"rootCause"`         // Identified root cause in natural language / 识别出的自然语言根因
	Suggestions       []RemediationSuggestion `json:"suggestions"`       // List of suggested remediation actions / 建议的处置动作列表
	LLMInteraction    *LLMInteractionDetails  `json:"llmInteraction"`    // Details about the LLM interaction / 大模型交互详情
	KnowledgeBaseHits []KnowledgeBaseHit      `json:"knowledgeBaseHits"` // Details about knowledge base hits / 知识库命中详情
	Error             string                  `json:"error"`             // Error message if diagnosis failed / 如果诊断失败的错误信息
}

// LLMInteractionDetails represents details about the interaction with the LLM.
// LLMInteractionDetails 表示与大模型交互的详情。
type LLMInteractionDetails struct {
	Provider string `json:"provider"` // LLM provider used / 使用的 LLM 提供商
	Model    string `json:"model"`    // Model used / 使用的模型
	Prompt   string `json:"prompt"`   // The prompt sent to the LLM / 发送给 LLM 的提示
	Response string `json:"response"` // The raw response from the LLM / 从 LLM 收到的原始响应
	// Potentially add token usage, latency, etc.
	// 可以添加 token 使用量, 延迟等。
}

// KnowledgeBaseHit represents a relevant entry found in the knowledge base.
// KnowledgeBaseHit 表示在知识库中找到的相关条目。
type KnowledgeBaseHit struct {
	ID      string  `json:"id"`      // ID of the knowledge entry / 知识条目的 ID
	Source  string  `json:"source"`  // Source document or origin / 来源文档或出处
	Content string  `json:"content"` // Relevant content snippet / 相关的片段内容
	Score   float64 `json:"score"`   // Relevance score / 相关性分数
	// Potentially add link to original document
	// 可以添加原始文档链接
}

// BusinessData represents generic data received from a Business SDK.
// BusinessData 表示从业务 SDK 接收到的通用数据。
type BusinessData struct {
	Source     BusinessSDKEndpoint    `json:"source"`     // The business service endpoint it came from / 数据来源的业务服务终点
	Timestamp  time.Time              `json:"timestamp"`  // Time when the data was collected / 数据采集时间
	DataType   enum.DataSourceType    `json:"dataType"`   // Type of data (e.g., Log, Metric, Event, Status) / 数据类型 (例如, 日志, 指标, 事件, 状态)
	Content    map[string]interface{} `json:"content"`    // The actual data payload / 实际数据载荷
	ResourceID string                 `json:"resourceId"` // Identifier for the specific business resource (e.g., user ID, order ID) / 特定业务资源的标识符 (例如, 用户 ID, 订单 ID)
}
