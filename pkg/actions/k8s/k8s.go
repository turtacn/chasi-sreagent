package k8s

import (
	"context"
	"fmt"
	"strings"

	"github.com/turtacn/chasi-sreagent/pkg/common/enum"
	"github.com/turtacn/chasi-sreagent/pkg/common/errors"
	"github.com/turtacn/chasi-sreagent/pkg/common/log"
	"github.com/turtacn/chasi-sreagent/pkg/common/types"
	"github.com/turtacn/chasi-sreagent/pkg/framework/action"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	// Add import for the K8s data collector to get clients or data
	// 添加 K8s 数据采集器的导入，以获取客户端或数据
	// k8sdatacollector "github.com/turtacn/chasi-sreagent/pkg/datacollectors/k8s" // Assuming collector instance is accessible
)

// Package k8s provides remediation actions for Kubernetes resources.
// 包 k8s 提供 Kubernetes 资源的处置动作。

// RestartPodAction is an action to restart a Kubernetes Pod by deleting it.
// RestartPodAction 是一个通过删除 Pod 来重启 Kubernetes Pod 的动作。
type RestartPodAction struct {
	// Add dependencies here, e.g., Kubernetes client or access to K8s data collector clients
	// 在这里添加依赖项，例如，Kubernetes 客户端或访问 K8s 数据采集器的客户端
	// k8sClient kubernetes.Interface // Example, but ideally use collector's client management
}

// Ensure RestartPodAction implements the action.Action interface.
// 确保 RestartPodAction 实现了 action.Action 接口。
var _ action.Action = &RestartPodAction{}

// NewRestartPodAction creates a new RestartPodAction instance.
// NewRestartPodAction 创建一个新的 RestartPodAction 实例。
// It needs access to Kubernetes clients.
// 它需要访问 Kubernetes 客户端。
func NewRestartPodAction( /* Add dependencies like K8s client access */ ) (*RestartPodAction, error) {
	// TODO: Pass Kubernetes client access, maybe via the K8sDataCollector instance
	// TODO: 传递 Kubernetes 客户端访问，可能通过 K8sDataCollector 实例

	log.L().Info("Initialized K8s Restart Pod Action")
	return &RestartPodAction{}, nil // Placeholder / 占位符
}

// Name returns the name of the action.
// Name 返回动作的名称。
func (a *RestartPodAction) Name() string {
	return "restart-pod"
}

// Description returns a brief description.
// Description 返回简要描述。
func (a *RestartPodAction) Description() string {
	return "Restarts a Kubernetes Pod by deleting it. Requires the deployment/statefulset to recreate it."
}

// Type returns the action type.
// Type 返回动作类型。
func (a *RestartPodAction) Type() enum.ActionType {
	// This action can be automated, but should be suggested first unless configured otherwise.
	// Let's mark it as automated type, but execution logic needs checks.
	// 此动作可以自动化，但除非另有配置，否则应首先建议。
	// 我们将其标记为自动化类型，但执行逻辑需要检查。
	return enum.ActionTypeAutomated
}

