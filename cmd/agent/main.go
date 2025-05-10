package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"github.com/turtacn/chasi-sreagent/pkg/common/types/enum"
	"github.com/turtacn/chasi-sreagent/pkg/framework/action"
	"github.com/turtacn/chasi-sreagent/pkg/framework/analyzer"
	"github.com/turtacn/chasi-sreagent/pkg/framework/datacollector"
	"github.com/turtacn/chasi-sreagent/pkg/framework/llm"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/turtacn/chasi-sreagent/pkg/common/errors"
	"github.com/turtacn/chasi-sreagent/pkg/common/log"
	"github.com/turtacn/chasi-sreagent/pkg/common/types"
	"github.com/turtacn/chasi-sreagent/pkg/framework/engine"
	// Import concrete implementations to trigger their init() functions for registration
	// 导入具体实现以触发其 init() 函数进行注册
	k8saction "github.com/turtacn/chasi-sreagent/pkg/actions/k8s" // Need to import for RegisterRestartPodAction
	_ "github.com/turtacn/chasi-sreagent/pkg/analyzers/business"
	_ "github.com/turtacn/chasi-sreagent/pkg/analyzers/k8s"
	businessdatacollector "github.com/turtacn/chasi-sreagent/pkg/datacollectors/business" // Need to import for RegisterBusinessCollector
	k8sdatacollector "github.com/turtacn/chasi-sreagent/pkg/datacollectors/k8s"           // Need to import for RegisterK8sCollector
	vectorkb "github.com/turtacn/chasi-sreagent/pkg/knowledgebases/vector"                // Need to import for RegisterVectorDBKnowledgeBase
	deepseek "github.com/turtacn/chasi-sreagent/pkg/llmproviders/deepseek"                // Need to import for RegisterDeepSeekProvider
	localai "github.com/turtacn/chasi-sreagent/pkg/llmproviders/localai"                  // Need to import for RegisterLocalAIProvider

	"go.uber.org/zap"
	"gopkg.in/yaml.v2" // Using yaml.v2 for config parsing / 使用 yaml.v2 进行配置解析
)

