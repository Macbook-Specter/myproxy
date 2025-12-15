package ui

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"myproxy.com/p/internal/config"
	"myproxy.com/p/internal/database"
	"myproxy.com/p/internal/logging"
	"myproxy.com/p/internal/xray"
)

// ServerListPanel 管理服务器列表的显示和操作。
// 它支持服务器选择、延迟测试、代理启动/停止等功能，并提供右键菜单操作。
type ServerListPanel struct {
	appState           *AppState
	serverList         *widget.List
	subscriptionSelect *widget.Select // 订阅选择下拉菜单
	onServerSelect     func(server config.Server)
	statusPanel        *StatusPanel // 状态面板引用（用于刷新）
}

// NewServerListPanel 创建并初始化服务器列表面板。
// 该方法会创建服务器列表组件并设置选中事件处理。
// 参数：
//   - appState: 应用状态实例
//
// 返回：初始化后的服务器列表面板实例
func NewServerListPanel(appState *AppState) *ServerListPanel {
	slp := &ServerListPanel{
		appState: appState,
	}

	// 服务器列表
	slp.serverList = widget.NewList(
		slp.getServerCount,
		slp.createServerItem,
		slp.updateServerItem,
	)

	// 设置选中事件
	slp.serverList.OnSelected = slp.onSelected

	return slp
}

// SetOnServerSelect 设置服务器选中时的回调函数。
// 参数：
//   - callback: 当用户选中服务器时调用的回调函数
func (slp *ServerListPanel) SetOnServerSelect(callback func(server config.Server)) {
	slp.onServerSelect = callback
}

// SetStatusPanel 设置状态面板的引用，以便在服务器操作后更新状态显示。
// 参数：
//   - statusPanel: 状态面板实例
func (slp *ServerListPanel) SetStatusPanel(statusPanel *StatusPanel) {
	slp.statusPanel = statusPanel
}

// Build 构建并返回服务器列表面板的 UI 组件。
// 返回：包含操作按钮和服务器列表的容器组件
func (slp *ServerListPanel) Build() fyne.CanvasObject {
	// 操作按钮
	testAllBtn := widget.NewButton("一键测延迟", slp.onTestAll)
	startProxyBtn := widget.NewButton("启动代理", slp.onStartProxyFromSelected)
	stopProxyBtn := widget.NewButton("停止代理", slp.onStopProxy)

	// 订阅选择下拉菜单
	slp.subscriptionSelect = widget.NewSelect([]string{"加载中..."}, nil)
	slp.updateSubscriptionSelect(slp.subscriptionSelect)

	// 服务器列表标题和按钮
	headerArea := container.NewHBox(
		widget.NewLabel("服务器列表"),
		layout.NewSpacer(),
		widget.NewLabel("订阅："),
		slp.subscriptionSelect,
		testAllBtn,
		startProxyBtn,
		stopProxyBtn,
	)

	// 服务器列表滚动区域（不再展示右侧详情）
	serverScroll := container.NewScroll(slp.serverList)

	// 返回包含标题和列表的容器
	return container.NewBorder(
		headerArea,
		nil,
		nil,
		nil,
		serverScroll,
	)
}

// updateSubscriptionSelect 更新订阅选择下拉菜单
func (slp *ServerListPanel) updateSubscriptionSelect(selectWidget *widget.Select) {
	// 获取所有订阅
	subscriptions, err := database.GetAllSubscriptions()
	if err != nil {
		selectWidget.Options = []string{"全部"}
		selectWidget.Refresh()
		return
	}

	// 创建选项列表，第一个选项为"全部"
	options := []string{"全部"}
	optionToID := map[string]int64{"全部": 0}

	// 添加所有订阅
	for _, sub := range subscriptions {
		option := fmt.Sprintf("%s", sub.Label)
		options = append(options, option)
		optionToID[option] = sub.ID
	}

	// 设置选项
	selectWidget.Options = options

	// 设置当前选中项
	currentSubscriptionID := slp.appState.ServerManager.GetSelectedSubscriptionID()
	if currentSubscriptionID == 0 {
		selectWidget.SetSelected("全部")
	} else {
		for option, id := range optionToID {
			if id == currentSubscriptionID {
				selectWidget.SetSelected(option)
				break
			}
		}
	}

	// 设置选择事件处理函数
	selectWidget.OnChanged = func(selected string) {
		// 获取选中的订阅ID
		subscriptionID := optionToID[selected]

		// 设置选中的订阅
		slp.appState.ServerManager.SetSelectedSubscriptionID(subscriptionID)

		// 刷新服务器列表
		slp.Refresh()

		// 更新状态面板
		if slp.statusPanel != nil {
			slp.statusPanel.Refresh()
		}
	}

	selectWidget.Refresh()
}

