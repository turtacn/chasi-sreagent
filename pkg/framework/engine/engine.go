package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid" // Using uuid for unique IDs / 使用 uuid 生成唯一 ID
	"github.com/turtacn/chasi-sreagent/pkg/common/enum"
	"github.com/turtacn/chasi-sreagent/pkg/common/errors"
	"github.com/turtacn/chasi-sreagent/pkg/common/log"
	"github.com/turtacn/chasi-sreagent/pkg/common/types"
	"github.com/turtacn/chasi-sreagent/pkg/framework/action"
	"github.com/turtacn/chasi-sreagent/pkg/framework/analyzer"
	"github.com/turtacn/chasi-sreagent/pkg/framework/datacollector"
	"github.com/turtacn/chasi-sreagent/pkg/framework/knowledgebase"
	"github.com/turtacn/chasi-sreagent/pkg/framework/llm"
	"go.uber.org/zap"
)

// Package engine defines the core orchestration engine for the SRE agent.
// 包 engine 定义了 SRE 代理的核心编排引擎。

// Engine is the interface for the main SRE agent orchestration engine.
// Engine 是主 SRE 代理编排引擎的接口。
type Engine interface {
	// RunAnalysis performs a one-time analysis run based on the current configuration.
	// RunAnalysis 根据当前配置执行一次性分析运行。
	// It collects data, runs enabled analyzers, and returns the analysis result.
	// 它收集数据，运行启用的分析器，并返回分析结果。
	RunAnalysis(ctx context.Context, options map[string]interface{}) (*types.AnalysisResult, error)

	// RunDiagnosis performs a diagnosis based on a given analysis result.
	// RunDiagnosis 基于给定的分析结果执行诊断。
	// It uses the LLM and knowledge base to determine root causes and suggest remediations.
	// 它使用 LLM 和知识库来确定根因并建议处置方案。
	RunDiagnosis(ctx context.Context, analysisResult *types.AnalysisResult) (*types.DiagnosisResult, error)

	// SuggestActions plans potential remediation actions based on a diagnosis result.
	// SuggestActions 基于诊断结果规划潜在的处置动作。
	// It returns a list of suggested actions.
	// 它返回一个建议动作列表。
	SuggestActions(ctx context.Context, diagnosisResult *types.DiagnosisResult) ([]types.RemediationSuggestion, error)

	// ExecuteAction attempts to execute a specific automated action.
	// ExecuteAction 尝试执行特定的自动化动作。
	// This should only be called for actions of type ActionTypeAutomated.
	// 只应为 ActionTypeAutomated 类型的动作调用此方法。
	ExecuteAction(ctx context.Context, suggestion types.RemediationSuggestion) (string, error) // Result description or error
}

// SREAgentEngine is the concrete implementation of the Engine interface.
// SREAgentEngine 是 Engine 接口的具体实现。
type SREAgentEngine struct {
	config *types.Config

	// Dependencies
	// 依赖项
	dataCollectors []datacollector.DataCollector
	analyzers      []analyzer.Analyzer
	knowledgeBase  knowledgebase.KnowledgeBase // Optional
	llmProvider    llm.LLM
	actions        []action.Action
	// Potentially add more dependencies like metric clients, notification clients, etc.
	// 可能添加更多依赖项，例如指标客户端、通知客户端等。
}

// NewSREAgentEngine creates a new instance of the SRE Agent Engine.
// NewSREAgentEngine 创建 SRE 代理引擎的新实例。
// It takes the configuration and initialized dependencies.
// 它接收配置和已初始化的依赖项。
func NewSREAgentEngine(
	cfg *types.Config,
	dataCollectors []datacollector.DataCollector,
	analyzers []analyzer.Analyzer,
	kb knowledgebase.KnowledgeBase, // kb can be nil if disabled
	llmProvider llm.LLM,
	actions []action.Action,
) (*SREAgentEngine, error) {
	if cfg == nil {
		return nil, errors.New(errors.ErrorCodeInvalidInput, "config cannot be nil", "")
	}
	if llmProvider == nil {
		return nil, errors.New(errors.ErrorCodeInvalidInput, "LLM provider cannot be nil", "LLM provider is required for diagnosis")
	}

	engine := &SREAgentEngine{
		config:         cfg,
		dataCollectors: dataCollectors,
		analyzers:      analyzers,
		knowledgeBase:  kb,
		llmProvider:    llmProvider,
		actions:        actions,
	}

	log.L().Info("SRE Agent Engine initialized")

	return engine, nil
}

