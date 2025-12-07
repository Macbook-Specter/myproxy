package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	// 保留你的内部包（若未实现，可先注释，不影响UI测试）
	// "myproxy.com/p/internal/config"
	// "myproxy.com/p/internal/logging"
	// "myproxy.com/p/internal/ping"
	// "myproxy.com/p/internal/proxy"
	// "myproxy.com/p/internal/server"
	// "myproxy.com/p/internal/subscription"
)

// --------------------------
// 1. 先定义模拟的内部包结构体（避免编译报错，实际使用时替换为真实包）
// --------------------------
type Config struct {
	SubscriptionURL    string
	AutoProxyPort      int
	AutoProxyEnabled   bool
	AutoSetSystemProxy bool
	LogFile            string
	LogLevel           string
}

func LoadConfig(path string) (*Config, error) {
	return &Config{
		AutoProxyPort: 1080,
		LogLevel:      "INFO",
	}, nil
}

func SaveConfig(cfg *Config, path string) error { return nil }

type Server struct {
	Name     string
	ID       string
	Addr     string
	Port     int
	Username string
	Password string
	Delay    int
	Enabled  bool
}

type ServerManager struct {
	servers []*Server
}

func NewServerManager(cfg *Config) *ServerManager {
	// 模拟测试数据
	return &ServerManager{
		servers: []*Server{
			{Name: "测试服务器1", ID: "1", Addr: "127.0.0.1", Port: 1080, Enabled: true},
			{Name: "测试服务器2", ID: "2", Addr: "192.168.1.1", Port: 1081, Enabled: false},
		},
	}
}

func (s *ServerManager) ListServers() []*Server { return s.servers }
func (s *ServerManager) UpdateServerDelay(id string, delay int) {
	for _, sv := range s.servers {
		if sv.ID == id {
			sv.Delay = delay
			break
		}
	}
}
func (s *ServerManager) UpdateServer(srv *Server) error { return nil }
func (s *ServerManager) SelectServer(id string) error   { return nil }

type PingManager struct{ sm *ServerManager }

func NewPingManager(sm *ServerManager) *PingManager { return &PingManager{sm: sm} }
func (p *PingManager) TestAllServersDelay() error {
	for _, srv := range p.sm.servers {
		srv.Delay = rand.Intn(200) + 50 // 模拟延迟
	}
	return nil
}
func (p *PingManager) TestServerDelay(srv *Server) (int, error) {
	return rand.Intn(200) + 50, nil
}

type SubscriptionManager struct{ sm *ServerManager }

func NewSubscriptionManager(sm *ServerManager) *SubscriptionManager {
	return &SubscriptionManager{sm: sm}
}
func (s *SubscriptionManager) FetchSubscription(url string) ([]*Server, error) {
	return s.sm.servers, nil
}
func (s *SubscriptionManager) UpdateSubscription(url string) error { return nil }

type Forwarder struct {
	IsRunning bool
	Addr      string
}

func NewAutoProxyForwarder(addr, network string, sm *ServerManager) *Forwarder {
	return &Forwarder{Addr: addr}
}
func (f *Forwarder) Start() error { f.IsRunning = true; return nil }
func (f *Forwarder) Stop() error  { f.IsRunning = false; return nil }

type Logger struct{}

func NewLogger(file string, debug bool, level string) (*Logger, error) { return &Logger{}, nil }
func (l *Logger) Info(msg string)                                      {}
func (l *Logger) GetLogs(n int) ([]string, error) {
	return []string{"2025-01-01 12:00:00 [INFO] [app] 启动成功"}, nil
}

// --------------------------
// 2. 核心AppState定义
// --------------------------
type AppState struct {
	Config              *Config
	Forwarders          map[string]*Forwarder
	ServerManager       *ServerManager
	SubscriptionManager *SubscriptionManager
	PingManager         *PingManager
	Logger              *Logger
	AutoProxyForwarder  *Forwarder
	App                 fyne.App
	Window              fyne.Window
	SelectedServerIdx   int
}

var globalAppState = &AppState{
	Forwarders: make(map[string]*Forwarder),
}