// Refresh 刷新服务器列表的显示，使 UI 反映最新的服务器数据。
func (slp *ServerListPanel) Refresh() {
	fyne.Do(func() {
		slp.serverList.Refresh()
	})
}

// getServerCount 获取服务器数量
func (slp *ServerListPanel) getServerCount() int {
	return len(slp.appState.ServerManager.ListServers())
}

// createServerItem 创建服务器列表项
func (slp *ServerListPanel) createServerItem() fyne.CanvasObject {
	return NewServerListItem()
}

// updateServerItem 更新服务器列表项
func (slp *ServerListPanel) updateServerItem(id widget.ListItemID, obj fyne.CanvasObject) {
	servers := slp.appState.ServerManager.ListServers()
	if id < 0 || id >= len(servers) {
		return
	}

	srv := servers[id]
	item := obj.(*ServerListItem)

	// 设置面板引用和ID
	item.panel = slp
	item.id = id

	// 使用新的Update方法更新多列信息
	item.Update(srv)
}

// onSelected 服务器选中事件
func (slp *ServerListPanel) onSelected(id widget.ListItemID) {
	servers := slp.appState.ServerManager.ListServers()
	if id < 0 || id >= len(servers) {
		return
	}

	srv := servers[id]
	slp.appState.SelectedServerID = srv.ID

	// 更新状态绑定（使用双向绑定，UI 会自动更新）
	if slp.appState != nil {
		slp.appState.UpdateProxyStatus()
	}

	// 调用回调
	if slp.onServerSelect != nil {
		slp.onServerSelect(srv)
	}
}

// onRightClick 右键菜单
func (slp *ServerListPanel) onRightClick(id widget.ListItemID, ev *fyne.PointEvent) {
	servers := slp.appState.ServerManager.ListServers()
	if id < 0 || id >= len(servers) {
		return
	}

	srv := servers[id]
	slp.appState.SelectedServerID = srv.ID

	// 创建右键菜单
	menu := fyne.NewMenu("",
		fyne.NewMenuItem("测速", func() {
			slp.onTestSpeed(id)
		}),
		fyne.NewMenuItem("启动代理", func() {
			slp.onStartProxy(id)
		}),
		fyne.NewMenuItem("停止代理", func() {
			slp.onStopProxy()
		}),
	)

	// 显示菜单
	popup := widget.NewPopUpMenu(menu, slp.appState.Window.Canvas())
	popup.ShowAtPosition(ev.AbsolutePosition)
}

// onTestSpeed 测速
func (slp *ServerListPanel) onTestSpeed(id widget.ListItemID) {
	servers := slp.appState.ServerManager.ListServers()
	if id < 0 || id >= len(servers) {
		return
	}

	srv := servers[id]

	// 在goroutine中执行测速
	go func() {
		// 记录开始测速日志
		if slp.appState != nil {
			slp.appState.AppendLog("INFO", "ping", fmt.Sprintf("开始测试服务器延迟: %s (%s:%d)", srv.Name, srv.Addr, srv.Port))
		}

		delay, err := slp.appState.PingManager.TestServerDelay(srv)
		if err != nil {
			// 记录失败日志
			if slp.appState != nil {
				slp.appState.AppendLog("ERROR", "ping", fmt.Sprintf("服务器 %s 测速失败: %v", srv.Name, err))
			}
			fyne.Do(func() {
				slp.appState.Window.SetTitle(fmt.Sprintf("测速失败: %v", err))
			})
			return
		}

		// 更新服务器延迟
		slp.appState.ServerManager.UpdateServerDelay(srv.ID, delay)

		// 记录成功日志
		if slp.appState != nil {
			slp.appState.AppendLog("INFO", "ping", fmt.Sprintf("服务器 %s 测速完成: %d ms", srv.Name, delay))
		}

		// 更新UI（需要在主线程中执行）
		fyne.Do(func() {
			slp.Refresh()
			slp.onSelected(id) // 刷新详情
			// 更新状态绑定（使用双向绑定，UI 会自动更新）
			if slp.appState != nil {
				slp.appState.UpdateProxyStatus()
			}
			slp.appState.Window.SetTitle(fmt.Sprintf("测速完成: %d ms", delay))
		})
	}()
}