// RunAnalysis performs a one-time analysis run.
// RunAnalysis 执行一次性分析运行。
func (e *SREAgentEngine) RunAnalysis(ctx context.Context, options map[string]interface{}) (*types.AnalysisResult, error) {
	analysisID := uuid.New().String()
	start := time.Now()
	logger := log.LWithContext(ctx).With(zap.String("analysisID", analysisID))
	logger.Info("Starting analysis run")

	result := &types.AnalysisResult{
		ID:           analysisID,
		Timestamp:    start,
		Status:       enum.AnalysisStatusRunning,
		Issues:       []types.Issue{},
		AnalyzersRun: []string{},
	}

	// --- Data Collection ---
	// This is a simplified example. In reality, data collection might need to be coordinated
	// or analyzers might trigger specific data collection needs.
	// 这里是一个简化的示例。实际上，数据收集可能需要协调进行，
	// 或者分析器可能会触发特定的数据收集需求。
	collectedData := make(map[enum.DataSourceType]interface{})
	for _, collector := range e.dataCollectors {
		collectorLogger := logger.With(zap.String("collector", collector.Name()), zap.String("dataType", collector.Type().String()))
		collectorLogger.Debug("Collecting data")
		data, err := collector.Collect(ctx, options)
		if err != nil {
			collectorLogger.Error("Failed to collect data", zap.Error(err))
			// Decide if collection failure is fatal or if analysis can proceed with partial data
			// 决定收集失败是否致命，或者分析是否可以使用部分数据继续进行
			// For now, log error and continue with other collectors
			// 目前，记录错误并继续其他采集器
			continue
		}
		collectedData[collector.Type()] = data
		collectorLogger.Debug("Data collected successfully")
	}

	// --- Analysis Execution ---
	allIssues := []types.Issue{}
	analyzersRun := []string{}

	for _, analyzer := range e.analyzers {
		analyzerLogger := logger.With(zap.String("analyzer", analyzer.Name()))
		analyzerLogger.Debug("Running analyzer")

		// Check if required data is available
		// 检查所需数据是否可用
		missingData := false
		for _, requiredType := range analyzer.RequiredDataSources() {
			if _, ok := collectedData[requiredType]; !ok {
				analyzerLogger.Warn("Skipping analyzer, required data source not collected", zap.String("requiredType", requiredType.String()))
				missingData = true
				break
			}
			// TODO: Pass relevant data from collectedData to the analyzer's Analyze method
			// TODO: 将 collectedData 中相关的数据传递给分析器的 Analyze 方法
			// The current Analyzer interface doesn't take collected data as input directly.
			// This needs refinement. A better approach might be:
			// type Analyzer interface { Analyze(ctx context.Context, data map[enum.DataSourceType]interface{}) ([]types.Issue, error) }
			// Alternatively, analyzers might fetch data themselves using the collectors registry.
			// 当前的 Analyzer 接口不直接接收收集的数据作为输入。
			// 这需要改进。一个更好的方法可能是:
			// type Analyzer interface { Analyze(ctx context.Context, data map[enum.DataSourceType]interface{}) ([]types.Issue, error) }
			// 或者，分析器可以通过采集器注册表自行获取数据。
			// For now, let's assume analyzers fetch data internally using the registered collectors.
			// 目前，我们假定分析器使用已注册的采集器在内部获取数据。
		}

		if missingData {
			result.AnalyzersRun = append(result.AnalyzersRun, analyzer.Name()+" (Skipped)")
			continue
		}

		issues, err := analyzer.Analyze(ctx) // Assuming analyzers fetch data internally
		if err != nil {
			analyzerLogger.Error("Analyzer failed", zap.Error(err))
			// Decide if analyzer failure is fatal to the whole run
			// 决定分析器失败是否对整个运行是致命的
			// For now, log error and continue with other analyzers
			// 目前，记录错误并继续其他分析器
			result.AnalyzersRun = append(result.AnalyzersRun, analyzer.Name()+" (Failed)")
			continue
		}
		allIssues = append(allIssues, issues...)
		result.AnalyzersRun = append(result.AnalyzersRun, analyzer.Name()+" (Completed)")
		analyzerLogger.Debug("Analyzer completed", zap.Int("issuesFound", len(issues)))
	}

	result.Issues = allIssues
	result.Duration = time.Since(start)
	result.Status = enum.AnalysisStatusCompleted
	logger.Info("Analysis run completed", zap.Int("totalIssues", len(allIssues)), zap.Duration("duration", result.Duration))

	return result, nil
}

