package errors

import (
	"fmt"
)

// Package errors defines custom error types and error handling utilities for chasi-sreagent.
// 包 errors 定义了 chasi-sreagent 的自定义错误类型和错误处理工具。

// AgentError represents a custom error type for chasi-sreagent operations.
// AgentError 表示 chasi-sreagent 操作的自定义错误类型。
type AgentError struct {
	Code    ErrorCode // A structured error code / 结构化错误码
	Message string    // Human-readable error message / 人类可读的错误信息
	Details string    // More detailed information about the error / 关于错误的更详细信息
	Wrapped error     // The original error, if wrapping another / 包装的原始错误 (如果存在)
}

// ErrorCode represents a specific type of error.
// ErrorCode 表示特定类型的错误。
type ErrorCode string

const (
	// ErrorCodeUnknown indicates an unknown error.
	// ErrorCodeUnknown 表示未知错误。
	ErrorCodeUnknown ErrorCode = "UNKNOWN"
	// ErrorCodeConfigLoadingFailed indicates failure to load configuration.
	// ErrorCodeConfigLoadingFailed 表示配置加载失败。
	ErrorCodeConfigLoadingFailed ErrorCode = "CONFIG_LOADING_FAILED"
	// ErrorCodeKubernetesConnectionFailed indicates failure to connect to Kubernetes.
	// ErrorCodeKubernetesConnectionFailed 表示连接 Kubernetes 失败。
	ErrorCodeKubernetesConnectionFailed ErrorCode = "KUBERNETES_CONNECTION_FAILED"
	// ErrorCodeAnalyzerFailed indicates an analyzer encountered an error.
	// ErrorCodeAnalyzerFailed 表示分析器遇到错误。
	ErrorCodeAnalyzerFailed ErrorCode = "ANALYZER_FAILED"
	// ErrorCodeLLMProviderError indicates an error from an LLM provider.
	// ErrorCodeLLMProviderError 表示来自 LLM 提供商的错误。
	ErrorCodeLLMProviderError ErrorCode = "LLM_PROVIDER_ERROR"
	// ErrorCodeKnowledgeBaseError indicates an error interacting with the knowledge base.
	// ErrorCodeKnowledgeBaseError 表示与知识库交互错误。
	ErrorCodeKnowledgeBaseError ErrorCode = "KNOWLEDGE_BASE_ERROR"
	// ErrorCodeBusinessSDKError indicates an error calling a business SDK.
	// ErrorCodeBusinessSDKError 表示调用业务 SDK 错误。
	ErrorCodeBusinessSDKError ErrorCode = "BUSINESS_SDK_ERROR"
	// ErrorCodeActionExecutionFailed indicates failure to execute an action.
	// ErrorCodeActionExecutionFailed 表示执行动作失败。
	ErrorCodeActionExecutionFailed ErrorCode = "ACTION_EXECUTION_FAILED"
	// ErrorCodeInvalidInput indicates invalid input parameters.
	// ErrorCodeInvalidInput 表示输入参数无效。
	ErrorCodeInvalidInput ErrorCode = "INVALID_INPUT"
	// ErrorCodeNotFound indicates a requested resource was not found.
	// ErrorCodeNotFound 表示请求的资源未找到。
	ErrorCodeNotFound ErrorCode = "NOT_FOUND"
	// ErrorCodePermissionDenied indicates insufficient permissions.
	// ErrorCodePermissionDenied 表示权限不足。
	ErrorCodePermissionDenied ErrorCode = "PERMISSION_DENIED"
)

// Error implements the error interface for AgentError.
// Error 为 AgentError 实现 error 接口。
func (e *AgentError) Error() string {
	if e.Wrapped != nil {
		return fmt.Sprintf("[%s] %s: %s (details: %s, wrapped: %v)", e.Code, e.Message, e.Details, e.Wrapped)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
}

// Wrap wraps an existing error with an AgentError.
// Wrap 使用 AgentError 包装现有错误。
func Wrap(code ErrorCode, message string, err error, details string) *AgentError {
	return &AgentError{
		Code:    code,
		Message: message,
		Details: details,
		Wrapped: err,
	}
}

// New creates a new AgentError without wrapping an existing error.
// New 创建一个新的 AgentError，不包装现有错误。
func New(code ErrorCode, message string, details string) *AgentError {
	return &AgentError{
		Code:    code,
		Message: message,
		Details: details,
		Wrapped: nil,
	}
}

// IsErrorCode checks if an error is an AgentError with the specific code.
// IsErrorCode 检查错误是否为具有特定代码的 AgentError。
func IsErrorCode(err error, code ErrorCode) bool {
	if ae, ok := err.(*AgentError); ok {
		return ae.Code == code
	}
	// Optionally, check wrapped errors recursively
	// (Optional) 可以递归检查包装的错误
	// if ae, ok := err.(*AgentError); ok && ae.Wrapped != nil {
	// 	return IsErrorCode(ae.Wrapped, code)
	// }
	return false
}

// GetErrorCode attempts to extract the ErrorCode from an error.
// GetErrorCode 尝试从错误中提取 ErrorCode。
func GetErrorCode(err error) ErrorCode {
	if ae, ok := err.(*AgentError); ok {
		return ae.Code
	}
	// Optionally, check wrapped errors recursively
	// (Optional) 可以递归检查包装的错误
	// if ae, ok := err.(*AgentError); ok && ae.Wrapped != nil {
	// 	return GetErrorCode(ae.Wrapped)
	// }
	return ErrorCodeUnknown
}
