package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"myproxy.com/p/internal/config"
	"myproxy.com/p/internal/database"
	"myproxy.com/p/internal/logging"
	"myproxy.com/p/internal/ping"
	"myproxy.com/p/internal/proxy"
	"myproxy.com/p/internal/server"
	"myproxy.com/p/internal/subscription"
)

// AppState 管理应用的整体状态，包括配置、管理器、日志和 UI 组件。
// 它作为应用的核心状态容器，协调各个组件之间的交互。
type AppState struct {
	Config              *config.Config
	ServerManager       *server.ServerManager
	SubscriptionManager *subscription.SubscriptionManager
	PingManager         *ping.PingManager
	Logger              *logging.Logger
	App                 fyne.App
	Window              fyne.Window
	SelectedServerID    string

	// 代理转发器 - 用于实际启动和管理代理服务
	ProxyForwarder *proxy.Forwarder

	// 绑定数据 - 用于状态面板自动更新
	ProxyStatusBinding binding.String // 代理状态文本
	PortBinding        binding.String // 端口文本
	ServerNameBinding  binding.String // 服务器名称文本

	// 订阅标签绑定 - 用于订阅管理面板自动更新
	SubscriptionLabelsBinding binding.StringList // 订阅标签列表

	// 主窗口引用 - 用于刷新日志面板
	MainWindow *MainWindow

	// 日志面板引用 - 用于追加日志
	LogsPanel *LogsPanel
}

// NewAppState 创建并初始化新的应用状态。
// 参数：
//   - cfg: 应用配置
//   - logger: 日志记录器
//
// 返回：初始化后的应用状态实例
func NewAppState(cfg *config.Config, logger *logging.Logger) *AppState {
	serverManager := server.NewServerManager(cfg)
	subscriptionManager := subscription.NewSubscriptionManager(serverManager)
	pingManager := ping.NewPingManager(serverManager)

	// 创建绑定数据
	proxyStatusBinding := binding.NewString()
	portBinding := binding.NewString()
	serverNameBinding := binding.NewString()
	subscriptionLabelsBinding := binding.NewStringList()

	appState := &AppState{
		Config:                    cfg,
		ServerManager:             serverManager,
		SubscriptionManager:       subscriptionManager,
		PingManager:               pingManager,
		Logger:                    logger,
		SelectedServerID:          "",
		ProxyStatusBinding:        proxyStatusBinding,
		PortBinding:               portBinding,
		ServerNameBinding:         serverNameBinding,
		SubscriptionLabelsBinding: subscriptionLabelsBinding,
	}

	// 注意：不在构造函数中初始化绑定数据
	// 绑定数据需要在 Fyne 应用初始化后才能使用
	// 将在 InitApp() 之后初始化

	return appState
}

// LoadServersFromDB 将数据库中的服务器同步到内存配置，并更新选中状态。
func (a *AppState) LoadServersFromDB() {
	if a.ServerManager == nil {
		return
	}

	if err := a.ServerManager.LoadServersFromDB(); err != nil {
		if a.Logger != nil {
			a.Logger.Error("加载服务器列表失败: %v", err)
		}
		return
	}

	// 同步选中服务器ID
	a.SelectedServerID = a.Config.SelectedServerID
}

// updateStatusBindings 更新状态绑定数据
func (a *AppState) updateStatusBindings() {
	// 更新代理状态 - 基于实际运行的代理服务，而不是配置标志
	isRunning := false
	proxyPort := 0

	if a.ProxyForwarder != nil && a.ProxyForwarder.IsRunning {
		// 代理服务正在运行
		isRunning = true
		// 从本地地址中提取端口
		if a.Config != nil && a.Config.AutoProxyPort > 0 {
			proxyPort = a.Config.AutoProxyPort
		}
	}

	if isRunning {
		a.ProxyStatusBinding.Set("代理状态: 运行中")
		if proxyPort > 0 {
			a.PortBinding.Set(fmt.Sprintf("动态端口: %d", proxyPort))
		} else {
			a.PortBinding.Set("动态端口: -")
		}
	} else {
		a.ProxyStatusBinding.Set("代理状态: 未启动")
		a.PortBinding.Set("动态端口: -")
	}

	// 更新当前服务器
	if a.ServerManager != nil && a.SelectedServerID != "" {
		server, err := a.ServerManager.GetServer(a.SelectedServerID)
		if err == nil && server != nil {
			a.ServerNameBinding.Set(fmt.Sprintf("当前服务器: %s (%s:%d)", server.Name, server.Addr, server.Port))
		} else {
			a.ServerNameBinding.Set("当前服务器: 未知")
		}
	} else {
		a.ServerNameBinding.Set("当前服务器: 无")
	}
}

// UpdateProxyStatus 更新代理状态并刷新 UI 绑定数据。
// 该方法会检查代理转发器的实际运行状态，并更新相关的绑定数据，
// 使状态面板能够自动反映最新的代理状态。
func (a *AppState) UpdateProxyStatus() {
	a.updateStatusBindings()
}

// InitApp 初始化 Fyne 应用和窗口。
// 该方法会创建应用实例、设置主题、创建主窗口，并初始化数据绑定。
// 注意：必须在创建 UI 组件之前调用此方法。
func (a *AppState) InitApp() {
	a.App = app.New()
	// 从数据库加载主题配置，默认使用黑色主题
	themeVariant := theme.VariantDark
	if themeStr, err := database.GetAppConfigWithDefault("theme", "dark"); err == nil && themeStr == "light" {
		themeVariant = theme.VariantLight
	}
	a.App.Settings().SetTheme(NewMonochromeTheme(themeVariant))
	a.Window = a.App.NewWindow("SOCKS5 代理客户端")
	a.Window.Resize(fyne.NewSize(900, 700))

	// Fyne 应用初始化后，可以初始化绑定数据
	a.updateStatusBindings()
	a.updateSubscriptionLabels()
}

// updateSubscriptionLabels 更新订阅标签绑定数据
func (a *AppState) updateSubscriptionLabels() {
	// 从数据库获取所有订阅
	subscriptions, err := database.GetAllSubscriptions()
	if err != nil {
		// 如果获取失败，记录日志并设置为空列表（统一错误处理）
		if a.Logger != nil {
			a.Logger.Error("获取订阅列表失败: %v", err)
		}
		a.SubscriptionLabelsBinding.Set([]string{})
		return
	}

	// 提取标签列表
	labels := make([]string, 0, len(subscriptions))
	for _, sub := range subscriptions {
		if sub.Label != "" {
			labels = append(labels, sub.Label)
		}
	}

	// 更新绑定数据
	a.SubscriptionLabelsBinding.Set(labels)
}

// UpdateSubscriptionLabels 从数据库获取所有订阅并更新标签绑定数据。
// 该方法会触发订阅管理面板的自动更新，使 UI 能够反映最新的订阅列表。
func (a *AppState) UpdateSubscriptionLabels() {
	a.updateSubscriptionLabels()
}

// AppendLog 追加一条日志到日志面板（全局接口）
// 该方法可以从任何地方调用，会自动追加到日志缓冲区并更新显示
// 参数：
//   - level: 日志级别 (DEBUG, INFO, WARN, ERROR, FATAL)
//   - logType: 日志类型 (app, proxy)
//   - message: 日志消息
func (a *AppState) AppendLog(level, logType, message string) {
	if a.LogsPanel != nil {
		a.LogsPanel.AppendLog(level, logType, message)
	}
}
