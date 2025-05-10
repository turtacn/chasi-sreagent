// go.mod defines the module path and dependencies for the chasi-sreagent project.
// go.mod 定义了 chasi-sreagent 项目的模块路径和依赖项。
module github.com/turtacn/chasi-sreagent

// Specify the Go version required for this module.
// 指定此模块所需的 Go 版本。
go 1.20

// Require direct dependencies.
// 引入直接依赖项。
require (
	github.com/google/uuid v1.6.0 // Used for generating unique IDs / 用于生成唯一 ID
	github.com/spf13/cobra v1.8.0 // Used for building the CLI / 用于构建 CLI
	go.uber.org/automaxprocs v1.5.3 // Automatically set GOMAXPROCS for container environments / 自动设置容器环境的 GOMAXPROCS
	go.uber.org/zap v1.27.0 // High-performance logging library / 高性能日志库
	gopkg.in/yaml.v2 v2.4.0 // Used for parsing YAML configuration files / 用于解析 YAML 配置文件

	// Kubernetes client libraries. Choose versions compatible with your target K8s cluster version.
	// For go 1.20.2, k8s.io/client-go v0.28.x or v0.29.x are common choices.
	// Kubernetes 客户端库。选择与目标 K8s 集群版本兼容的版本。
	// 对于 go 1.20.2，k8s.io/client-go v0.28.x 或 v0.29.x 是常见的选择。
	k8s.io/api v0.28.4 // Kubernetes API types / Kubernetes API 类型
	k8s.io/apimachinery v0.28.4 // Kubernetes API machinery utilities / Kubernetes API machinery 工具库
	k8s.io/client-go v0.28.4 // Kubernetes client library / Kubernetes 客户端库

	// controller-runtime is used for building the Kubernetes Operator.
	// controller-runtime 用于构建 Kubernetes Operator。
	// We require it even if the operator main is a stub, as its types might be used elsewhere.
	// 即使 operator main 是一个占位符，我们也需要引入它，因为其类型可能在其他地方使用。
	sigs.k8s.io/controller-runtime v0.16.3
)

require go.uber.org/multierr v1.11.0 // indirect

// Indicate indirect dependencies. These are dependencies required by the direct dependencies.
// Go modules automatically manage these, but 'go mod tidy' will add them explicitly.
// 标明间接依赖项。这些是直接依赖项所需的依赖项。
// Go Modules 会自动管理这些，但 'go mod tidy' 会将其显式添加。
// Example:
// require (
// 	github.com/beorn7/perks v1.0.1 // indirect
// 	github.com/cespare/xxhash/v2 v2.2.0 // indirect
// 	github.com/go-logr/logr v1.2.4 // indirect // Required by controller-runtime
// 	...
// )

// Replace directives can be used for local development or forking.
// Replace 指令可用于本地开发或分叉。
// replace example.com/some/module => ../some/module