// RunDiagnosis performs a diagnosis based on analysis results.
// RunDiagnosis 基于分析结果执行诊断。
func (e *SREAgentEngine) RunDiagnosis(ctx context.Context, analysisResult *types.AnalysisResult) (*types.DiagnosisResult, error) {
	diagnosisID := uuid.New().String()
	start := time.Now()
	logger := log.LWithContext(ctx).With(zap.String("diagnosisID", diagnosisID), zap.String("analysisID", analysisResult.ID))
	logger.Info("Starting diagnosis run")

	diagnosis := &types.DiagnosisResult{
		AnalysisResultID:  analysisResult.ID,
		Timestamp:         start,
		Suggestions:       []types.RemediationSuggestion{},
		KnowledgeBaseHits: []types.KnowledgeBaseHit{},
	}

	if len(analysisResult.Issues) == 0 {
		logger.Info("No issues found in analysis, skipping diagnosis")
		diagnosis.RootCause = "No issues detected."
		diagnosis.Duration = time.Since(start)
		return diagnosis, nil
	}

	// --- Prepare Prompt for LLM ---
	// This is a crucial step. The prompt needs to include:
	// - System context (agent's role, environment description - multi-tenant vcluster K8s)
	// - Details of detected issues
	// - Relevant knowledge retrieved from KB (RAG)
	// - Instructions for the LLM (task: root cause analysis, suggest remedies; format: desired output structure)
	// 这是一个关键步骤。提示需要包括:
	// - 系统上下文 (代理的角色, 环境描述 - 多租户 vcluster K8s)
	// - 检测到的问题的详细信息
	// - 从知识库检索到的相关知识 (RAG)
	// - 对 LLM 的指令 (任务: 根因分析, 建议处置方案; 格式: 期望的输出结构)

	promptBuilder := new(StringBuilder)                                                                                                             // Helper to build prompt / 构建提示的助手
	promptBuilder.WriteString("You are an AI SRE agent assisting with troubleshooting Kubernetes issues in a multi-tenant vcluster environment.\n") // System role / 系统角色
	promptBuilder.WriteString("Analyze the following issues detected in the cluster:\n\n")                                                          // Task instruction / 任务指令

	for i, issue := range analysisResult.Issues {
		promptBuilder.WriteString(fmt.Sprintf("Issue %d:\n", i+1))
		promptBuilder.WriteString(fmt.Sprintf("  Name: %s\n", issue.Name))
		promptBuilder.WriteString(fmt.Sprintf("  Severity: %s\n", issue.Severity.String()))
		promptBuilder.WriteString(fmt.Sprintf("  Message: %s\n", issue.Message))
		if issue.Resource != nil {
			promptBuilder.WriteString(fmt.Sprintf("  Resource: Type=%s, Name=%s, Namespace=%s, VCluster=%s\n",
				issue.Resource.Type, issue.Resource.Name, issue.Resource.Namespace, issue.Resource.VCluster))
		}
		// Include other issue context / 包括其他问题上下文
		promptBuilder.WriteString("\n")
	}

	// --- RAG: Retrieve relevant knowledge ---
	// Need to formulate a query based on the issues
	// 需要根据问题构建查询
	kbQuery := fmt.Sprintf("Diagnose Kubernetes issues: %v", analysisResult.Issues) // Simplify query for now
	var kbHits []types.KnowledgeBaseHit
	var kbErr error
	if e.knowledgeBase != nil {
		logger.Debug("Retrieving knowledge from KB for diagnosis")
		// Basic query options, refine later
		// 基本查询选项，后续改进
		kbHits, kbErr = e.knowledgeBase.Retrieve(ctx, kbQuery, map[string]interface{}{"k": 5}) // Get top 5 hits
		if kbErr != nil {
			logger.Error("Failed to retrieve knowledge from KB", zap.Error(kbErr))
			// Continue diagnosis even if KB retrieval fails
			// 即使知识库检索失败也继续诊断
		} else {
			diagnosis.KnowledgeBaseHits = kbHits
			if len(kbHits) > 0 {
				promptBuilder.WriteString("Relevant knowledge from SRE Knowledge Base:\n\n")
				for _, hit := range kbHits {
					promptBuilder.WriteString(fmt.Sprintf("Source: %s (Score: %.2f)\n", hit.Source, hit.Score))
					promptBuilder.WriteString(hit.Content + "\n\n")
				}
			}
			logger.Debug("Knowledge retrieved", zap.Int("hits", len(kbHits)))
		}
	}

	promptBuilder.WriteString("Based on the issues and relevant knowledge, provide:\n")
	promptBuilder.WriteString("1. A concise root cause analysis.\n")
	promptBuilder.WriteString("2. Suggested remediation steps.\n")
	// Instruct LLM on output format (e.g., JSON, Markdown list)
	// 指导 LLM 输出格式 (例如, JSON, Markdown 列表)
	promptBuilder.WriteString("Please provide the root cause analysis as a paragraph and the suggestions as a numbered list.\n") // Simple format instruction

	finalPrompt := promptBuilder.String()
	logger.Debug("Sending prompt to LLM", zap.String("prompt", finalPrompt))

	// --- Call LLM ---
	llmStart := time.Now()
	llmResponse, llmErr := e.llmProvider.GenerateText(
		llm.WithTimeout(ctx, e.config.LLM.Timeout), // Use LLM specific timeout
		finalPrompt,
		nil, // LLM options, e.g., temperature, max tokens
	)
	llmDuration := time.Since(llmStart)

	diagnosis.LLMInteraction = &types.LLMInteractionDetails{
		Provider: e.llmProvider.Name(),
		Model:    e.config.LLM.GetEnabledLLMProvider(e.config.LLM).Model, // Need a helper to get model name
		Prompt:   finalPrompt,
		Response: llmResponse,
		// Add latency, token usage if LLM interface provides them
		// 如果 LLM 接口提供，添加延迟、token 使用量
	}

	if llmErr != nil {
		logger.Error("LLM text generation failed", zap.Error(llmErr))
		diagnosis.RootCause = "Failed to perform diagnosis due to LLM error."
		diagnosis.Error = errors.Wrap(errors.ErrorCodeLLMProviderError, "LLM diagnosis failed", llmErr, "").Error()
		diagnosis.Duration = time.Since(start)
		return diagnosis, diagnosisErr // Return the diagnosis object with partial info and the error
	}
	logger.Debug("LLM response received", zap.Duration("llmDuration", llmDuration))

	// --- Parse LLM Response ---
	// This requires parsing the LLM's natural language response into structured data.
	// This can be complex and might need prompt engineering to guide the LLM output format.
	// 这需要将 LLM 的自然语言响应解析为结构化数据。
	// 这可能很复杂，并且可能需要提示工程来指导 LLM 输出格式。
	parsedRootCause, parsedSuggestions, parseErr := parseLLMDiagnosisResponse(llmResponse) // TODO: Implement this parsing function
	if parseErr != nil {
		logger.Error("Failed to parse LLM response", zap.Error(parseErr))
		// Use raw LLM response or provide a generic error
		// 使用原始 LLM 响应或提供通用错误
		diagnosis.RootCause = "LLM response received but parsing failed: " + llmResponse
		diagnosis.Error = errors.Wrap(errors.ErrorCodeUnknown, "Failed to parse LLM response", parseErr, "").Error()
		// Still proceed to generate suggestions if parsing is partial
		// 如果解析是部分的，仍继续生成建议
	} else {
		diagnosis.RootCause = parsedRootCause
		// Convert parsed suggestions into types.RemediationSuggestion
		// 将解析出的建议转换为 types.RemediationSuggestion
		for _, s := range parsedSuggestions {
			suggestion := types.RemediationSuggestion{
				IssueID:     "N/A", // LLM might provide suggestions for multiple issues, mapping needs logic
				Description: s,
				ActionType:  enum.ActionTypeSuggestion, // Default to suggestion from LLM
				Confidence:  0.8,                       // Placeholder, maybe LLM can provide confidence?
				Source:      "LLM",
			}
			diagnosis.Suggestions = append(diagnosis.Suggestions, suggestion)
		}
	}

	diagnosis.Duration = time.Since(start)
	logger.Info("Diagnosis run completed", zap.Duration("duration", diagnosis.Duration))

	return diagnosis, nil
}

