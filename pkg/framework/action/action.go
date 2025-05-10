package action

import (
	"context"
	"fmt"
	"github.com/uber-go/zap"
	"sync"

	"github.com/turtacn/chasi-sreagent/pkg/common/errors"
	"github.com/turtacn/chasi-sreagent/pkg/common/log"
	"github.com/turtacn/chasi-sreagent/pkg/common/types"
)

// Package action defines the interface for remediation actions and a registry.
// 包 action 定义了处置动作的接口和一个注册表。

// Action is the interface that all remediation action components must implement.
// Action 是所有处置动作组件必须实现的接口。
// An action represents a possible step to resolve an issue.
// 动作代表解决问题的可能步骤。
type Action interface {
	// Name returns the unique name of the action.
	// Name 返回动作的唯一名称。
	Name() string

	// Description returns a brief description of what the action does.
	// Description 返回动作功能的简要描述。
	Description() string

	// Type returns the type of action (Suggestion or Automated).
	// Type 返回动作的类型 (建议 或 自动化)。
	Type() enum.ActionType

	// Plan determines if this action is relevant for the given diagnosis result
	// Plan 根据给定的诊断结果确定此动作是否相关，
	// and prepares the parameters for execution.
	// 并准备执行参数。
	// It returns a boolean indicating relevance and the prepared payload/command.
	// 它返回一个布尔值表示相关性以及准备好的载荷/命令。
	Plan(ctx context.Context, diagnosis *types.DiagnosisResult) (bool, types.RemediationSuggestion, error)

	// Execute performs the action if its type is Automated.
	// Execute 如果动作类型是自动化，则执行该动作。
	// This method should be implemented with extreme caution and proper checks.
	// 此方法应极其谨慎地实现，并进行适当检查。
	// It returns the result of the execution or an error.
	// 它返回执行结果或一个错误。
	Execute(ctx context.Context, suggestion types.RemediationSuggestion) (string, error) // Result description or error
}

// ActionRegistry is a global registry for managing Action implementations.
// ActionRegistry 是一个用于管理 Action 实现的全局注册表。
type ActionRegistry struct {
	actions map[string]Action
	mu      sync.RWMutex
}

// Global registry instance.
// 全局注册表实例。
var globalActionRegistry = &ActionRegistry{
	actions: make(map[string]Action),
}

// RegisterAction registers an Action with the global registry.
// RegisterAction 在全局注册表中注册一个 Action。
// It panics if an action with the same name is already registered.
// 如果同名的动作已被注册，则会 panic。
func RegisterAction(action Action) {
	globalActionRegistry.mu.Lock()
	defer globalActionRegistry.mu.Unlock()

	name := action.Name()
	if _, exists := globalActionRegistry.actions[name]; exists {
		panic(fmt.Sprintf("action with name '%s' already registered", name))
	}
	globalActionRegistry.actions[name] = action
	log.L().Info("Registered action", zap.String("name", name), zap.String("type", action.Type().String()))
}

// GetAction retrieves an Action from the global registry by name.
// GetAction 按名称从全局注册表中检索一个 Action。
// It returns the Action and true if found, otherwise nil and false.
// 如果找到，返回 Action 和 true，否则返回 nil 和 false。
func GetAction(name string) (Action, bool) {
	globalActionRegistry.mu.RLock()
	defer globalActionRegistry.mu.RUnlock()

	action, found := globalActionRegistry.actions[name]
	return action, found
}

// ListActions returns a list of all registered Action names.
// ListActions 返回所有已注册的动作名称列表。
func ListActions() []string {
	globalActionRegistry.mu.RLock()
	defer globalActionRegistry.mu.RUnlock()

	names := make([]string, 0, len(globalActionRegistry.actions))
	for name := range globalActionRegistry.actions {
		names = append(names, name)
	}
	return names
}
