package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"myproxy.com/p/internal/database"
	"myproxy.com/p/internal/logging"
	"myproxy.com/p/internal/systemproxy"
)

// 系统代理模式常量定义
const (
	// 完整模式名称
	SystemProxyModeClear      = "清除系统代理"
	SystemProxyModeAuto       = "自动配置系统代理"
	SystemProxyModeTerminal   = "环境变量代理"

	// 简短模式名称（用于UI显示）
	SystemProxyModeShortClear    = "清除"
	SystemProxyModeShortAuto     = "系统"
	SystemProxyModeShortTerminal = "终端"
)

// StatusPanel 显示代理状态、端口和当前服务器信息。
// 它使用 Fyne 的双向数据绑定机制，当应用状态更新时自动刷新显示。
type StatusPanel struct {
	appState         *AppState
	proxyStatusLabel *widget.Label
	portLabel        *widget.Label
	serverNameLabel  *widget.Label
	proxyModeSelect  *widget.Select
	systemProxy      *systemproxy.SystemProxy
}

// NewStatusPanel 创建并初始化状态信息面板。
// 该方法会创建绑定到应用状态的标签组件，实现自动更新。
// 参数：
//   - appState: 应用状态实例
//
// 返回：初始化后的状态面板实例
func NewStatusPanel(appState *AppState) *StatusPanel {
	sp := &StatusPanel{
		appState: appState,
	}

	// 检查绑定数据是否已初始化
	if appState == nil {
		// 如果 appState 为 nil，创建默认标签（不应该发生，但作为安全措施）
		sp.proxyStatusLabel = widget.NewLabel("代理状态: 未知")
		sp.portLabel = widget.NewLabel("动态端口: -")
		sp.serverNameLabel = widget.NewLabel("当前服务器: 无")
		return sp
	}

	// 使用绑定数据创建标签，实现自动更新
	// 代理状态标签 - 绑定到 ProxyStatusBinding
	if appState.ProxyStatusBinding != nil {
		sp.proxyStatusLabel = widget.NewLabelWithData(appState.ProxyStatusBinding)
	} else {
		sp.proxyStatusLabel = widget.NewLabel("代理状态: 未知")
	}
	sp.proxyStatusLabel.Wrapping = fyne.TextWrapOff

	// 端口标签 - 绑定到 PortBinding
	if appState.PortBinding != nil {
		sp.portLabel = widget.NewLabelWithData(appState.PortBinding)
	} else {
		sp.portLabel = widget.NewLabel("动态端口: -")
	}
	sp.portLabel.Wrapping = fyne.TextWrapOff

	// 服务器名称标签 - 绑定到 ServerNameBinding
	if appState.ServerNameBinding != nil {
		sp.serverNameLabel = widget.NewLabelWithData(appState.ServerNameBinding)
	} else {
		sp.serverNameLabel = widget.NewLabel("当前服务器: 无")
	}
	sp.serverNameLabel.Wrapping = fyne.TextWrapOff

	// 创建系统代理管理器（默认使用 localhost:10080）
	sp.systemProxy = systemproxy.NewSystemProxy("127.0.0.1", 10080)

	// 创建系统代理设置下拉框（只读，用于显示当前状态，不绑定 change 事件）
	// 选项使用简短文本显示，但在内部映射到完整功能
	sp.proxyModeSelect = widget.NewSelect(
		[]string{
			SystemProxyModeShortClear,
			SystemProxyModeShortAuto,
			SystemProxyModeShortTerminal,
		},
		nil, // 不绑定 change 事件，只在启动时恢复状态
	)
	sp.proxyModeSelect.PlaceHolder = "系统代理设置"

	// 恢复系统代理状态（在应用启动时）
	sp.restoreSystemProxyState()

	return sp
}