// Agent main entry point.
// 代理主入口点。
func main() {
	// 1. Load configuration
	// 1. 加载配置
	configPath := flag.String("config", types.DefaultConfigPath, "Path to the configuration file")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		// Use a basic logger to report config loading failure
		// 使用基本 logger 报告配置加载失败
		log.L().Fatal("Failed to load configuration", zap.String("configPath", *configPath), zap.Error(err))
		os.Exit(1)
	}

	// 2. Initialize logging with config
	// 2. 使用配置初始化日志记录
	if err := log.Init(&cfg.Log); err != nil {
		// If logger init fails, fall back to printing to stderr
		// 如果 logger 初始化失败，回退到打印到 stderr
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	logger := log.L().With(zap.String("component", "agent"))
	logger.Info("Configuration loaded successfully")

	// Context with cancellation for graceful shutdown
	// 带有取消功能的 context，用于优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Put config in context
	// 将配置放入 context
	ctx = context.WithValue(ctx, types.ContextKeyConfig, cfg)

	// 3. Initialize Dependencies and Framework Components
	// 3. 初始化依赖项和框架组件

	// Initialize Data Collectors
	// 初始化数据采集器
	k8sCollector, err := k8sdatacollector.NewK8sDataCollector(&cfg.Kubernetes)
	if err != nil {
		logger.Fatal("Failed to initialize Kubernetes data collector", zap.Error(err))
	}
	k8sdatacollector.RegisterK8sCollector(k8sCollector)
	logger.Info("Kubernetes data collector initialized and registered")

	businessCollector, err := businessdatacollector.NewBusinessDataCollector(&cfg.BusinessSDK)
	if err != nil {
		logger.Fatal("Failed to initialize Business data collector", zap.Error(err))
	}
	businessdatacollector.RegisterBusinessCollector(businessCollector)
	logger.Info("Business data collector initialized and registered")

	// Initialize LLM Provider
	// 初始化 LLM 提供商
	var llmProvider llm.LLM
	switch cfg.LLM.Provider {
	case localai.LocalAIProviderInstance.Name(): // Using the registered name constant
		providerCfg := &types.LLMProviderConfig{
			URL:    cfg.LLM.LocalAI.URL,
			Model:  cfg.LLM.LocalAI.Model,
			APIKey: cfg.LLM.LocalAI.APIKey,
		}
		llmProvider, err = localai.NewLocalAIProvider(providerCfg, cfg.LLM.Timeout)
		if err != nil {
			logger.Fatal("Failed to initialize LocalAI provider", zap.Error(err))
		}
		localai.RegisterLocalAIProvider(llmProvider.(*localai.LocalAIProvider)) // Cast back to concrete type for registration
	case deepseek.DeepSeekProviderInstance.Name(): // Using the registered name constant
		providerCfg := &types.LLMProviderConfig{
			URL:    cfg.LLM.DeepSeek.URL,
			Model:  cfg.LLM.DeepSeek.Model,
			APIKey: cfg.LLM.DeepSeek.APIKey,
		}
		llmProvider, err = deepseek.NewDeepSeekProvider(providerCfg, cfg.LLM.Timeout)
		if err != nil {
			logger.Fatal("Failed to initialize DeepSeek provider", zap.Error(err))
		}
		deepseek.RegisterDeepSeekProvider(llmProvider.(*deepseek.DeepSeekProvider)) // Cast back to concrete type for registration
	// TODO: Add other LLM providers
	default:
		logger.Fatal("Unsupported LLM provider configured", zap.String("provider", cfg.LLM.Provider))
	}
	logger.Info("LLM provider initialized and registered", zap.String("provider", llmProvider.Name()))

	// Initialize Knowledge Base (Optional)
	// 初始化知识库 (可选)
	var knowledgeBase kb.KnowledgeBase
	if cfg.KnowledgeBase.Enabled {
		switch cfg.KnowledgeBase.Provider {
		case vectorkb.VectorDBKnowledgeBaseInstance.Name(): // Using the registered name constant
			// Need to pass embedding provider to KB initialization if embedding is client-side
			// 如果 embedding 是客户端完成，需要将 embedding 提供商传递给知识库初始化
			kbCfg := &cfg.KnowledgeBase // Pass sub-config
			knowledgeBase, err = vectorkb.NewVectorDBKnowledgeBase(kbCfg)
			if err != nil {
				logger.Fatal("Failed to initialize Vector DB knowledge base", zap.Error(err))
			}
			vectorkb.RegisterVectorDBKnowledgeBase(knowledgeBase.(*vectorkb.VectorDBKnowledgeBase)) // Cast back for registration
		// TODO: Add other KB providers
		default:
			logger.Fatal("Unsupported knowledge base provider configured", zap.String("provider", cfg.KnowledgeBase.Provider))
		}
		logger.Info("Knowledge base initialized and registered", zap.String("provider", knowledgeBase.Name()))
	} else {
		logger.Info("Knowledge base is disabled")
	}

	// Initialize Actions
	// 初始化动作
	// Some actions might need dependencies (like K8s client). Pass them here.
	// 某些动作可能需要依赖项 (如 K8s 客户端)。在这里传递。
	restartPodAction, err := k8saction.NewRestartPodAction( /* Pass dependencies like k8sCollector clients */ )
	if err != nil {
		logger.Fatal("Failed to initialize Restart Pod action", zap.Error(err))
	}
	k8saction.RegisterRestartPodAction(restartPodAction)
	logger.Info("K8s Restart Pod action initialized and registered")

	// Get all registered Analyzers, DataCollectors, Actions
	// 获取所有已注册的分析器、数据采集器、动作
	enabledAnalyzers, err := analyzer.GetEnabledAnalyzers(&cfg.Analysis)
	if err != nil {
		logger.Fatal("Failed to get enabled analyzers", zap.Error(err))
	}
	allDataCollectors := datacollector.GetDataCollectorsByType(enum.DataSourceTypeKubernetesAPI)                            // Example: get K8s collectors
	allDataCollectors = append(allDataCollectors, datacollector.GetDataCollectorsByType(enum.DataSourceTypeBusinessSDK)...) // Example: get Business collectors
	allActions := action.ListActions()                                                                                      // Get names of all registered actions. Need to get the instances by name.

	// TODO: Need to retrieve action instances by name from registry
	// TODO: 需要从注册表按名称检索动作实例
	var actionInstances []action.Action
	for _, name := range allActions {
		if act, found := action.GetAction(name); found {
			actionInstances = append(actionInstances, act)
		}
	}
	logger.Info("Retrieved registered components",
		zap.Int("analyzers", len(enabledAnalyzers)),
		zap.Int("dataCollectors", len(allDataCollectors)),
		zap.Int("actions", len(actionInstances)),
	)

	// Create the Engine
	// 创建引擎
	sreEngine, err := engine.NewSREAgentEngine(
		cfg,
		allDataCollectors, // Pass all registered collectors
		enabledAnalyzers,
		knowledgeBase,   // Pass initialized KB instance (can be nil)
		llmProvider,     // Pass initialized LLM provider
		actionInstances, // Pass all registered actions
	)
	if err != nil {
		logger.Fatal("Failed to create SRE Agent Engine", zap.Error(err))
	}
	logger.Info("SRE Agent Engine created")

	// 4. Start the main agent loop (e.g., periodic analysis)
	// 4. 启动主代理循环 (例如, 周期性分析)
	go func() {
		// Initial delay before the first run (optional)
		// 第一次运行前的初始延迟 (可选)
		time.Sleep(5 * time.Second)

		ticker := time.NewTicker(cfg.Analysis.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Info("Agent loop received shutdown signal, stopping")
				return
			case <-ticker.C:
				// Run analysis task
				// 运行分析任务
				analysisCtx, analysisCancel := context.WithTimeout(ctx, 30*time.Minute)                    // Timeout for analysis run
				analysisCtx = context.WithValue(analysisCtx, types.ContextKeyTraceID, uuid.New().String()) // Add a trace ID
				logger.Info("Starting scheduled analysis and diagnosis run", zap.String("traceID", analysisCtx.Value(types.ContextKeyTraceID).(string)))

				analysisResult, err := sreEngine.RunAnalysis(analysisCtx, nil) // Pass options if needed
				if err != nil {
					logger.Error("Analysis run failed", zap.Error(err), zap.String("traceID", analysisCtx.Value(types.ContextKeyTraceID).(string)))
					analysisCancel()
					continue // Continue to next tick / 继续下一个周期
				}
				logger.Info("Analysis run completed", zap.Int("issuesFound", len(analysisResult.Issues)), zap.String("traceID", analysisCtx.Value(types.ContextKeyTraceID).(string)))

				if len(analysisResult.Issues) > 0 {
					// Run diagnosis task if issues are found
					// 如果找到问题，运行诊断任务
					diagnosisCtx, diagnosisCancel := context.WithTimeout(analysisCtx, 10*time.Minute) // Timeout for diagnosis
					diagnosisResult, err := sreEngine.RunDiagnosis(diagnosisCtx, analysisResult)
					if err != nil {
						logger.Error("Diagnosis run failed", zap.Error(err), zap.String("traceID", analysisCtx.Value(types.ContextKeyTraceID).(string)))
						diagnosisCancel()
						// Continue to next tick / 继续下一个周期
					} else {
						logger.Info("Diagnosis run completed", zap.String("rootCause", diagnosisResult.RootCause), zap.Int("suggestions", len(diagnosisResult.Suggestions)), zap.String("traceID", diagnosisCtx.Value(types.ContextKeyTraceID).(string)))

						// TODO: Output results and suggestions (e.g., log, send notification, expose via API)
						// TODO: 输出结果和建议 (例如, 记录日志, 发送通知, 通过 API 暴露)
						// For now, just log them
						// 目前，只记录日志
						logger.Info("Diagnosis Result:", zap.Any("result", diagnosisResult), zap.String("traceID", diagnosisCtx.Value(types.ContextKeyTraceID).(string)))

						// Plan and potentially execute automated actions
						// 规划并可能执行自动化动作
						suggestions, err := sreEngine.SuggestActions(diagnosisCtx, diagnosisResult)
						if err != nil {
							logger.Error("Action planning failed", zap.Error(err), zap.String("traceID", diagnosisCtx.Value(types.ContextKeyTraceID).(string)))
						} else {
							logger.Info("Action planning completed", zap.Int("suggestions", len(suggestions)), zap.String("traceID", diagnosisCtx.Value(types.ContextKeyTraceID).(string)))
							for _, s := range suggestions {
								logger.Info("Action Suggestion:", zap.Any("suggestion", s), zap.String("traceID", diagnosisCtx.Value(types.ContextKeyTraceID).(string)))
								if s.ActionType == enum.ActionTypeAutomated && cfg.Actions.Enabled {
									// Execute automated action
									// 执行自动化动作
									execCtx, execCancel := context.WithTimeout(diagnosisCtx, 5*time.Minute) // Timeout for action execution
									execResult, execErr := sreEngine.ExecuteAction(execCtx, s)
									if execErr != nil {
										logger.Error("Automated action execution failed", zap.Error(execErr), zap.String("action", s.Name()), zap.String("traceID", execCtx.Value(types.ContextKeyTraceID).(string)))
									} else {
										logger.Info("Automated action executed successfully", zap.String("action", s.Name()), zap.String("result", execResult), zap.String("traceID", execCtx.Value(types.ContextKeyTraceID).(string)))
									}
									execCancel()
								}
							}
						}
					}
					diagnosisCancel() // Cancel diagnosis context / 取消诊断 context
				} else {
					logger.Info("No issues found, skipping diagnosis and actions", zap.String("traceID", analysisCtx.Value(types.ContextKeyTraceID).(string)))
				}

				analysisCancel() // Cancel analysis context / 取消分析 context
			}
		}
	}()

	// 5. Handle graceful shutdown
	// 5. 处理优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Received shutdown signal, initiating graceful shutdown")
	cancel() // Signal goroutines to stop / 通知 goroutine 停止

	// Wait for a short period for goroutines to finish
	// 等待短时间让 goroutine 完成
	time.Sleep(5 * time.Second) // TODO: Implement proper waiting for goroutines to exit
	logger.Info("Agent shutting down")
}

