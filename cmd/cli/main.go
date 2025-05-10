package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"                           // Using cobra for CLI
	"github.com/turtacn/chasi-sreagent/pkg/common/log" // Using common logging
	"go.uber.org/zap"
)

// CLI main entry point.
// CLI 主入口点。
func main() {
	// Initialize logging for CLI commands
	// 初始化 CLI 命令的日志记录
	// CLI might need slightly different logging config (e.g., console output always)
	// CLI 可能需要稍微不同的日志配置 (例如, 始终输出到控制台)
	log.Init(nil) // Initialize with default config for now
	logger := log.L().With(zap.String("component", "cli"))

	// Create the root command
	// 创建根命令
	var rootCmd = &cobra.Command{
		Use:   "chasi-sreagent-cli",
		Short: "chasi-sreagent-cli is a command-line tool for interacting with the SRE agent",
		Long: `A command-line interface for triggering analysis, viewing results,
and managing the chasi-sreagent AI SRE agent.`,
	}

	// Add subcommands
	// 添加子命令
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(statusCmd)
	// TODO: Add more commands: diagnose, suggest, execute, list-analyzers, list-actions, config, etc.
	// TODO: 添加更多命令: diagnose, suggest, execute, list-analyzers, list-actions, config 等。

	// Execute the root command
	// 执行根命令
	if err := rootCmd.Execute(); err != nil {
		logger.Error("CLI command failed", zap.Error(err))
		os.Exit(1)
	}
}

// analyzeCmd represents the analyze command
// analyzeCmd 表示 analyze 命令
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Trigger an analysis run",
	Long:  `Triggers a one-time analysis run by the SRE agent.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement analysis trigger logic
		// This might involve:
		// - Loading config (similar to agent)
		// - Initializing minimal dependencies required to interact with agent/API
		// - Calling the agent's API or function to trigger analysis
		// TODO: 实现分析触发逻辑
		// 这可能涉及:
		// - 加载配置 (类似于 agent)
		// - 初始化与 agent/API 交互所需的最小依赖项
		// - 调用 agent 的 API 或函数触发分析

		log.L().Info("Triggering analysis run (placeholder)")
		fmt.Println("Analysis trigger functionality not yet implemented.")
	},
}

// statusCmd represents the status command
// statusCmd 表示 status 命令
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get the status of the SRE agent",
	Long:  `Retrieves the current status and recent activity of the SRE agent.`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement status retrieval logic
		// This might involve:
		// - Loading config
		// - Calling the agent's API to get status
		// TODO: 实现状态检索逻辑
		// 这可能涉及:
		// - 加载配置
		// - 调用 agent 的 API 获取状态

		log.L().Info("Getting agent status (placeholder)")
		fmt.Println("Agent status retrieval functionality not yet implemented.")
	},
}
