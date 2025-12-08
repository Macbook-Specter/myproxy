package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// StatusPanel 状态信息面板（显示代理状态和端口信息）
// 使用 Fyne 双向绑定，自动同步数据更新
type StatusPanel struct {
	appState         *AppState
	proxyStatusLabel *widget.Label
	portLabel        *widget.Label
	serverNameLabel  *widget.Label
}

// NewStatusPanel 创建状态信息面板
func NewStatusPanel(appState *AppState) *StatusPanel {
	sp := &StatusPanel{
		appState: appState,
	}

	// 使用绑定数据创建标签，实现自动更新
	// 代理状态标签 - 绑定到 ProxyStatusBinding
	sp.proxyStatusLabel = widget.NewLabelWithData(appState.ProxyStatusBinding)
	sp.proxyStatusLabel.Wrapping = fyne.TextWrapOff

	// 端口标签 - 绑定到 PortBinding
	sp.portLabel = widget.NewLabelWithData(appState.PortBinding)
	sp.portLabel.Wrapping = fyne.TextWrapOff

	// 服务器名称标签 - 绑定到 ServerNameBinding
	sp.serverNameLabel = widget.NewLabelWithData(appState.ServerNameBinding)
	sp.serverNameLabel.Wrapping = fyne.TextWrapOff

	return sp
}

// Build 构建状态面板 UI
func (sp *StatusPanel) Build() fyne.CanvasObject {
	// 使用水平布局显示所有信息，所有元素横向排列
	// 使用 HBox 布局，元素从左到右排列，保持最小尺寸
	statusArea := container.NewHBox(
		sp.proxyStatusLabel,
		widget.NewSeparator(), // 分隔符
		sp.portLabel,
		widget.NewSeparator(), // 分隔符
		sp.serverNameLabel,
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

	// 确保容器可见
	result.Show()

	return result
}

// Refresh 刷新状态信息（现在通过绑定自动更新，此方法保留用于兼容性）
func (sp *StatusPanel) Refresh() {
	// 使用双向绑定后，只需要更新绑定数据，UI 会自动更新
	if sp.appState != nil {
		sp.appState.UpdateProxyStatus()
	}
}