// parseLLMDiagnosisResponse is a placeholder for the function that parses the LLM's response.
// parseLLMDiagnosisResponse 是一个占位符函数，用于解析 LLM 的响应。
// Its implementation will depend heavily on the expected LLM output format.
// 它的实现将很大程度上取决于期望的 LLM 输出格式。
func parseLLMDiagnosisResponse(response string) (string, []string, error) {
	// TODO: Implement robust parsing logic here
	// TODO: 在这里实现健壮的解析逻辑

	// This is a very basic example assuming a specific format:
	// Root Cause: <root cause paragraph>
	// Suggestions:
	// 1. Suggestion 1
	// 2. Suggestion 2
	// ...

	rootCause := "Parsing logic not yet implemented. Raw LLM response:\n" + response
	suggestions := []string{"Parsing failed, cannot extract suggestions from raw response."}
	// Return nil error for now to allow basic flow
	// 目前返回 nil 错误以允许基本流程
	return rootCause, suggestions, nil // Placeholder implementation
}

// SuggestActions plans potential remediation actions based on a diagnosis result.
// SuggestActions 基于诊断结果规划潜在的处置动作。
// This function leverages the registered Action implementations.
// 此函数利用已注册的 Action 实现。
func (e *SREAgentEngine) SuggestActions(ctx context.Context, diagnosisResult *types.DiagnosisResult) ([]types.RemediationSuggestion, error) {
	logger := log.LWithContext(ctx).With(zap.String("diagnosisID", diagnosisResult.AnalysisResultID))
	logger.Info("Planning remediation actions")

	allSuggestions := []types.RemediationSuggestion{}

	if !e.config.Actions.Enabled {
		logger.Info("Automated actions are disabled, only suggesting manual steps from LLM")
		// If automated actions are disabled, maybe filter out ActionTypeAutomated from LLM suggestions
		// 如果自动化动作被禁用，可以过滤掉 LLM 建议中的 ActionTypeAutomated 类型
		// Or just rely on the LLM output directly as manual suggestions.
		// 或者直接依赖 LLM 输出作为人工建议。
		// For now, just return the suggestions generated by LLM in RunDiagnosis.
		// 目前，只返回 RunDiagnosis 中由 LLM 生成的建议。
		// In a more complex setup, actions could augment/refine LLM suggestions.
		// 在更复杂的设置中，actions 可以增强/改进 LLM 建议。
		// Let's assume for now the Actions framework primarily handles automated execution,
		// while LLM generates manual suggestions.
		// 现在，我们假设 Actions 框架主要处理自动化执行，而 LLM 生成人工建议。
		// We still iterate through registered actions to see if any can *plan* automated actions
		// based on the diagnosis, even if execution is disabled.
		// 我们仍然遍历已注册的动作，看看是否有任何动作可以基于诊断结果 *规划* 自动化动作，
		// 即使执行是禁用的。
	}

	// Iterate through registered actions to see if they are relevant
	// 遍历已注册的动作，看看它们是否相关
	for _, act := range e.actions {
		actionLogger := logger.With(zap.String("action", act.Name()))
		actionLogger.Debug("Planning with action")

		// Only plan automated actions if enabled
		// 只有在启用自动化动作时才规划自动化动作
		if act.Type() == enum.ActionTypeAutomated && !e.config.Actions.Enabled {
			actionLogger.Debug("Skipping automated action planning as actions are disabled")
			continue
		}

		isRelevant, suggestion, err := act.Plan(ctx, diagnosisResult)
		if err != nil {
			actionLogger.Error("Action planning failed", zap.Error(err))
			// Continue with other actions
			// 继续其他动作
			continue
		}

		if isRelevant {
			allSuggestions = append(allSuggestions, suggestion)
			actionLogger.Debug("Action planned", zap.Any("suggestion", suggestion))
		}
	}

	// Combine LLM suggestions with action-planned suggestions
	// 合并 LLM 建议和动作规划的建议
	// Avoid duplicates if possible. Logic needed here.
	// 尽可能避免重复。这里需要逻辑。
	// For simplicity, let's just add LLM suggestions if they exist.
	// 为简化起见，如果存在 LLM 建议，就直接添加。
	// A real implementation needs de-duplication and merging logic.
	// 实际实现需要去重和合并逻辑。

	combinedSuggestions := diagnosisResult.Suggestions // Start with LLM suggestions / 从 LLM 建议开始
	// Add unique suggestions from actions
	// 添加来自动作的唯一建议
	suggestionMap := make(map[string]struct{}) // Use map to track unique suggestions / 使用 map 跟踪唯一建议
	for _, s := range combinedSuggestions {
		// Basic uniqueness check, could be improved based on suggestion content/payload
		// 基本的唯一性检查，可以根据建议内容/载荷改进
		suggestionMap[s.Description] = struct{}{}
	}

	for _, s := range allSuggestions {
		if _, exists := suggestionMap[s.Description]; !exists {
			combinedSuggestions = append(combinedSuggestions, s)
			suggestionMap[s.Description] = struct{}{}
		}
	}

	logger.Info("Remediation actions planned", zap.Int("totalSuggestions", len(combinedSuggestions)))
	return combinedSuggestions, nil
}

