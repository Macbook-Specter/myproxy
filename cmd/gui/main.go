package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"myproxy.com/p/internal/config"
	"myproxy.com/p/internal/database"
	"myproxy.com/p/internal/logging"
	"myproxy.com/p/internal/ui"
)

func main() {
	// 配置文件路径
	configPath := "./config.json"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	// 加载配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 初始化数据库
	dbPath := filepath.Join(filepath.Dir(configPath), "data", "myproxy.db")
	if err := database.InitDB(dbPath); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer database.CloseDB()

	// 初始化日志
	logger, err := logging.NewLogger(cfg.LogFile, cfg.LogLevel == "debug", cfg.LogLevel)
	if err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	// 创建应用状态
	appState := ui.NewAppState(cfg, logger)

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
		// 保存配置文件
		if err := config.SaveConfig(cfg, configPath); err != nil {
			log.Printf("保存配置失败: %v", err)
		}
		appState.Window.Close()
	})

	// 运行应用
	appState.Window.ShowAndRun()
}
