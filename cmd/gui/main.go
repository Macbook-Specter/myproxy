package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"myproxy.com/p/internal/config"
	"myproxy.com/p/internal/database"
	"myproxy.com/p/internal/logging"
	"myproxy.com/p/internal/ui"
)

func main() {
	// 配置文件路径（用于向后兼容，实际配置存储在数据库）
	configPath := "./config.json"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	// 初始化数据库（必须在加载配置之前初始化）
	dbPath := filepath.Join(filepath.Dir(configPath), "data", "myproxy.db")
	if err := database.InitDB(dbPath); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer database.CloseDB()

	// 从数据库加载配置，如果不存在则从 JSON 文件加载并迁移
	cfg, err := loadConfigFromDB(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志（使用数据库中的配置）
	logger, err := logging.NewLogger(cfg.LogFile, cfg.LogLevel == "debug", cfg.LogLevel)
	if err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	// 创建应用状态
	appState := ui.NewAppState(cfg, logger)

	// 启动时从数据库同步服务器列表，保证 UI 有数据
	appState.LoadServersFromDB()

	// 初始化应用
	appState.InitApp()

	// 创建主窗口
	mainWindow := ui.NewMainWindow(appState)

	// 设置窗口内容
	appState.Window.SetContent(mainWindow.Build())

	// 延迟刷新，确保所有组件都已初始化
	go func() {
		// 等待窗口完全显示
		time.Sleep(100 * time.Millisecond)
		mainWindow.Refresh()
	}()

	// 设置窗口关闭事件，保存配置
	appState.Window.SetCloseIntercept(func() {
		// 保存布局配置到数据库
		mainWindow.SaveLayoutConfig()
		// 保存应用配置到数据库
		if err := saveConfigToDB(cfg); err != nil {
			log.Printf("保存配置到数据库失败: %v", err)
		}
		appState.Window.Close()
	})

	// 运行应用
	appState.Window.ShowAndRun()
}

// loadConfigFromDB 从数据库加载配置，如果不存在则从 JSON 文件加载并迁移到数据库。
func loadConfigFromDB(configPath string) (*config.Config, error) {
	// 尝试从数据库加载配置
	logLevel, _ := database.GetAppConfigWithDefault("logLevel", "")
	logFile, _ := database.GetAppConfigWithDefault("logFile", "")
	autoProxyEnabledStr, _ := database.GetAppConfigWithDefault("autoProxyEnabled", "")
	autoProxyPortStr, _ := database.GetAppConfigWithDefault("autoProxyPort", "")

	// 如果数据库中有配置，使用数据库配置
	if logLevel != "" || logFile != "" || autoProxyEnabledStr != "" || autoProxyPortStr != "" {
		cfg := config.DefaultConfig()
		if logLevel != "" {
			cfg.LogLevel = logLevel
		}
		if logFile != "" {
			cfg.LogFile = logFile
		}
		if autoProxyEnabledStr != "" {
			if enabled, err := strconv.ParseBool(autoProxyEnabledStr); err == nil {
				cfg.AutoProxyEnabled = enabled
			}
		}
		if autoProxyPortStr != "" {
			// 允许任意端口值，不进行有效性检查（用户可根据实际情况选择端口）
			if port, err := strconv.Atoi(autoProxyPortStr); err == nil {
				cfg.AutoProxyPort = port
			}
		}
		return cfg, nil
	}

	// 数据库中没有配置，从 JSON 文件加载（向后兼容）
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	// 将配置迁移到数据库
	if err := saveConfigToDB(cfg); err != nil {
		log.Printf("迁移配置到数据库失败: %v", err)
		// 即使迁移失败，也返回配置，保证应用可以启动
	}

	return cfg, nil
}

// saveConfigToDB 将配置保存到数据库。
func saveConfigToDB(cfg *config.Config) error {
	if err := database.SetAppConfig("logLevel", cfg.LogLevel); err != nil {
		return err
	}
	if err := database.SetAppConfig("logFile", cfg.LogFile); err != nil {
		return err
	}
	if err := database.SetAppConfig("autoProxyEnabled", strconv.FormatBool(cfg.AutoProxyEnabled)); err != nil {
		return err
	}
	if err := database.SetAppConfig("autoProxyPort", strconv.Itoa(cfg.AutoProxyPort)); err != nil {
		return err
	}
	return nil
}
