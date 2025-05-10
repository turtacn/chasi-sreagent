package k8s

import (
	"context"
	"fmt"

	"github.com/turtacn/chasi-sreagent/pkg/common/enum"
	"github.com/turtacn/chasi-sreagent/pkg/common/log"
	"github.com/turtacn/chasi-sreagent/pkg/common/types"
	"github.com/turtacn/chasi-sreagent/pkg/framework/datacollector"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Package k8s provides a data collector for Kubernetes resources.
// 包 k8s 提供一个用于收集 Kubernetes 资源的数据采集器。

// K8sDataCollector collects data from Kubernetes API, supporting multiple vclusters.
// K8sDataCollector 从 Kubernetes API 收集数据，支持多个 vcluster。
type K8sDataCollector struct {
	// clients maps vcluster name (or "host" for host cluster) to Kubernetes clients.
	// clients 将 vcluster 名称 (或 "host" 表示宿主机集群) 映射到 Kubernetes 客户端。
	clients map[string]kubernetes.Interface
	config  *types.KubernetesConfig
}

// Ensure K8sDataCollector implements the datacollector.DataCollector interface.
// 确保 K8sDataCollector 实现了 datacollector.DataCollector 接口。
var _ datacollector.DataCollector = &K8sDataCollector{}

// NewK8sDataCollector creates a new K8sDataCollector instance.
// NewK8sDataCollector 创建一个新的 K8sDataCollector 实例。
// It initializes Kubernetes clients for the host cluster and configured vclusters.
// 它为宿主机集群和配置的 vcluster 初始化 Kubernetes 客户端。
func NewK8sDataCollector(cfg *types.KubernetesConfig) (*K8sDataCollector, error) {
	clients := make(map[string]kubernetes.Interface)

	// Get host cluster config and client
	// 获取宿主机集群配置和客户端
	hostConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: cfg.KubeconfigPath},
		&clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		// If kubeconfigPath is empty and not in-cluster, this might fail.
		// If in-cluster, rest.InClusterConfig() is the alternative.
		// 如果 kubeconfigPath 为空且不在集群内，这可能会失败。
		// 如果在集群内，rest.InClusterConfig() 是替代方案。
		inClusterConfig, inClusterErr := rest.InClusterConfig()
		if inClusterErr != nil {
			log.L().Error("Failed to get host cluster config", zap.Error(err), zap.Error(inClusterErr))
			return nil, fmt.Errorf("failed to get host cluster config: %w", err)
		}
		hostConfig = inClusterConfig
		log.L().Info("Using in-cluster configuration for host cluster")
	} else {
		log.L().Info("Using kubeconfig file for host cluster", zap.String("path", cfg.KubeconfigPath))
	}

	hostClient, err := kubernetes.NewForConfig(hostConfig)
	if err != nil {
		log.L().Error("Failed to create host cluster client", zap.Error(err))
		return nil, fmt.Errorf("failed to create host cluster client: %w", err)
	}
	clients["host"] = hostClient
	log.L().Info("Initialized client for host cluster")

	// Get clients for vclusters
	// 获取 vcluster 的客户端
	for _, vcfg := range cfg.Vclusters {
		var vclusterConfig *rest.Config
		var vclusterErr error

		if vcfg.Kubeconfig != "" {
			// Use inline kubeconfig
			// 使用内联 kubeconfig
			loadingRules := &clientcmd.ClientConfigLoadingRules{}
			loadingRules.LoadFromFile = "" // Prevent loading from default file paths / 防止从默认文件路径加载
			apiConfig, parseErr := clientcmd.Load([]byte(vcfg.Kubeconfig))
			if parseErr != nil {
				log.L().Error("Failed to parse inline vcluster kubeconfig", zap.String("vcluster", vcfg.Name), zap.Error(parseErr))
				continue // Skip this vcluster / 跳过此 vcluster
			}
			vclusterConfig, vclusterErr = clientcmd.NewDefaultClientConfig(*apiConfig, &clientcmd.ConfigOverrides{}).ClientConfig()

		} else if vcfg.Context != "" {
			// Use context from main kubeconfig
			// 使用主 kubeconfig 中的上下文
			vclusterConfig, vclusterErr = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
				&clientcmd.ClientConfigLoadingRules{ExplicitPath: cfg.KubeconfigPath},
				&clientcmd.ConfigOverrides{CurrentContext: vcfg.Context}).ClientConfig()
		} else {
			log.L().Warn("Vcluster config missing kubeconfig or context", zap.String("vcluster", vcfg.Name))
			continue // Skip this vcluster / 跳过此 vcluster
		}

		if vclusterErr != nil {
			log.L().Error("Failed to get vcluster config", zap.String("vcluster", vcfg.Name), zap.Error(vclusterErr))
			continue // Skip this vcluster / 跳过此 vcluster
		}

		vclusterClient, err := kubernetes.NewForConfig(vclusterConfig)
		if err != nil {
			log.L().Error("Failed to create vcluster client", zap.String("vcluster", vcfg.Name), zap.Error(err))
			continue // Skip this vcluster / 跳过此 vcluster
		}
		clients[vcfg.Name] = vclusterClient
		log.L().Info("Initialized client for vcluster", zap.String("name", vcfg.Name))
	}

	if len(clients) == 0 {
		return nil, fmt.Errorf("no Kubernetes clients could be initialized")
	}

	return &K8sDataCollector{
		clients: clients,
		config:  cfg,
	}, nil
}