// Build 构建并返回状态信息面板的 UI 组件。
// 返回：包含代理状态、端口和服务器名称的水平布局容器
func (sp *StatusPanel) Build() fyne.CanvasObject {
	// 使用水平布局显示所有信息，所有元素横向排列
	// 使用 HBox 布局，元素从左到右排列，保持最小尺寸
	// 添加 Spacer 将下拉框推到最右边
	statusArea := container.NewHBox(
		sp.proxyStatusLabel,
		widget.NewSeparator(), // 分隔符
		sp.portLabel,
		widget.NewSeparator(), // 分隔符
		sp.serverNameLabel,
		layout.NewSpacer(),    // Spacer，将下拉框推到最右边
		widget.NewSeparator(), // 分隔符
		sp.proxyModeSelect,    // 系统代理设置下拉框（最右边）
	)

	// 使用 Border 布局，顶部添加分隔线，确保区域可见
	// Border 布局：top=分隔线，center=状态信息内容（水平布局）
	result := container.NewBorder(
		widget.NewSeparator(), // 顶部：分隔线
		nil,                   // 底部：无
		nil,                   // 左侧：无
		nil,                   // 右侧：无
		statusArea,            // 中间：状态信息内容（HBox 水平布局）
	)

	return result
}

// Refresh 刷新状态信息显示。
// 注意：由于使用了双向数据绑定，通常只需要更新绑定数据即可，UI 会自动更新。
// 此方法保留用于兼容性，实际更新通过 AppState.UpdateProxyStatus() 完成。
func (sp *StatusPanel) Refresh() {
	// 使用双向绑定后，只需要更新绑定数据，UI 会自动更新
	if sp.appState != nil {
		sp.appState.UpdateProxyStatus()
	}
	// 更新系统代理管理器的端口
	sp.updateSystemProxyPort()
}

// updateSystemProxyPort 更新系统代理管理器的端口
func (sp *StatusPanel) updateSystemProxyPort() {
	if sp.appState == nil || sp.appState.Config == nil {
		return
	}

	// 从配置或 xray 实例获取端口
	proxyPort := 10080 // 默认端口
	if sp.appState.XrayInstance != nil && sp.appState.XrayInstance.IsRunning() {
		if port := sp.appState.XrayInstance.GetPort(); port > 0 {
			proxyPort = port
		}
	} else if sp.appState.Config.AutoProxyPort > 0 {
		proxyPort = sp.appState.Config.AutoProxyPort
	}

	// 更新系统代理管理器
	sp.systemProxy = systemproxy.NewSystemProxy("127.0.0.1", proxyPort)
}

// getFullModeName 将简短文本映射到完整的功能名称
func (sp *StatusPanel) getFullModeName(shortText string) string {
	switch shortText {
	case SystemProxyModeShortClear:
		return SystemProxyModeClear
	case SystemProxyModeShortAuto:
		return SystemProxyModeAuto
	case SystemProxyModeShortTerminal:
		return SystemProxyModeTerminal
	default:
		return shortText
	}
}

// getShortModeName 将完整的功能名称映射到简短文本
func (sp *StatusPanel) getShortModeName(fullName string) string {
	switch fullName {
	case SystemProxyModeClear:
		return SystemProxyModeShortClear
	case SystemProxyModeAuto:
		return SystemProxyModeShortAuto
	case SystemProxyModeTerminal:
		return SystemProxyModeShortTerminal
	default:
		return ""
	}
}

