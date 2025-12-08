package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// LogsPanel 日志显示面板
type LogsPanel struct {
	appState   *AppState
	logContent *widget.Entry // 使用 Entry 替代 TextGrid，更好地支持中文
	levelSel   *widget.Select
	typeSel    *widget.Select
}

// NewLogsPanel 创建日志显示面板
func NewLogsPanel(appState *AppState) *LogsPanel {
	lp := &LogsPanel{
		appState: appState,
	}

	// 日志内容 - 使用 Entry 设置为多行只读模式，更好地支持中文
	lp.logContent = widget.NewMultiLineEntry()
	lp.logContent.Wrapping = fyne.TextWrapOff // 关闭自动换行，使用水平滚动
	lp.logContent.MultiLine = true
	lp.logContent.Disable() // 设置为只读

	// 设置日志文本样式，使用等宽字体，更适合日志显示
	lp.logContent.TextStyle = fyne.TextStyle{
		Monospace: true, // 使用等宽字体，更适合日志显示
	}

	// 设置背景色为深色，文本颜色会自动调整为浅色（通过主题）
	// 这样可以提高日志的可读性
	// 注意：Entry 的背景色可以通过自定义容器来实现

	// 日志级别选择
	lp.levelSel = widget.NewSelect(
		[]string{"全部", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"},
		func(value string) {
			if lp.typeSel != nil { // 确保 typeSel 已初始化
				lp.refreshLogs()
			}
		},
	)

	// 日志类型选择
	lp.typeSel = widget.NewSelect(
		[]string{"全部", "app", "proxy"},
		func(value string) {
			if lp.levelSel != nil { // 确保 levelSel 已初始化
				lp.refreshLogs()
			}
		},
	)

	// 等所有组件创建完成后再设置默认值和刷新
	lp.levelSel.SetSelected("全部")
	lp.typeSel.SetSelected("全部")

	// 初始刷新
	lp.refreshLogs()

	return lp
}

// Build 构建日志面板 UI
func (lp *LogsPanel) Build() fyne.CanvasObject {
	// 顶部控制栏
	topBar := container.NewHBox(
		widget.NewLabel("日志: "),
		container.NewGridWrap(fyne.NewSize(80, 40), lp.levelSel),
		widget.NewLabel("类型: "),
		container.NewGridWrap(fyne.NewSize(80, 40), lp.typeSel),
		widget.NewButton("刷新", func() {
			lp.refreshLogs()
		}),
		layout.NewSpacer(),
	)

	// 日志内容区域
	logScroll := container.NewScroll(lp.logContent)

	return container.NewBorder(
		topBar,
		nil,
		nil,
		nil,
		logScroll,
	)
}

// refreshLogs 刷新日志
func (lp *LogsPanel) refreshLogs() {
	// 检查组件是否已初始化
	if lp.appState == nil || lp.logContent == nil || lp.levelSel == nil || lp.typeSel == nil {
		return
	}

	if lp.appState.Logger == nil {
		lp.logContent.Enable()
		lp.logContent.SetText("日志系统未初始化")
		lp.logContent.Disable()
		return
	}

	logLines, err := lp.appState.Logger.GetLogs(100)
	if err != nil {
		lp.logContent.Enable()
		lp.logContent.SetText("获取日志失败: " + err.Error())
		lp.logContent.Disable()
		return
	}

	var filteredLines []string
	levelFilter := lp.levelSel.Selected
	typeFilter := lp.typeSel.Selected

	for _, line := range logLines {
		// 按级别过滤
		if levelFilter != "全部" && !lp.containsLogLevel(line, levelFilter) {
			continue
		}

		// 按类型过滤
		if !lp.containsLogType(line, typeFilter) {
			continue
		}

		filteredLines = append(filteredLines, line)
	}

	// 构建显示文本
	var b strings.Builder
	for _, line := range filteredLines {
		b.WriteString(line)
		b.WriteByte('\n')
	}

	// 更新日志内容
	lp.logContent.Enable()
	lp.logContent.SetText(b.String())
	lp.logContent.Disable()

	// 滚动到底部显示最新日志
	lp.logContent.CursorRow = len(filteredLines)
}

// containsLogLevel 检查日志行是否包含指定级别
func (lp *LogsPanel) containsLogLevel(logLine, level string) bool {
	if level == "全部" {
		return true
	}
	// 日志格式: timestamp [LEVEL] [type] message
	// 查找 [LEVEL] 格式
	levelPattern := "[" + level + "]"
	return strings.Contains(logLine, levelPattern)
}

// containsLogType 检查日志行是否包含指定类型
func (lp *LogsPanel) containsLogType(logLine, logType string) bool {
	if logType == "全部" {
		return true
	}
	// 日志格式: timestamp [LEVEL] [type] message
	// 查找 [type] 格式
	typePattern := "[" + logType + "]"
	return strings.Contains(logLine, typePattern)
}

// Refresh 刷新日志（供外部调用）
func (lp *LogsPanel) Refresh() {
	lp.refreshLogs()
}