// Name returns the name of the data collector.
// Name 返回数据采集器的名称。
func (c *K8sDataCollector) Name() string {
	return "kubernetes-collector"
}

// Description returns a brief description of the collector.
// Description 返回采集器的简要描述。
func (c *K8sDataCollector) Description() string {
	return "Collects data from Kubernetes API in host cluster and vclusters."
}

// Type returns the data source type.
// Type 返回数据源类型。
func (c *K8sDataCollector) Type() enum.DataSourceType {
	return enum.DataSourceTypeKubernetesAPI
}

// Collect gathers data from the Kubernetes API.
// Collect 从 Kubernetes API 收集数据。
// Options should include "resourceType" (e.g., "Pod", "Node", "Event") and optionally "vcluster" name, "namespace", "name", etc.
// Options 应包含 "resourceType" (例如, "Pod", "Node", "Event") 并可选包含 "vcluster" 名称, "namespace", "name" 等。
// Returns a slice of Kubernetes objects (e.g., []corev1.Pod, []corev1.Node) or an error.
// 返回 Kubernetes 对象切片 (例如, []corev1.Pod, []corev1.Node) 或一个错误。
func (c *K8sDataCollector) Collect(ctx context.Context, options map[string]interface{}) (interface{}, error) {
	logger := log.LWithContext(ctx).With(zap.String("collector", c.Name()))
	logger.Debug("Collecting Kubernetes data", zap.Any("options", options))

	resourceType, ok := options["resourceType"].(string)
	if !ok || resourceType == "" {
		return nil, errors.New(errors.ErrorCodeInvalidInput, "missing or invalid 'resourceType' in options", "")
	}

	vclusterName, _ := options["vcluster"].(string) // Optional: specific vcluster / 可选: 特定 vcluster
	namespace, _ := options["namespace"].(string)   // Optional: specific namespace / 可选: 特定命名空间
	name, _ := options["name"].(string)             // Optional: specific resource name / 可选: 特定资源名称

	clientsToUse := c.clients // Default to all clients / 默认使用所有客户端
	if vclusterName != "" {
		client, found := c.clients[vclusterName]
		if !found {
			return nil, errors.New(errors.ErrorCodeNotFound, "vcluster client not found", fmt.Sprintf("client for vcluster '%s' not found", vclusterName))
		}
		clientsToUse = map[string]kubernetes.Interface{vclusterName: client}
	}

	var allCollectedData []interface{}

	for clusterName, client := range clientsToUse {
		clusterLogger := logger.With(zap.String("cluster", clusterName), zap.String("resourceType", resourceType))
		clusterLogger.Debug("Collecting from cluster")

		var collectedData interface{}
		var err error

		// --- Collection Logic based on ResourceType ---
		// Need to implement collection logic for each supported resource type.
		// Requires calling the appropriate client-go methods.
		// 需要为每个支持的资源类型实现收集逻辑。
		// 需要调用相应的 client-go 方法。
		switch resourceType {
		case "Pod":
			podList, listErr := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
			if listErr != nil {
				err = fmt.Errorf("failed to list pods: %w", listErr)
			} else {
				// Augment Pods with vcluster name before returning
				// 在返回之前使用 vcluster 名称增强 Pod 信息
				podsWithVCluster := make([]corev1.Pod, len(podList.Items))
				for i := range podList.Items {
					podsWithVCluster[i] = podList.Items[i]
					// Add vcluster info to labels or annotations if needed by analyzers
					// 如果分析器需要，将 vcluster 信息添加到标签或注解中
					if podsWithVCluster[i].Labels == nil {
						podsWithVCluster[i].Labels = make(map[string]string)
					}
					podsWithVCluster[i].Labels["chasi.turtacn.com/vcluster"] = clusterName
				}
				collectedData = podsWithVCluster
			}
		case "Node":
			// Nodes are typically only in the host cluster, unless vcluster has node awareness
			// 节点通常只存在于宿主机集群，除非 vcluster 具有节点感知能力
			if clusterName == "host" {
				nodeList, listErr := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
				if listErr != nil {
					err = fmt.Errorf("failed to list nodes: %w", listErr)
				} else {
					collectedData = nodeList.Items // Assuming analysts want the list directly
				}
			} else {
				clusterLogger.Debug("Skipping Node collection in vcluster")
			}
		case "Event":
			eventList, listErr := client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
			if listErr != nil {
				err = fmt.Errorf("failed to list events: %w", listErr)
			} else {
				// Augment Events with vcluster name
				// 使用 vcluster 名称增强事件信息
				eventsWithVCluster := make([]corev1.Event, len(eventList.Items))
				for i := range eventList.Items {
					eventsWithVCluster[i] = eventList.Items[i]
					if eventsWithVCluster[i].Labels == nil {
						eventsWithVCluster[i].Labels = make(map[string]string)
					}
					eventsWithVCluster[i].Labels["chasi.turtacn.com/vcluster"] = clusterName
				}
				collectedData = eventsWithVCluster
			}
		// TODO: Add more resource types like Deployment, StatefulSet, DaemonSet, Service, Ingress, etc.
		// TODO: 添加更多资源类型，例如 Deployment, StatefulSet, DaemonSet, Service, Ingress 等。
		default:
			err = errors.New(errors.ErrorCodeInvalidInput, "unsupported resource type", fmt.Sprintf("resource type '%s' is not supported by collector", resourceType))
		}

		if err != nil {
			clusterLogger.Error("Failed to collect data for resource type", zap.Error(err))
			// Decide if failure for one resource type/vcluster is fatal or just log and continue
			// 决定一个资源类型/vcluster 的失败是致命的还是只记录并继续
			// For now, log error and continue
			// 目前，记录错误并继续
			continue
		}

		// Append collected data from this cluster.
		// Note: This assumes all collected data for a resource type can be combined into a single slice of interface{}.
		// A more robust approach might return a map[string]interface{} where keys are cluster names.
		// 附加从此集群收集到的数据。
		// 注意: 这假设收集到的所有资源类型数据可以合并到 interface{} 的单个切片中。
		// 更健壮的方法可能返回一个 map[string]interface{}，其中键是集群名称。
		if collectedData != nil {
			// Need to handle different underlying types based on resourceType
			// 需要根据 resourceType 处理不同的底层类型
			// Example: append []corev1.Pod, []corev1.Event, etc.
			// 示例: 附加 []corev1.Pod, []corev1.Event 等。
			// For simplicity of the stub, let's just add it as a single item in a generic slice for now.
			// For real implementation, reflect or type assertions are needed.
			// 为了占位符的简单性，我们暂时将其作为通用切片中的单个项目添加。
			// 对于实际实现，需要使用 reflect 或类型断言。
			allCollectedData = append(allCollectedData, collectedData)
		}
	}

	// Returning a generic slice of interface{} requires downstream consumers (Analyzers)
	// to know the expected types and perform type assertions.
	// 返回 interface{} 的通用切片要求下游消费者 (分析器) 知道期望的类型并执行类型断言。
	// A better design might be to have type-specific Collect methods (e.g., CollectPods)
	// or return a more structured result.
	// 更好的设计可能是具有类型特定的 Collect 方法 (例如, CollectPods)
	// 或者返回更结构化的结果。
	// For now, we return the slice of interface{} slices.
	// 目前，我们返回 interface{} 切片组成的切片。
	// Example: result might be []interface{}{[]corev1.Pod, []corev1.Event}
	// 示例: 结果可能是 []interface{}{[]corev1.Pod, []corev1.Event}

	logger.Debug("Kubernetes data collection finished", zap.Int("numClusters", len(clientsToUse)), zap.String("resourceType", resourceType), zap.Int("itemsCollected", len(allCollectedData)))

	// Return the collected data. The structure depends on how Analyzers expect it.
	// Returning a map keyed by resource type or cluster name might be better.
	// 返回收集到的数据。结构取决于分析器如何期望它。
	// 返回按资源类型或集群名称为键的映射可能更好。
	// Let's return a map[string]interface{} where key is resource type for now.
	// 现在，我们返回一个 map[string]interface{}，其中键是资源类型。
	// This assumes we only collected one resource type in this call.
	// 这假设我们在此调用中只收集了一种资源类型。
	// If options allowed collecting multiple types, the structure needs re-thinking.
	// 如果 options 允许收集多种类型，则结构需要重新思考。
	// Assuming options requests data for ONE resource type across clusters:
	// 假设 options 请求跨集群的一种资源类型的数据:
	// We need to combine collected data from all clusters for that resource type.
	// 我们需要合并所有集群中该资源类型收集到的数据。
	// This combining logic depends on the resource type.
	// 这种合并逻辑取决于资源类型。
	// Example: For Pods, combine all []corev1.Pod slices.
	// 示例: 对于 Pod，合并所有 []corev1.Pod 切片。

	combinedResult, err := combineCollectedK8sData(resourceType, allCollectedData) // TODO: Implement combineCollectedK8sData
	if err != nil {
		logger.Error("Failed to combine collected data", zap.String("resourceType", resourceType), zap.Error(err))
		return nil, fmt.Errorf("failed to combine collected data for type %s: %w", resourceType, err)
	}

	return combinedResult, nil
}