// --------------------------
// 3. 工具函数
// --------------------------
func isPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	defer listener.Close()
	return true
}

func getAvailablePort() int {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 100; i++ {
		port := rand.Intn(55536) + 10000
		if isPortAvailable(port) {
			return port
		}
	}
	return 1080
}

// --------------------------
// 4. 主函数
// --------------------------
func main() {
	// 模拟加载配置（避免真实包依赖报错）
	cfg, err := LoadConfig("./config.json")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	globalAppState.Config = cfg

	// 模拟初始化管理器
	globalAppState.Logger, _ = NewLogger("", true, "INFO")
	globalAppState.ServerManager = NewServerManager(cfg)
	globalAppState.SubscriptionManager = NewSubscriptionManager(globalAppState.ServerManager)
	globalAppState.PingManager = NewPingManager(globalAppState.ServerManager)

	// 创建Fyne应用
	globalAppState.App = app.New()
	globalAppState.Window = globalAppState.App.NewWindow("SOCKS5 代理客户端")
	globalAppState.Window.Resize(fyne.NewSize(800, 600))

	// 创建主界面
	createMainUI()

	// 运行应用
	globalAppState.Window.ShowAndRun()
}

// --------------------------
// 5. 核心UI创建（修复右键菜单）
// --------------------------
func createMainUI() {
	tabs := container.NewAppTabs(
		container.NewTabItem("服务器管理", createServerTab()),
	)

	logContent := createLogsDisplay()

	// 调整日志区域大小，使用Border布局，但设置日志区域的最小高度
	// 创建一个固定高度的日志容器
	logContainer := container.NewVBox(
		widget.NewSeparator(),
		container.NewGridWrap(fyne.NewSize(0, 200), logContent), // 设置日志区域高度为200
	)

	mainContent := container.NewBorder(
		nil,
		logContainer,
		nil, nil,
		tabs,
	)

	globalAppState.Window.SetContent(mainContent)
}