// applySystemProxyMode 应用系统代理模式（在启动时调用）
func (sp *StatusPanel) applySystemProxyMode(fullModeName string) error {
	if sp.appState == nil {
		return fmt.Errorf("appState 未初始化")
	}

	// 更新系统代理端口
	sp.updateSystemProxyPort()

	var err error
	var logMessage string

	switch fullModeName {
	case SystemProxyModeClear:
		// 清除系统代理
		err = sp.systemProxy.ClearSystemProxy()
		// 同时清除环境变量代理，避免污染环境
		terminalErr := sp.systemProxy.ClearTerminalProxy()
		if err == nil && terminalErr == nil {
			logMessage = "已清除系统代理设置和环境变量代理"
		} else if err != nil && terminalErr != nil {
			logMessage = fmt.Sprintf("清除系统代理失败: %v; 清除环境变量代理失败: %v", err, terminalErr)
			err = fmt.Errorf("清除失败: %v; %v", err, terminalErr)
		} else if err != nil {
			logMessage = fmt.Sprintf("清除系统代理失败: %v; 已清除环境变量代理", err)
		} else {
			logMessage = fmt.Sprintf("已清除系统代理设置; 清除环境变量代理失败: %v", terminalErr)
			err = terminalErr
		}

	case SystemProxyModeAuto:
		// 先清除之前的代理设置，再设置新的
		_ = sp.systemProxy.ClearSystemProxy()
		_ = sp.systemProxy.ClearTerminalProxy()
		// 然后设置系统代理
		err = sp.systemProxy.SetSystemProxy()
		if err == nil {
			proxyPort := 10080
			if sp.appState.XrayInstance != nil && sp.appState.XrayInstance.IsRunning() {
				if port := sp.appState.XrayInstance.GetPort(); port > 0 {
					proxyPort = port
				}
			} else if sp.appState.Config != nil && sp.appState.Config.AutoProxyPort > 0 {
				proxyPort = sp.appState.Config.AutoProxyPort
			}
			logMessage = fmt.Sprintf("已自动配置系统代理: 127.0.0.1:%d", proxyPort)
		} else {
			logMessage = fmt.Sprintf("自动配置系统代理失败: %v", err)
		}

	case SystemProxyModeTerminal:
		// 先清除之前的代理设置
		_ = sp.systemProxy.ClearSystemProxy()
		_ = sp.systemProxy.ClearTerminalProxy()
		// 然后设置环境变量代理
		err = sp.systemProxy.SetTerminalProxy()
		if err == nil {
			proxyPort := 10080
			if sp.appState.XrayInstance != nil && sp.appState.XrayInstance.IsRunning() {
				if port := sp.appState.XrayInstance.GetPort(); port > 0 {
					proxyPort = port
				}
			} else if sp.appState.Config != nil && sp.appState.Config.AutoProxyPort > 0 {
				proxyPort = sp.appState.Config.AutoProxyPort
			}
			logMessage = fmt.Sprintf("已设置环境变量代理: socks5://127.0.0.1:%d (已写入shell配置文件)", proxyPort)
		} else {
			logMessage = fmt.Sprintf("设置环境变量代理失败: %v", err)
		}
	}

	// 输出日志到日志区域
	if err == nil {
		sp.appState.AppendLog("INFO", "app", logMessage)
		if sp.appState.Logger != nil {
			sp.appState.Logger.InfoWithType(logging.LogTypeApp, logMessage)
		}
	} else {
		sp.appState.AppendLog("ERROR", "app", logMessage)
		if sp.appState.Logger != nil {
			sp.appState.Logger.Error(logMessage)
		}
	}

	return err
}

// saveSystemProxyState 保存系统代理状态到数据库
func (sp *StatusPanel) saveSystemProxyState(mode string) {
	if err := database.SetAppConfig("systemProxyMode", mode); err != nil {
		if sp.appState != nil && sp.appState.Logger != nil {
			sp.appState.Logger.Error("保存系统代理状态失败: %v", err)
		}
	}
}

// restoreSystemProxyState 从数据库恢复系统代理状态（在应用启动时调用）
func (sp *StatusPanel) restoreSystemProxyState() {
	// 从数据库读取保存的系统代理模式
	mode, err := database.GetAppConfig("systemProxyMode")
	if err != nil || mode == "" {
		// 如果没有保存的状态，不执行任何操作
		return
	}

	// 应用系统代理模式
	restoreErr := sp.applySystemProxyMode(mode)

	// 更新下拉框显示文本（使用简短文本）
	if restoreErr == nil {
		shortText := sp.getShortModeName(mode)
		if shortText != "" {
			sp.proxyModeSelect.SetSelected(shortText)
		}
	}
}