// combineCollectedK8sData is a helper to combine data slices from multiple clusters for a given resource type.
// combineCollectedK8sData 是一个辅助函数，用于合并给定资源类型来自多个集群的数据切片。
func combineCollectedK8sData(resourceType string, data []interface{}) (interface{}, error) {
	if len(data) == 0 {
		// Return an empty slice of the expected type if possible
		// 如果可能，返回期望类型的空切片
		switch resourceType {
		case "Pod":
			return []corev1.Pod{}, nil
		case "Node":
			return []corev1.Node{}, nil
		case "Event":
			return []corev1.Event{}, nil
		// Add other types
		default:
			return nil, fmt.Errorf("unsupported resource type for combining: %s", resourceType)
		}
	}

	// This requires type assertion and appending based on resourceType
	// 这需要根据 resourceType 进行类型断言和附加
	// Example for Pods:
	// var allPods []corev1.Pod
	// for _, item := range data {
	// 	pods, ok := item.([]corev1.Pod)
	// 	if !ok {
	// 		return nil, fmt.Errorf("unexpected data type during combining for Pods: %T", item)
	// 	}
	// 	allPods = append(allPods, pods...)
	// }
	// return allPods, nil

	// Placeholder implementation: just return the first item for now
	// 占位符实现: 暂时只返回第一个项目
	log.L().Warn("combineCollectedK8sData is a placeholder, returning first item only", zap.String("resourceType", resourceType), zap.Int("numItems", len(data)))
	return data[0], nil // DANGER: This is not a proper implementation! / 危险: 这不是一个正确的实现!
}

