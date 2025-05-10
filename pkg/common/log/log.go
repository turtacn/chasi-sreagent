package log

import (
	"context"
	"os"
	"sync"

	"github.com/turtacn/chasi-sreagent/pkg/common/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Package log provides a centralized logging setup for the chasi-sreagent project.
// 包 log 为 chasi-sreagent 项目提供集中的日志设置。

var (
	// logger is the package-level logger instance.
	// logger 是包级别的 logger 实例。
	logger *zap.Logger
	// mu is a mutex to protect logger initialization.
	// mu 是用于保护 logger 初始化的互斥锁。
	mu sync.Mutex
)

// Init initializes the global logger based on the provided configuration.
// Init 根据提供的配置初始化全局 logger。
// It should be called once at the beginning of the program.
// 它应该在程序启动时调用一次。
func Init(cfg *types.LogConfig) error {
	mu.Lock()
	defer mu.Unlock()

	if logger != nil {
		// Logger already initialized
		// Logger 已经初始化
		return nil
	}

	// Default config if cfg is nil
	// 如果 cfg 为 nil，使用默认配置
	if cfg == nil {
		cfg = &types.LogConfig{
			Level:  "info",
			Format: "console",
			Output: "stdout",
		}
	}

	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "fatal":
		level = zapcore.FatalLevel
	case "panic":
		level = zapcore.PanicLevel
	default:
		level = zapcore.InfoLevel // Default to info level
		// Note: Can log a warning here if default is used, but logger might not be fully ready.
		// 注意: 如果使用了默认值，可以在这里记录一个警告，但 logger 可能尚未完全准备好。
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	if cfg.Format == "console" {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder        // Consistent time format / 一致的时间格式
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Add color for console / console 添加颜色

	var output zapcore.WriteSyncer
	switch cfg.Output {
	case "stdout":
		output = zapcore.Lock(os.Stdout)
	case "stderr":
		output = zapcore.Lock(os.Stderr)
	default:
		// Assume it's a file path
		// 假定是一个文件路径
		file, err := os.OpenFile(cfg.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			// Fallback to stderr if file opening fails
			// 如果文件打开失败，回退到 stderr
			// Use a temporary basic logger to report this
			// 使用临时基础 logger 报告此情况
			tempLogger, _ := zap.NewDevelopment(zap.AddCaller())
			tempLogger.Error("Failed to open log file, falling back to stderr", zap.String("filepath", cfg.Output), zap.Error(err))
			output = zapcore.Lock(os.Stderr)
		} else {
			output = zapcore.Lock(file)
		}
	}

	core := zapcore.NewCore(
		zapcore.NewEncoder(encoderConfig),
		output,
		level,
	)

	// Add caller skip to hide logging internal frames
	// 添加 caller skip 以隐藏 logging 内部调用栈帧
	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	// Replace global logger for zap if needed by other libraries using zap
	// 如果其他使用 zap 的库需要，替换 zap 的全局 logger
	// zap.ReplaceGlobals(logger)

	return nil
}

// L returns the package-level logger instance.
// L 返回包级别的 logger 实例。
// It should be called after Init has been successfully called.
// 应在 Init 成功调用后调用。
func L() *zap.Logger {
	mu.Lock()
	defer mu.Unlock()
	if logger == nil {
		// Fallback to a no-op logger or panic if Init was not called.
		// 如果 Init 未被调用，回退到无操作 logger 或 panic。
		// For robustness, returning a no-op logger is better.
		// 为了健壮性，返回一个无操作 logger 更好。
		// In production, failure to initialize logger should probably be fatal earlier.
		// 在生产环境中，初始化 logger 失败通常应该在更早的阶段就致命退出。
		// Let's return a development logger with a warning for now in development.
		// 在开发环境中，我们暂时返回一个带警告的开发 logger。
		tempLogger, _ := zap.NewDevelopment(zap.AddCaller())
		tempLogger.Warn("Log.Init was not called! Using a temporary logger.")
		return tempLogger
	}
	return logger
}

// LWithContext returns a logger with fields extracted from context, like traceID or vcluster name.
// LWithContext 返回一个包含从 context 中提取的字段 (如 traceID 或 vcluster 名称) 的 logger。
func LWithContext(ctx context.Context) *zap.Logger {
	l := L() // Get the base logger / 获取基础 logger
	fields := []zap.Field{}

	if traceID, ok := ctx.Value(types.ContextKeyTraceID).(string); ok && traceID != "" {
		fields = append(fields, zap.String("traceID", traceID))
	}
	if vclusterName, ok := ctx.Value(types.ContextKeyVClusterName).(string); ok && vclusterName != "" {
		fields = append(fields, zap.String("vcluster", vclusterName))
	}
	// Add other relevant context fields here
	// 在这里添加其他相关的 context 字段

	if len(fields) > 0 {
		return l.With(fields...)
	}
	return l
}