// loadConfig loads the configuration from the specified path.
// loadConfig 从指定路径加载配置。
func loadConfig(configPath string) (*types.Config, error) {
	// Read the file
	// 读取文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		// If file not found, try default path or return error
		// 如果文件未找到，尝试默认路径或返回错误
		if os.IsNotExist(err) && configPath != types.DefaultConfigPath {
			log.L().Warn("Config file not found, trying default path", zap.String("path", configPath), zap.String("defaultPath", types.DefaultConfigPath))
			data, err = os.ReadFile(types.DefaultConfigPath)
			if err != nil {
				return nil, errors.Wrap(errors.ErrorCodeConfigLoadingFailed, "failed to read default config file", err, types.DefaultConfigPath)
			}
			configPath = types.DefaultConfigPath // Update path if default is used
		} else if err != nil {
			return nil, errors.Wrap(errors.ErrorCodeConfigLoadingFailed, "failed to read config file", err, configPath)
		}
	}

	var cfg types.Config
	// Unmarshal YAML
	// Unmarshal YAML
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, errors.Wrap(errors.ErrorCodeConfigLoadingFailed, "failed to unmarshal config YAML", err, configPath)
	}

	// Set default timeouts if not specified
	// 如果未指定，设置默认超时时间
	if cfg.LLM.Timeout == 0 {
		cfg.LLM.Timeout = types.DefaultLLMTimeout * time.Second
	}
	if cfg.BusinessSDK.Timeout == 0 {
		cfg.BusinessSDK.Timeout = types.DefaultBusinessSDKTimeout * time.Second
	}
	if cfg.Analysis.Interval == 0 {
		cfg.Analysis.Interval = types.DefaultAnalysisInterval * time.Second
	}

	return &cfg, nil
}