// Register the data collector with the global registry.
// 在全局注册表中注册数据采集器。
func init() {
	// Note: The collector needs the Config to initialize clients.
	// Registration alone isn't enough. The engine or main function
	// should create and initialize the collector with config, then register it.
	// This global init() is typically for stateless registries or factories.
	// For stateful components like this collector, the main setup logic is better.
	// Let's keep the init() here for now to mark its place, but the actual
	// instantiation with config will happen in the engine setup.
	// 注意: 采集器需要 Config 来初始化客户端。
	// 单独的注册是不够的。引擎或 main 函数应该
	// 使用配置创建并初始化采集器，然后注册它。
	// 这个全局 init() 通常用于无状态注册表或工厂。
	// 对于像这个采集器这样的有状态组件，主设置逻辑更好。
	// 我们暂时保留 init() 在这里标记其位置，但实际的
	// 使用配置进行实例化将在引擎设置中进行。

	// datacollector.RegisterDataCollector(&K8sDataCollector{}) // Cannot register like this as it needs config
	// We will rely on the engine initialization to create and register the collector.
	// 我们将依赖于引擎初始化来创建和注册采集器。
	log.L().Debug("Kubernetes data collector init() called, will be registered during engine setup.")
}

// Global instance placeholder, will be initialized in main.
// 全局实例占位符，将在 main 中初始化。
var K8sCollectorInstance *K8sDataCollector

// RegisterK8sCollector registers the initialized K8sDataCollector instance.
// RegisterK8sCollector 注册已初始化的 K8sDataCollector 实例。
// This should be called after NewK8sDataCollector is successful.
// 应在 NewK8sDataCollector 成功后调用此函数。
func RegisterK8sCollector(collector *K8sDataCollector) {
	datacollector.RegisterDataCollector(collector)
	K8sCollectorInstance = collector // Keep a global reference if needed elsewhere
}
