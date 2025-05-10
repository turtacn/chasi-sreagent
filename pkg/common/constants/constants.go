package constants

// Package common/constants defines shared constants used throughout the chasi-sreagent project.
// 包 common/constants 定义了 chasi-sreagent 项目中使用的共享常量。

const (
	// ProjectName is the name of the project.
	// ProjectName 是项目的名称。
	ProjectName = "chasi-sreagent"

	// DefaultConfigPath is the default path for the configuration file.
	// DefaultConfigPath 是配置文件的默认路径。
	DefaultConfigPath = "/etc/chasi-sreagent/config.yaml"

	// DefaultLLMProvider is the default LLM provider if not specified in config.
	// DefaultLLMProvider 是配置文件中未指定时使用的默认 LLM 提供商。
	DefaultLLMProvider = "localai"

	// DefaultLLMTimeout is the default timeout for LLM API calls.
	// DefaultLLMTimeout 是 LLM API 调用的默认超时时间。
	DefaultLLMTimeout = 60 // seconds / 秒

	// DefaultBusinessSDKTimeout is the default timeout for calling business SDK endpoints.
	// DefaultBusinessSDKTimeout 是调用业务 SDK 终点的默认超时时间。
	DefaultBusinessSDKTimeout = 10 // seconds / 秒

	// DefaultAnalysisInterval is the default interval for continuous analysis.
	// DefaultAnalysisInterval 是持续分析的默认间隔。
	DefaultAnalysisInterval = 5 * 60 // seconds / 秒 (5 minutes)

	// VClusterKubeConfigKey is the key used in the vcluster config map entry for the kubeconfig.
	// VClusterKubeConfigKey 是 vcluster 配置映射条目中用于存储 kubeconfig 的键。
	VClusterKubeConfigKey = "config"
)

// Analyzer names
// 分析器名称
const (
	// AnalyzerKubernetesPod is the name for the Kubernetes Pod analyzer.
	// AnalyzerKubernetesPod 是 Kubernetes Pod 分析器的名称。
	AnalyzerKubernetesPod = "kubernetes-pod-analyzer"

	// AnalyzerBusinessLog is the name for the business log analyzer.
	// AnalyzerBusinessLog 是业务日志分析器的名称。
	AnalyzerBusinessLog = "business-log-analyzer"

	// Add other analyzer names here
	// 在这里添加其他分析器名称
)

// LLM Provider names
// LLM 提供商名称
const (
	// LLMProviderLocalAI is the name for the LocalAI LLM provider.
	// LLMProviderLocalAI 是 LocalAI LLM 提供商的名称。
	LLMProviderLocalAI = "localai"

	// LLMProviderDeepSeek is the name for the DeepSeek LLM provider.
	// LLMProviderDeepSeek 是 DeepSeek LLM 提供商的名称。
	LLMProviderDeepSeek = "deepseek"

	// LLMProviderOpenAI is the name for the OpenAI LLM provider.
	// LLMProviderOpenAI 是 OpenAI LLM 提供商的名称。
	LLMProviderOpenAI = "openai"

	// Add other LLM provider names here
	// 在这里添加其他 LLM 提供商名称
)

// Knowledge Base Provider names
// 知识库提供商名称
const (
	// KBProviderVectorDB is the name for the vector database knowledge base provider.
	// KBProviderVectorDB 是向量数据库知识库提供商的名称。
	KBProviderVectorDB = "vector-db"
)

// Business SDK Discovery Methods
// 业务 SDK 发现方法
const (
	// BusinessSDKDiscoveryKubernetesService means discovering business SDK endpoints via Kubernetes service labels.
	// BusinessSDKDiscoveryKubernetesService 意味着通过 Kubernetes 服务标签发现业务 SDK 终点。
	BusinessSDKDiscoveryKubernetesService = "kubernetes-service"

	// BusinessSDKDiscoveryStaticList means providing a static list of business SDK endpoints in the config.
	// BusinessSDKDiscoveryStaticList 意味着在配置中提供业务 SDK 终点的静态列表。
	BusinessSDKDiscoveryStaticList = "static-list"
)