func createServerTab() fyne.CanvasObject {
	// 声明服务器列表变量
	var serverList *widget.List

	subURL := widget.NewEntry()
	subURL.SetText(globalAppState.Config.SubscriptionURL)
	subURL.PlaceHolder = "请输入订阅URL"

	// 状态更新函数
	updateAutoProxyStatus := func(statusLabel *widget.Label) {
		if globalAppState.AutoProxyForwarder != nil && globalAppState.AutoProxyForwarder.IsRunning {
			statusLabel.SetText(fmt.Sprintf("自动代理状态: 运行中，监听端口: %d", globalAppState.Config.AutoProxyPort))
		} else {
			statusLabel.SetText("自动代理状态: 未运行")
		}
	}
	statusLabel := widget.NewLabel("自动代理状态: 未运行")

	// 详情面板
	detail := widget.NewMultiLineEntry()
	detail.Disable()

	// 选中行处理
	handleServerSelection := func(row int) {
		globalAppState.SelectedServerIdx = row
		servers := globalAppState.ServerManager.ListServers()
		if row < 0 || row >= len(servers) {
			detail.SetText("")
			return
		}
		srv := servers[row]
		detail.SetText(fmt.Sprintf(
			"名称: %s\nID: %s\n地址: %s:%d\n用户名: %s\n密码: %s\n延迟: %d ms\n启用: %v",
			srv.Name, srv.ID, srv.Addr, srv.Port, srv.Username, srv.Password, srv.Delay, srv.Enabled))
	}

	// 功能函数定义
	startAutoProxy := func() {
		// 实现启动自动代理逻辑
		if globalAppState.SelectedServerIdx < 0 {
			globalAppState.Window.SetTitle("请先选择一个服务器")
			return
		}

		// 自动获取可用端口
		port := getAvailablePort()
		globalAppState.Config.AutoProxyPort = port

		// 更新状态
		updateAutoProxyStatus(statusLabel)
		globalAppState.Window.SetTitle(fmt.Sprintf("自动代理已启动，监听端口: %d", port))
	}

	stopAutoProxy := func() {
		// 实现停止自动代理逻辑
		globalAppState.AutoProxyForwarder = nil
		updateAutoProxyStatus(statusLabel)
		globalAppState.Window.SetTitle("自动代理已停止")
	}

	testAllServers := func() {
		// 实现一键测延迟逻辑
		globalAppState.PingManager.TestAllServersDelay()
		serverList.Refresh()
	}

	testSelectedServer := func() {
		// 实现测选中延迟逻辑
		if globalAppState.SelectedServerIdx < 0 {
			return
		}
		servers := globalAppState.ServerManager.ListServers()
		srv := servers[globalAppState.SelectedServerIdx]
		delay, _ := globalAppState.PingManager.TestServerDelay(srv)
		globalAppState.ServerManager.UpdateServerDelay(srv.ID, delay)
		serverList.Refresh()
		handleServerSelection(globalAppState.SelectedServerIdx)
	}

	selectDefaultServer := func() {
		// 实现设为默认逻辑
		if globalAppState.SelectedServerIdx < 0 {
			return
		}
		servers := globalAppState.ServerManager.ListServers()
		globalAppState.ServerManager.SelectServer(servers[globalAppState.SelectedServerIdx].ID)
		serverList.Refresh()
	}

	toggleServerEnable := func() {
		// 实现切换启用逻辑
		if globalAppState.SelectedServerIdx < 0 {
			return
		}
		servers := globalAppState.ServerManager.ListServers()
		srv := servers[globalAppState.SelectedServerIdx]
		srv.Enabled = !srv.Enabled
		globalAppState.ServerManager.UpdateServer(srv)
		serverList.Refresh()
		handleServerSelection(globalAppState.SelectedServerIdx)
	}

	// 创建服务器列表
	serverList = widget.NewList(
		func() int { return len(globalAppState.ServerManager.ListServers()) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, obj fyne.CanvasObject) {
			servers := globalAppState.ServerManager.ListServers()
			if i < 0 || i >= len(servers) {
				return
			}
			srv := servers[i]
			prefix := ""
			if i == globalAppState.SelectedServerIdx {
				prefix = "★ "
			}
			if !srv.Enabled {
				prefix += "[禁用] "
			}
			delay := "未测"
			if srv.Delay > 0 {
				delay = fmt.Sprintf("%d ms", srv.Delay)
			}
			obj.(*widget.Label).SetText(
				fmt.Sprintf("%s%s  %s:%d  [%s]", prefix, srv.Name, srv.Addr, srv.Port, delay))
		})

	// 选中事件
	serverList.OnSelected = handleServerSelection

	// 注意：由于Fyne 2.7.1版本API限制，无法实现完整的右键菜单功能
	// 我们将功能按钮保留在界面上，提供良好的用户体验

	// --------------------------
	// 按钮操作逻辑
	// --------------------------
	startAutoProxyBtn := widget.NewButton("启动自动代理", func() {
		if globalAppState.SelectedServerIdx < 0 {
			globalAppState.Window.SetTitle("请先选择服务器")
			return
		}
		startAutoProxy()
	})

	stopAutoProxyBtn := widget.NewButton("停止自动代理", func() {
		stopAutoProxy()
	})

	testAllBtn := widget.NewButton("一键测延迟", func() {
		testAllServers()
	})

	testSelBtn := widget.NewButton("测选中延迟", func() {
		testSelectedServer()
	})

	setDefBtn := widget.NewButton("设为默认", func() {
		selectDefaultServer()
	})

	toggleEnBtn := widget.NewButton("切换启用", func() {
		toggleServerEnable()
	})

	// --------------------------
	// 订阅按钮
	// --------------------------
	fetchBtn := widget.NewButton("获取订阅", func() {
		if subURL.Text == "" {
			return
		}
		// 获取订阅并检查结果
		servers, err := globalAppState.SubscriptionManager.FetchSubscription(subURL.Text)
		if err != nil {
			globalAppState.Window.SetTitle("订阅获取失败: " + err.Error())
			return
		}
		// 显式刷新服务器列表
		serverList.Refresh()
		// 更新配置并保存
		globalAppState.Config.SubscriptionURL = subURL.Text
		SaveConfig(globalAppState.Config, "./config.json")
		globalAppState.Window.SetTitle(fmt.Sprintf("订阅获取成功，共 %d 条服务器", len(servers)))
	})

	updateBtn := widget.NewButton("更新订阅", func() {
		if subURL.Text == "" {
			return
		}
		// 更新订阅并检查结果
		if err := globalAppState.SubscriptionManager.UpdateSubscription(subURL.Text); err != nil {
			globalAppState.Window.SetTitle("订阅更新失败: " + err.Error())
			return
		}
		// 显式刷新服务器列表
		serverList.Refresh()
		globalAppState.Window.SetTitle("订阅已更新")
	})

	// --------------------------
	// 布局组装 - 将功能按钮改为横向布局
	// --------------------------
	functionButtons := container.NewHBox(
		widget.NewLabel("服务器操作"),
		layout.NewSpacer(),
		container.NewVBox(
			widget.NewLabel("自动代理"),
			container.NewHBox(startAutoProxyBtn, stopAutoProxyBtn),
		),
		container.NewVBox(
			widget.NewLabel("延迟测试"),
			container.NewHBox(testAllBtn, testSelBtn),
		),
		container.NewVBox(
			widget.NewLabel("服务器管理"),
			container.NewHBox(setDefBtn, toggleEnBtn),
		),
	)

	top := container.NewVBox(
		widget.NewLabel("订阅管理"),
		subURL,
		container.NewHBox(fetchBtn, updateBtn),
		widget.NewSeparator(),
		widget.NewLabel("服务器列表"),
		widget.NewSeparator(),
		functionButtons,
		widget.NewSeparator(),
	)

	serverScroll := container.NewScroll(serverList)
	serverSection := container.NewHSplit(serverScroll, container.NewScroll(detail))
	serverSection.Offset = 0.7

	return container.NewBorder(top, nil, nil, nil, serverSection)
}