// Plan determines if this action is relevant for the diagnosis result.
// Plan 根据诊断结果确定此动作是否相关。
// It might look for issues related to crashing pods or LLM suggestions mentioning pod restart.
// 它可能会查找与崩溃 Pod 相关的问题，或 LLM 建议中提到 Pod 重启的建议。
func (a *RestartPodAction) Plan(ctx context.Context, diagnosis *types.DiagnosisResult) (bool, types.RemediationSuggestion, error) {
	logger := log.LWithContext(ctx).With(zap.String("action", a.Name()), zap.String("diagnosisID", diagnosis.AnalysisResultID))
	logger.Debug("Planning action")

	// Example Plan Logic:
	// Check if any issue in the analysis result is a PodCrashLoopBackOff
	// 检查分析结果中是否有任何问题是 PodCrashLoopBackOff
	// Check if the LLM suggestion includes restarting a pod
	// 检查 LLM 建议是否包含重启 Pod

	isRelevant := false
	var targetResource *types.IssueResource

	// Check analysis issues
	// 检查分析问题
	for _, issue := range diagnosis.Issues {
		if issue.Name == types.AnalyzerKubernetesPod && issue.Resource != nil && issue.Resource.Type == "Pod" {
			// More specific checks needed, e.g., look into issue.Message or issue.Context
			// 需要更具体的检查，例如，查看 issue.Message 或 issue.Context
			if strings.Contains(issue.Message, "CrashLoopBackOff") {
				isRelevant = true
				targetResource = issue.Resource
				logger.Debug("Action relevant based on issue", zap.String("issueID", issue.ID))
				break // Found a relevant issue, plan the action
			}
		}
	}

	// Also check LLM suggestions if they exist (as they might suggest actions not directly tied to specific issues)
	// 如果存在 LLM 建议，也检查它们 (因为它们可能建议不直接与特定问题关联的动作)
	if !isRelevant && diagnosis.LLMInteraction != nil && diagnosis.LLMInteraction.Response != "" {
		// Perform basic text search in LLM response or parsed suggestions
		// 在 LLM 响应或解析出的建议中执行基本文本搜索
		if strings.Contains(diagnosis.LLMInteraction.Response, "restart the pod") || strings.Contains(diagnosis.RootCause, "pod crashing") { // Very basic check
			isRelevant = true
			// If LLM suggests it, we need to identify WHICH pod. This is hard without structure.
			// If LLM suggests it, assume the target resource is one of the resources in the analysis issues.
			// 如果 LLM 建议，我们需要识别是哪个 Pod。如果没有结构，这很难。
			// 如果 LLM 建议，假设目标资源是分析问题中的某个资源。
			// For this stub, let's just use the targetResource found from issues, or pick the first Pod issue's resource.
			// 对于这个占位符，我们暂时使用从问题中找到的 targetResource，或者选择第一个 Pod 问题的资源。
			if targetResource == nil && len(diagnosis.Issues) > 0 {
				for _, issue := range diagnosis.Issues {
					if issue.Resource != nil && issue.Resource.Type == "Pod" {
						targetResource = issue.Resource
						break
					}
				}
			}
			logger.Debug("Action relevant based on LLM suggestion")
		}
	}

	if isRelevant && targetResource != nil {
		// Action is relevant, prepare the suggestion
		// 动作相关，准备建议
		suggestion := types.RemediationSuggestion{
			IssueID:     diagnosis.AnalysisResultID, // Or the specific issue ID it plans for / 或其规划的具体问题 ID
			Description: fmt.Sprintf("Restart Pod '%s' in namespace '%s' (%s vcluster) by deleting it.", targetResource.Name, targetResource.Namespace, targetResource.VCluster),
			ActionType:  a.Type(),
			Command:     fmt.Sprintf("kubectl delete pod %s -n %s", targetResource.Name, targetResource.Namespace), // Example CLI command
			Payload: map[string]interface{}{ // Payload for automated execution / 自动化执行的载荷
				"resourceType": targetResource.Type,
				"resourceName": targetResource.Name,
				"namespace":    targetResource.Namespace,
				"vcluster":     targetResource.VCluster,
				"resourceUID":  targetResource.UID, // Use UID for safer identification
			},
			Confidence: 0.7,      // Confidence in the suggestion / 对建议的置信度
			Source:     a.Name(), // The action that planned this / 规划此动作的动作
		}
		logger.Debug("Action planned successfully")
		return true, suggestion, nil
	}

	logger.Debug("Action not relevant for this diagnosis")
	return false, types.RemediationSuggestion{}, nil
}