// onStartProxyFromSelected 从当前选中的服务器启动代理
func (slp *ServerListPanel) onStartProxyFromSelected() {
	if slp.appState.SelectedServerID == "" {
		slp.appState.Window.SetTitle("请先选择一个服务器")
		return
	}

	servers := slp.appState.ServerManager.ListServers()
	var srv *config.Server
	for i := range servers {
		if servers[i].ID == slp.appState.SelectedServerID {
			srv = &servers[i]
			break
		}
	}

	if srv == nil {
		slp.appState.Window.SetTitle("选中的服务器不存在")
		return
	}

	// 如果已有代理在运行，先停止
	if slp.appState.XrayInstance != nil {
		slp.appState.XrayInstance.Stop()
		slp.appState.XrayInstance = nil
	}

	// 把当前的设置为选中
	slp.appState.ServerManager.SelectServer(srv.ID)
	slp.appState.SelectedServerID = srv.ID

	// 启动代理
	slp.startProxyWithServer(srv)
}

// onStartProxy 启动代理（右键菜单使用）
func (slp *ServerListPanel) onStartProxy(id widget.ListItemID) {
	servers := slp.appState.ServerManager.ListServers()
	if id < 0 || id >= len(servers) {
		return
	}

	srv := servers[id]
	slp.appState.ServerManager.SelectServer(srv.ID)
	slp.appState.SelectedServerID = srv.ID

	// 如果已有代理在运行，先停止
	if slp.appState.XrayInstance != nil {
		slp.appState.XrayInstance.Stop()
		slp.appState.XrayInstance = nil
	}

	// 启动代理
	slp.startProxyWithServer(&srv)
}

// startProxyWithServer 使用指定的服务器启动代理
func (slp *ServerListPanel) startProxyWithServer(srv *config.Server) {
	// 使用固定的10080端口监听本地SOCKS5
	proxyPort := 10080

	// 记录开始启动日志
	if slp.appState != nil {
		slp.appState.AppendLog("INFO", "xray", fmt.Sprintf("开始启动xray-core代理: %s", srv.Name))
	}

	// 使用统一的日志文件路径（与应用日志使用同一个文件）
	unifiedLogPath := slp.appState.Logger.GetLogFilePath()

	// 创建xray配置，设置日志文件路径为统一日志文件
	xrayConfigJSON, err := xray.CreateXrayConfig(proxyPort, srv, unifiedLogPath)
	if err != nil {
		slp.logAndShowError("创建xray配置失败", err)
		slp.appState.Config.AutoProxyEnabled = false
		slp.appState.XrayInstance = nil
		slp.appState.UpdateProxyStatus()
		slp.saveConfigToDB()
		return
	}

	// 记录配置创建成功日志
	if slp.appState != nil {
		slp.appState.AppendLog("DEBUG", "xray", fmt.Sprintf("xray配置已创建: %s", srv.Name))
	}

	// 创建日志回调函数，将 xray 日志转发到应用日志系统
	logCallback := func(level, message string) {
		if slp.appState != nil {
			slp.appState.AppendLog(level, "xray", message)
		}
	}

	// 创建xray实例，并设置日志回调
	xrayInstance, err := xray.NewXrayInstanceFromJSONWithCallback(xrayConfigJSON, logCallback)
	if err != nil {
		slp.logAndShowError("创建xray实例失败", err)
		slp.appState.Config.AutoProxyEnabled = false
		slp.appState.XrayInstance = nil
		slp.appState.UpdateProxyStatus()
		slp.saveConfigToDB()
		return
	}

	// 启动xray实例
	err = xrayInstance.Start()
	if err != nil {
		slp.logAndShowError("启动xray实例失败", err)
		slp.appState.Config.AutoProxyEnabled = false
		slp.appState.XrayInstance = nil
		slp.appState.UpdateProxyStatus()
		slp.saveConfigToDB()
		return
	}

	// 启动成功，设置端口信息
	xrayInstance.SetPort(proxyPort)
	slp.appState.XrayInstance = xrayInstance
	slp.appState.Config.AutoProxyEnabled = true
	slp.appState.Config.AutoProxyPort = proxyPort

	// 记录日志（统一日志记录）
	if slp.appState.Logger != nil {
		slp.appState.Logger.InfoWithType(logging.LogTypeProxy, "xray-core代理已启动: %s (端口: %d)", srv.Name, proxyPort)
	}

	// 追加日志到日志面板
	if slp.appState != nil {
		slp.appState.AppendLog("INFO", "xray", fmt.Sprintf("xray-core代理已启动: %s (端口: %d)", srv.Name, proxyPort))
		slp.appState.AppendLog("INFO", "xray", fmt.Sprintf("服务器信息: %s:%d, 协议: %s", srv.Addr, srv.Port, srv.ProtocolType))
	}

	slp.Refresh()
	// 更新状态绑定（使用双向绑定，UI 会自动更新）
	slp.appState.UpdateProxyStatus()

	slp.appState.Window.SetTitle(fmt.Sprintf("代理已启动: %s (端口: %d)", srv.Name, proxyPort))

	// 保存配置到数据库
	slp.saveConfigToDB()
}