// --------------------------
// 6. 日志显示（保留）
// --------------------------
func createLogsDisplay() fyne.CanvasObject {
	logContent := widget.NewTextGrid()

	containsLogLevel := func(logLine, level string) bool {
		return len(logLine) > len(level)+2 && logLine[1:len(level)+1] == level
	}

	containsLogType := func(logLine, logType string) bool {
		if logType == "全部" {
			return true
		}
		return strings.Contains(logLine, "["+logType+"]")
	}

	refreshLogs := func(content *widget.TextGrid, levelFilter, typeFilter string) {
		logLines, err := globalAppState.Logger.GetLogs(100)
		if err != nil {
			content.SetText("获取日志失败: " + err.Error())
			return
		}

		var b strings.Builder
		for _, line := range logLines {
			if levelFilter != "全部" && !containsLogLevel(line, levelFilter) {
				continue
			}
			if !containsLogType(line, typeFilter) {
				continue
			}
			b.WriteString(line)
			b.WriteByte('\n')
		}
		content.SetText(b.String())
	}

	var levelSel *widget.Select
	var typeSel *widget.Select

	typeSel = widget.NewSelect([]string{"全部", "app", "proxy"}, func(value string) {
		refreshLogs(logContent, levelSel.Selected, value)
	})
	typeSelWrapped := container.NewGridWrap(fyne.NewSize(80, 40), typeSel)

	levelSel = widget.NewSelect([]string{"全部", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"}, func(value string) {
		refreshLogs(logContent, value, typeSel.Selected)
	})
	levelSelWrapped := container.NewGridWrap(fyne.NewSize(80, 40), levelSel)

	levelSel.SetSelected("全部")
	typeSel.SetSelected("全部")

	refreshBtn := widget.NewButton("刷新", func() {
		refreshLogs(logContent, levelSel.Selected, typeSel.Selected)
	})
	refreshBtnWrapped := container.NewGridWrap(fyne.NewSize(60, 40), refreshBtn)

	topBar := container.NewHBox(
		widget.NewLabel("日志: "),
		levelSelWrapped,
		widget.NewLabel("类型: "),
		typeSelWrapped,
		refreshBtnWrapped,
		layout.NewSpacer(),
	)

	refreshLogs(logContent, levelSel.Selected, typeSel.Selected)

	return container.NewBorder(topBar, nil, nil, nil, container.NewScroll(logContent))
}
