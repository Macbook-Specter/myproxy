package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"myproxy.com/p/internal/config"
	"myproxy.com/p/internal/proxy"
)

// ServerListPanel 服务器列表面板
type ServerListPanel struct {
	appState       *AppState
	serverList     *widget.List
	onServerSelect func(server config.Server)
	statusPanel    *StatusPanel // 状态面板引用（用于刷新）
}

// NewServerListPanel 创建服务器列表面板
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

// SetOnServerSelect 设置服务器选中回调
func (slp *ServerListPanel) SetOnServerSelect(callback func(server config.Server)) {
	slp.onServerSelect = callback
}

// SetStatusPanel 设置状态面板引用
func (slp *ServerListPanel) SetStatusPanel(statusPanel *StatusPanel) {
	slp.statusPanel = statusPanel
}

// Build 构建服务器列表面板 UI
func (slp *ServerListPanel) Build() fyne.CanvasObject {
	// 操作按钮
	testAllBtn := widget.NewButton("一键测延迟", slp.onTestAll)
	setDefaultBtn := widget.NewButton("设为默认", slp.onSetDefault)
	toggleEnableBtn := widget.NewButton("切换启用", slp.onToggleEnable)

	// 服务器列表标题和按钮
	headerArea := container.NewHBox(
		widget.NewLabel("服务器列表"),
		layout.NewSpacer(),
		testAllBtn,
		setDefaultBtn,
		toggleEnableBtn,
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

// Refresh 刷新服务器列表
func (slp *ServerListPanel) Refresh() {
	slp.serverList.Refresh()
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

	// 构建显示文本
	prefix := ""
	if srv.Selected {
		prefix = "★ "
	}
	if !srv.Enabled {
		prefix += "[禁用] "
	}

	delay := "未测"
	if srv.Delay > 0 {
		delay = fmt.Sprintf("%d ms", srv.Delay)
	} else if srv.Delay < 0 {
		delay = "失败"
	}

	text := fmt.Sprintf("%s%s  %s:%d  [%s]", prefix, srv.Name, srv.Addr, srv.Port, delay)
	item.SetText(text)
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
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("设为默认", func() {
			slp.onSetDefaultServer(id)
		}),
		fyne.NewMenuItem("切换启用", func() {
			slp.onToggleEnableServer(id)
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
	delay, err := slp.appState.PingManager.TestServerDelay(srv)
	if err != nil {
		slp.appState.Window.SetTitle(fmt.Sprintf("测速失败: %v", err))
		return
	}

	slp.appState.ServerManager.UpdateServerDelay(srv.ID, delay)
	slp.Refresh()
	slp.onSelected(id) // 刷新详情
	// 更新状态绑定（使用双向绑定，UI 会自动更新）
	if slp.appState != nil {
		slp.appState.UpdateProxyStatus()
	}
	slp.appState.Window.SetTitle(fmt.Sprintf("测速完成: %d ms", delay))
}

// onStartProxy 启动代理
func (slp *ServerListPanel) onStartProxy(id widget.ListItemID) {
	servers := slp.appState.ServerManager.ListServers()
	if id < 0 || id >= len(servers) {
		return
	}

	srv := servers[id]
	slp.appState.ServerManager.SelectServer(srv.ID)
	slp.appState.SelectedServerID = srv.ID

	// 如果已有代理在运行，先停止
	if slp.appState.ProxyForwarder != nil && slp.appState.ProxyForwarder.IsRunning {
		slp.appState.ProxyForwarder.Stop()
	}

	// 如果端口为0，自动分配一个可用端口
	if slp.appState.Config.AutoProxyPort == 0 {
		slp.appState.Config.AutoProxyPort = 1080 // 默认端口
	}

	// 创建并启动自动代理转发器
	localAddr := fmt.Sprintf("127.0.0.1:%d", slp.appState.Config.AutoProxyPort)
	forwarder := proxy.NewAutoProxyForwarder(localAddr, "tcp", slp.appState.ServerManager)

	// 实际启动代理服务
	err := forwarder.Start()
	if err != nil {
		// 启动失败，记录日志并显示错误
		if slp.appState.Logger != nil {
			slp.appState.Logger.Error("启动代理失败: %v", err)
		}
		slp.appState.Config.AutoProxyEnabled = false
		slp.appState.ProxyForwarder = nil
		slp.appState.UpdateProxyStatus()
		slp.appState.Window.SetTitle(fmt.Sprintf("代理启动失败: %v", err))
		return
	}

	// 启动成功
	slp.appState.ProxyForwarder = forwarder
	slp.appState.Config.AutoProxyEnabled = true

	// 记录日志
	if slp.appState.Logger != nil {
		slp.appState.Logger.Info("代理已启动: %s (端口: %d)", srv.Name, slp.appState.Config.AutoProxyPort)
	}

	slp.Refresh()
	// 更新状态绑定（使用双向绑定，UI 会自动更新）
	slp.appState.UpdateProxyStatus()

	// 延迟刷新日志面板，确保日志已写入
	go func() {
		time.Sleep(200 * time.Millisecond)
		if slp.appState != nil && slp.appState.MainWindow != nil {
			slp.appState.MainWindow.Refresh()
		}
	}()

	slp.appState.Window.SetTitle(fmt.Sprintf("代理已启动: %s (端口: %d)", srv.Name, slp.appState.Config.AutoProxyPort))
}

// onStopProxy 停止代理
func (slp *ServerListPanel) onStopProxy() {
	// 如果代理正在运行，停止它
	if slp.appState.ProxyForwarder != nil && slp.appState.ProxyForwarder.IsRunning {
		err := slp.appState.ProxyForwarder.Stop()
		if err != nil {
			// 停止失败，记录日志
			if slp.appState.Logger != nil {
				slp.appState.Logger.Error("停止代理失败: %v", err)
			}
			slp.appState.Window.SetTitle(fmt.Sprintf("停止代理失败: %v", err))
			return
		}

		// 停止成功
		slp.appState.Config.AutoProxyEnabled = false
		slp.appState.Config.AutoProxyPort = 0

		// 记录日志
		if slp.appState.Logger != nil {
			slp.appState.Logger.Info("代理已停止")
		}

		// 更新状态绑定
		slp.appState.UpdateProxyStatus()

		// 延迟刷新日志面板，确保日志已写入
		go func() {
			time.Sleep(200 * time.Millisecond)
			if slp.appState != nil && slp.appState.MainWindow != nil {
				slp.appState.MainWindow.Refresh()
			}
		}()

		slp.appState.Window.SetTitle("代理已停止")
	} else {
		slp.appState.Window.SetTitle("代理未运行")
	}
}

// onTestAll 一键测延迟
func (slp *ServerListPanel) onTestAll() {
	results := slp.appState.PingManager.TestAllServersDelay()
	slp.Refresh()
	slp.appState.Window.SetTitle(fmt.Sprintf("测速完成，共测试 %d 个服务器", len(results)))
}

// onSetDefault 设为默认
func (slp *ServerListPanel) onSetDefault() {
	if slp.appState.SelectedServerID == "" {
		slp.appState.Window.SetTitle("请先选择一个服务器")
		return
	}

	err := slp.appState.ServerManager.SelectServer(slp.appState.SelectedServerID)
	if err != nil {
		slp.appState.Window.SetTitle("设置失败: " + err.Error())
		return
	}

	slp.Refresh()
	// 更新状态绑定（使用双向绑定，UI 会自动更新）
	if slp.appState != nil {
		slp.appState.UpdateProxyStatus()
	}
	slp.appState.Window.SetTitle("已设为默认服务器")
}

// onSetDefaultServer 设为默认（通过ID）
func (slp *ServerListPanel) onSetDefaultServer(id widget.ListItemID) {
	servers := slp.appState.ServerManager.ListServers()
	if id < 0 || id >= len(servers) {
		return
	}

	srv := servers[id]
	err := slp.appState.ServerManager.SelectServer(srv.ID)
	if err != nil {
		slp.appState.Window.SetTitle("设置失败: " + err.Error())
		return
	}

	slp.Refresh()
	// 更新状态绑定（使用双向绑定，UI 会自动更新）
	if slp.appState != nil {
		slp.appState.UpdateProxyStatus()
	}
	slp.appState.Window.SetTitle("已设为默认服务器")
	// 通知主窗口刷新状态面板
	if slp.appState != nil && slp.appState.Window != nil {
		// 可以通过事件机制通知，这里简化处理
	}
}

// onToggleEnable 切换启用
func (slp *ServerListPanel) onToggleEnable() {
	if slp.appState.SelectedServerID == "" {
		slp.appState.Window.SetTitle("请先选择一个服务器")
		return
	}

	srv, err := slp.appState.ServerManager.GetServer(slp.appState.SelectedServerID)
	if err != nil {
		slp.appState.Window.SetTitle("获取服务器失败: " + err.Error())
		return
	}

	// 切换启用状态
	newSrv := *srv
	newSrv.Enabled = !newSrv.Enabled
	err = slp.appState.ServerManager.UpdateServer(newSrv)
	if err != nil {
		slp.appState.Window.SetTitle("更新失败: " + err.Error())
		return
	}

	slp.Refresh()
	slp.onSelected(slp.getSelectedIndex())
}

// onToggleEnableServer 切换启用（通过ID）
func (slp *ServerListPanel) onToggleEnableServer(id widget.ListItemID) {
	servers := slp.appState.ServerManager.ListServers()
	if id < 0 || id >= len(servers) {
		return
	}

	srv := servers[id]
	newSrv := srv
	newSrv.Enabled = !newSrv.Enabled
	err := slp.appState.ServerManager.UpdateServer(newSrv)
	if err != nil {
		slp.appState.Window.SetTitle("更新失败: " + err.Error())
		return
	}

	slp.Refresh()
	slp.onSelected(id)
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

// ServerListItem 自定义服务器列表项（支持右键菜单）
type ServerListItem struct {
	widget.BaseWidget
	label *widget.Label
	id    widget.ListItemID
	panel *ServerListPanel
}

// NewServerListItem 创建新的服务器列表项
func NewServerListItem() *ServerListItem {
	item := &ServerListItem{
		label: widget.NewLabel(""),
	}
	item.ExtendBaseWidget(item)
	return item
}

// CreateRenderer 创建渲染器
func (s *ServerListItem) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(s.label)
}

// TappedSecondary 处理右键点击事件
func (s *ServerListItem) TappedSecondary(pe *fyne.PointEvent) {
	if s.panel == nil {
		return
	}
	s.panel.onRightClick(s.id, pe)
}

// SetText 设置文本
func (s *ServerListItem) SetText(text string) {
	s.label.SetText(text)
}