// logAndShowError 记录日志并显示错误对话框（统一错误处理）
func (slp *ServerListPanel) logAndShowError(message string, err error) {
	if slp.appState != nil && slp.appState.Logger != nil {
		slp.appState.Logger.Error("%s: %v", message, err)
	}
	if slp.appState != nil && slp.appState.Window != nil {
		slp.appState.Window.SetTitle(fmt.Sprintf("%s: %v", message, err))
	}
}

// saveConfigToDB 保存应用配置到数据库（统一配置保存）
func (slp *ServerListPanel) saveConfigToDB() {
	if slp.appState == nil || slp.appState.Config == nil {
		return
	}
	cfg := slp.appState.Config

	// 保存配置到数据库
	database.SetAppConfig("logLevel", cfg.LogLevel)
	database.SetAppConfig("logFile", cfg.LogFile)
	database.SetAppConfig("autoProxyEnabled", strconv.FormatBool(cfg.AutoProxyEnabled))
	database.SetAppConfig("autoProxyPort", strconv.Itoa(cfg.AutoProxyPort))
}

// onStopProxy 停止代理
func (slp *ServerListPanel) onStopProxy() {
	stopped := false

	// 停止xray实例
	if slp.appState.XrayInstance != nil {
		if slp.appState != nil {
			slp.appState.AppendLog("INFO", "xray", "正在停止xray-core代理...")
		}

		err := slp.appState.XrayInstance.Stop()
		if err != nil {
			// 停止失败，记录日志并显示错误（统一错误处理）
			slp.logAndShowError("停止xray代理失败", err)
			return
		}

		slp.appState.XrayInstance = nil
		stopped = true

		// 记录日志（统一日志记录）
		if slp.appState.Logger != nil {
			slp.appState.Logger.InfoWithType(logging.LogTypeProxy, "xray-core代理已停止")
		}

		// 追加日志到日志面板
		if slp.appState != nil {
			slp.appState.AppendLog("INFO", "xray", "xray-core代理已停止")
		}
	}

	if stopped {
		// 停止成功
		slp.appState.Config.AutoProxyEnabled = false
		slp.appState.Config.AutoProxyPort = 0

		// 更新状态绑定
		slp.appState.UpdateProxyStatus()

		// 保存配置到数据库
		slp.saveConfigToDB()

		slp.appState.Window.SetTitle("代理已停止")
	} else {
		slp.appState.Window.SetTitle("代理未运行")
	}
}

// onTestAll 一键测延迟
func (slp *ServerListPanel) onTestAll() {
	// 在goroutine中执行测速
	go func() {
		servers := slp.appState.ServerManager.ListServers()
		enabledCount := 0
		for _, s := range servers {
			if s.Enabled {
				enabledCount++
			}
		}

		// 记录开始测速日志
		if slp.appState != nil {
			slp.appState.AppendLog("INFO", "ping", fmt.Sprintf("开始一键测速，共 %d 个启用的服务器", enabledCount))
		}

		results := slp.appState.PingManager.TestAllServersDelay()

		// 统计结果并记录每个服务器的详细日志
		successCount := 0
		failCount := 0
		for _, srv := range servers {
			if !srv.Enabled {
				continue
			}
			delay, exists := results[srv.ID]
			if !exists {
				continue
			}
			if delay > 0 {
				successCount++
				if slp.appState != nil {
					slp.appState.AppendLog("INFO", "ping", fmt.Sprintf("服务器 %s (%s:%d) 测速完成: %d ms", srv.Name, srv.Addr, srv.Port, delay))
				}
			} else {
				failCount++
				if slp.appState != nil {
					slp.appState.AppendLog("ERROR", "ping", fmt.Sprintf("服务器 %s (%s:%d) 测速失败", srv.Name, srv.Addr, srv.Port))
				}
			}
		}

		// 记录完成日志
		if slp.appState != nil {
			slp.appState.AppendLog("INFO", "ping", fmt.Sprintf("一键测速完成: 成功 %d 个，失败 %d 个，共测试 %d 个服务器", successCount, failCount, len(results)))
		}

		// 更新UI（需要在主线程中执行）
		fyne.Do(func() {
			slp.Refresh()
			slp.appState.Window.SetTitle(fmt.Sprintf("测速完成，共测试 %d 个服务器", len(results)))
		})
	}()
}