// ExecuteAction attempts to execute a specific automated action.
// ExecuteAction 尝试执行特定的自动化动作。
func (e *SREAgentEngine) ExecuteAction(ctx context.Context, suggestion types.RemediationSuggestion) (string, error) {
	logger := log.LWithContext(ctx).With(zap.String("suggestionID", suggestion.IssueID), zap.String("actionType", suggestion.ActionType.String()))
	logger.Info("Attempting to execute action")

	if suggestion.ActionType != enum.ActionTypeAutomated {
		logger.Warn("Action is not automated, cannot execute", zap.Any("suggestion", suggestion))
		return "", errors.New(errors.ErrorCodeInvalidInput, "action type is not automated", fmt.Sprintf("suggestion ID %s has type %s", suggestion.IssueID, suggestion.ActionType.String()))
	}

	if !e.config.Actions.Enabled {
		logger.Error("Automated actions are disabled, cannot execute")
		return "", errors.New(errors.ErrorCodeActionExecutionFailed, "automated actions disabled", "")
	}

	// Find the registered action implementation
	// 找到已注册的动作实现
	// Need to map suggestion back to an action name. This mapping is missing.
	// 需要将建议映射回动作名称。这个映射是缺失的。
	// The Plan method should ideally return the name of the action that planned it.
	// Plan 方法理想情况下应该返回规划它的动作的名称。
	// Let's assume suggestion.Source might contain the action name, or add a field to RemediationSuggestion.
	// 假设 suggestion.Source 可能包含动作名称，或者在 RemediationSuggestion 中添加一个字段。
	// For now, we'll need a placeholder logic to find the action.
	// 目前，我们需要一个占位逻辑来查找动作。

	actionName := suggestion.Source // Assuming Source contains action name for simplicity
	act, found := action.GetAction(actionName)
	if !found {
		logger.Error("Automated action implementation not found", zap.String("actionName", actionName))
		return "", errors.New(errors.ErrorCodeNotFound, "action implementation not found", fmt.Sprintf("action '%s' not registered for execution", actionName))
	}

	if act.Type() != enum.ActionTypeAutomated {
		logger.Error("Found action is not of automated type", zap.String("actionName", actionName), zap.String("foundType", act.Type().String()))
		return "", errors.New(errors.ErrorCodeInvalidInput, "registered action is not automated", fmt.Sprintf("action '%s' registered but is not automated type", actionName))
	}

	// Execute the action
	// 执行动作
	result, err := act.Execute(ctx, suggestion)
	if err != nil {
		logger.Error("Action execution failed", zap.Error(err), zap.String("actionName", actionName))
		return "", errors.Wrap(errors.ErrorCodeActionExecutionFailed, "failed to execute action", err, fmt.Sprintf("action '%s' failed", actionName))
	}

	logger.Info("Action executed successfully", zap.String("actionName", actionName), zap.String("result", result))
	return result, nil
}

// StringBuilder is a helper to build strings efficiently.
// StringBuilder 是一个用于高效构建字符串的助手。
type StringBuilder struct {
	s string
}

func (sb *StringBuilder) WriteString(str string) {
	sb.s += str
}

func (sb *StringBuilder) String() string {
	return sb.s
}