// Execute performs the action (deleting the Pod).
// Execute 执行动作 (删除 Pod)。
func (a *RestartPodAction) Execute(ctx context.Context, suggestion types.RemediationSuggestion) (string, error) {
	logger := log.LWithContext(ctx).With(zap.String("action", a.Name()), zap.String("suggestionID", suggestion.IssueID))
	logger.Info("Executing action: Restart Pod")

	if suggestion.ActionType != enum.ActionTypeAutomated {
		logger.Error("Action is not automated, cannot execute")
		return "", errors.New(errors.ErrorCodeActionExecutionFailed, "action is not automated", fmt.Sprintf("suggestion ID %s has type %s", suggestion.IssueID, suggestion.ActionType.String()))
	}

	// Extract necessary info from the payload
	// 从载荷中提取必要信息
	payload := suggestion.Payload
	resourceType, ok := payload["resourceType"].(string)
	if !ok || resourceType != "Pod" {
		return "", errors.New(errors.ErrorCodeInvalidInput, "invalid or missing 'resourceType' in payload", "")
	}
	resourceName, ok := payload["resourceName"].(string)
	if !ok || resourceName == "" {
		return "", errors.New(errors.ErrorCodeInvalidInput, "invalid or missing 'resourceName' in payload", "")
	}
	namespace, ok := payload["namespace"].(string)
	if !ok || namespace == "" {
		return "", errors.New(errors.ErrorCodeInvalidInput, "invalid or missing 'namespace' in payload", "")
	}
	vclusterName, _ := payload["vcluster"].(string)   // Optional, default to host if empty
	resourceUID, _ := payload["resourceUID"].(string) // Optional, use for safer deletion

	// Get the appropriate Kubernetes client (host or vcluster)
	// 获取适当的 Kubernetes 客户端 (宿主机或 vcluster)
	// This needs access to the K8sDataCollector's clients or similar client management.
	// 这需要访问 K8sDataCollector 的客户端或类似的客户端管理。
	// For this stub, let's assume a global helper function or instance is available.
	// 对于这个占位符，假设有一个全局辅助函数或实例可用。

	// Example: Get client from the initialized K8sDataCollectorInstance
	// 示例: 从已初始化的 K8sDataCollectorInstance 获取客户端
	// if k8sdatacollector.K8sCollectorInstance == nil {
	// 	return "", fmt.Errorf("kubernetes data collector instance not initialized")
	// }
	//
	// client, found := k8sdatacollector.K8sCollectorInstance.clients[vclusterName] // Accessing internal field - bad practice, needs helper
	// if !found {
	// 	// Fallback to host client if vclusterName is empty? Depends on design.
	// 	// 如果 vclusterName 为空，回退到宿主机客户端? 取决于设计。
	// 	if vclusterName == "" {
	// 		client, found = k8sdatacollector.K8sCollectorInstance.clients["host"]
	// 		if !found {
	// 			return "", fmt.Errorf("host kubernetes client not found")
	// 		}
	// 		vclusterName = "host" // Correct the name for logging
	// 	} else {
	// 		return "", fmt.Errorf("kubernetes client for vcluster '%s' not found", vclusterName)
	// 	}
	// }

	// Placeholder: Use a dummy client or mock
	// 占位符: 使用虚拟客户端或 mock
	var client kubernetes.Interface // Need to initialize this appropriately

	logger.Info("Deleting Pod to trigger restart",
		zap.String("vcluster", vclusterName),
		zap.String("namespace", namespace),
		zap.String("podName", resourceName),
		zap.String("podUID", resourceUID),
	)

	// TODO: Implement actual deletion using client-go
	// TODO: 使用 client-go 实现实际的删除操作
	// deleteOptions := metav1.DeleteOptions{}
	// if resourceUID != "" {
	// 	// Use UID for guaranteed deletion of the specific instance
	// 	// 使用 UID 确保删除特定实例
	// 	deleteOptions.Preconditions = &metav1.Preconditions{UID: &resourceUID}
	// }
	//
	// err := client.CoreV1().Pods(namespace).Delete(ctx, resourceName, deleteOptions)
	// if err != nil {
	// 	// Handle NotFound error gracefully if the pod was already gone
	// 	// 如果 Pod 已经不存在，优雅地处理 NotFound 错误
	// 	if apierrors.IsNotFound(err) {
	// 		logger.Warn("Pod not found, likely already deleted", zap.String("podName", resourceName))
	// 		return fmt.Sprintf("Pod '%s' not found, likely already deleted.", resourceName), nil
	// 	}
	// 	logger.Error("Failed to delete pod", zap.Error(err))
	// 	return "", errors.Wrap(errors.ErrorCodeActionExecutionFailed, "failed to delete pod", err, fmt.Sprintf("pod %s/%s in %s", namespace, resourceName, vclusterName))
	// }

	logger.Warn("RestartPodAction.Execute not fully implemented, simulating success")
	// Simulate success
	// 模拟成功
	simulatedResult := fmt.Sprintf("Successfully requested deletion of Pod '%s' in namespace '%s' (%s vcluster).", resourceName, namespace, vclusterName)
	return simulatedResult, nil // Placeholder success / 占位符成功
}

// Register the action with the global registry.
// 在全局注册表中注册动作。
func init() {
	// This stateful action needs Kubernetes client access before registration.
	// We will rely on the engine initialization to create and register it.
	// 这个有状态的动作在注册之前需要 Kubernetes 客户端访问。
	// 我们将依赖于引擎初始化来创建和注册它。

	// action.RegisterAction(&RestartPodAction{}) // Cannot register like this
	log.L().Debug("K8s Restart Pod action init() called, will be registered during engine setup.")
}

// Global instance placeholder, will be initialized in main.
// 全局实例占位符，将在 main 中初始化。
var RestartPodActionInstance *RestartPodAction

// RegisterRestartPodAction registers the initialized RestartPodAction instance.
// RegisterRestartPodAction 注册已初始化的 RestartPodAction 实例。
// This should be called after NewRestartPodAction is successful.
// 应在 NewRestartPodAction 成功后调用此函数。
func RegisterRestartPodAction(act *RestartPodAction) {
	action.RegisterAction(act)
	RestartPodActionInstance = act // Keep a global reference if needed elsewhere
}