// getSelectedIndex 获取当前选中的索引
func (slp *ServerListPanel) getSelectedIndex() widget.ListItemID {
	servers := slp.appState.ServerManager.ListServers()
	for i, srv := range servers {
		if srv.ID == slp.appState.SelectedServerID {
			return widget.ListItemID(i)
		}
	}
	return -1
}

// ServerListItem 自定义服务器列表项（支持右键菜单和多列显示）
type ServerListItem struct {
	widget.BaseWidget
	id          widget.ListItemID
	panel       *ServerListPanel
	container   *fyne.Container
	nameLabel   *widget.Label
	addrLabel   *widget.Label
	portLabel   *widget.Label
	userLabel   *widget.Label
	delayLabel  *widget.Label
	statusLabel *widget.Label
}

// NewServerListItem 创建新的服务器列表项
func NewServerListItem() *ServerListItem {
	// 创建各列标签
	nameLabel := widget.NewLabel("")
	nameLabel.Wrapping = fyne.TextTruncate
	nameLabel.TextStyle = fyne.TextStyle{Bold: true}

	addrLabel := widget.NewLabel("")
	addrLabel.Wrapping = fyne.TextTruncate

	portLabel := widget.NewLabel("")
	portLabel.Alignment = fyne.TextAlignCenter

	userLabel := widget.NewLabel("")
	userLabel.Wrapping = fyne.TextTruncate

	delayLabel := widget.NewLabel("")
	delayLabel.Alignment = fyne.TextAlignCenter

	statusLabel := widget.NewLabel("")
	statusLabel.Alignment = fyne.TextAlignCenter

	// 创建容器，使用网格布局，确保所有列都能显示
	// 为每列添加一个包含标签的固定大小容器
	nameContainer := container.NewMax(nameLabel)
	nameContainer.Resize(fyne.NewSize(180, 40))

	addrContainer := container.NewMax(addrLabel)
	addrContainer.Resize(fyne.NewSize(120, 40))

	portContainer := container.NewMax(portLabel)
	portContainer.Resize(fyne.NewSize(60, 40))

	userContainer := container.NewMax(userLabel)
	userContainer.Resize(fyne.NewSize(100, 40))

	delayContainer := container.NewMax(delayLabel)
	delayContainer.Resize(fyne.NewSize(80, 40))

	statusContainer := container.NewMax(statusLabel)
	statusContainer.Resize(fyne.NewSize(80, 40))

	// 使用网格布局组织各列容器
	container := container.NewGridWithColumns(6,
		nameContainer,
		addrContainer,
		portContainer,
		userContainer,
		delayContainer,
		statusContainer,
	)

	item := &ServerListItem{
		container:   container,
		nameLabel:   nameLabel,
		addrLabel:   addrLabel,
		portLabel:   portLabel,
		userLabel:   userLabel,
		delayLabel:  delayLabel,
		statusLabel: statusLabel,
	}
	item.ExtendBaseWidget(item)
	return item
}

// CreateRenderer 创建渲染器
func (s *ServerListItem) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(s.container)
}

// TappedSecondary 处理右键点击事件
func (s *ServerListItem) TappedSecondary(pe *fyne.PointEvent) {
	if s.panel == nil {
		return
	}
	s.panel.onRightClick(s.id, pe)
}

// Update  更新服务器列表项的信息
func (s *ServerListItem) Update(server config.Server) {
	fyne.Do(func() {
		// 服务器名称（带选中标记）
		prefix := ""
		if server.Selected {
			prefix = "★ "
		}
		if !server.Enabled {
			prefix += "[禁用] "
		}
		s.nameLabel.SetText(prefix + server.Name)

		// 服务器地址
		s.addrLabel.SetText(server.Addr)

		// 端口
		s.portLabel.SetText(strconv.Itoa(server.Port))

		// 用户名
		username := server.Username
		if username == "" {
			username = "无"
		}
		s.userLabel.SetText(username)

		// 延迟
		delayText := "未测"
		if server.Delay > 0 {
			delayText = fmt.Sprintf("%d ms", server.Delay)
		} else if server.Delay < 0 {
			delayText = "失败"
		}
		s.delayLabel.SetText(delayText)

		// 状态
		status := "启用"
		if !server.Enabled {
			status = "禁用"
		}
		s.statusLabel.SetText(status)
	})
}
