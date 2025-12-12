package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// LogEntry 表示一条日志条目
type LogEntry struct {
	Timestamp time.Time
	Level     string
	Type      string
	Message   string
	Line      string // 完整的日志行
}

// LogsPanel 管理应用日志和代理日志的显示。
// 它支持按日志级别和类型过滤，并提供追加日志功能。
type LogsPanel struct {
	appState      *AppState
	logContent    *widget.RichText // 使用 RichText 以支持自定义文本颜色
	levelSel      *widget.Select
	typeSel       *widget.Select
	logBuffer     []LogEntry // 日志缓冲区
	bufferMutex   sync.Mutex // 保护日志缓冲区的互斥锁
	maxBufferSize int        // 最大缓冲区大小
}

// NewLogsPanel 创建并初始化日志显示面板。
// 该方法会创建日志内容区域、过滤控件，并加载初始日志。
// 参数：
//   - appState: 应用状态实例
//
// 返回：初始化后的日志面板实例
func NewLogsPanel(appState *AppState) *LogsPanel {
	lp := &LogsPanel{
		appState:      appState,
		logBuffer:     make([]LogEntry, 0),
		maxBufferSize: 1000, // 最多保存1000条日志
	}

	// 日志内容 - 使用 RichText 以支持自定义文本颜色
	lp.logContent = widget.NewRichText()
	lp.logContent.Wrapping = fyne.TextWrapOff // 关闭自动换行，使用水平滚动
	// 设置等宽字体样式
	lp.logContent.Segments = []widget.RichTextSegment{}

	// 日志级别选择
	lp.levelSel = widget.NewSelect(
		[]string{"全部", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"},
		func(value string) {
			if lp.typeSel != nil { // 确保 typeSel 已初始化
				lp.refreshDisplay()
			}
		},
	)

	// 日志类型选择
	lp.typeSel = widget.NewSelect(
		[]string{"全部", "app", "proxy"},
		func(value string) {
			if lp.levelSel != nil { // 确保 levelSel 已初始化
				lp.refreshDisplay()
			}
		},
	)

	// 等所有组件创建完成后再设置默认值和刷新
	lp.levelSel.SetSelected("全部")
	lp.typeSel.SetSelected("全部")

	// 初始加载历史日志
	lp.loadInitialLogs()

	return lp
}

// Build 构建并返回日志显示面板的 UI 组件。
// 返回：包含过滤控件和日志内容的容器组件
func (lp *LogsPanel) Build() fyne.CanvasObject {
	// 顶部控制栏
	topBar := container.NewHBox(
		widget.NewLabel("日志: "),
		container.NewGridWrap(fyne.NewSize(80, 40), lp.levelSel),
		widget.NewLabel("类型: "),
		container.NewGridWrap(fyne.NewSize(80, 40), lp.typeSel),
		widget.NewButton("刷新", func() {
			lp.loadInitialLogs()
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

// AppendLog 追加一条日志到日志面板（线程安全）
// 该方法可以从任何地方调用，会自动追加到日志缓冲区并更新显示
func (lp *LogsPanel) AppendLog(level, logType, message string) {
	if lp == nil {
		return
	}

	// 构建完整的日志行
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logLine := fmt.Sprintf("%s [%s] [%s] %s", timestamp, level, logType, message)

	// 创建日志条目
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Type:      logType,
		Message:   message,
		Line:      logLine,
	}

	// 线程安全地追加到缓冲区
	lp.bufferMutex.Lock()
	lp.logBuffer = append(lp.logBuffer, entry)

	// 如果超过最大大小，删除最旧的日志
	if len(lp.logBuffer) > lp.maxBufferSize {
		lp.logBuffer = lp.logBuffer[len(lp.logBuffer)-lp.maxBufferSize:]
	}
	lp.bufferMutex.Unlock()

	// 更新显示
	lp.refreshDisplay()
}

// loadInitialLogs 加载初始日志（从文件加载历史日志）
func (lp *LogsPanel) loadInitialLogs() {
	if lp.appState == nil || lp.appState.Logger == nil {
		return
	}

	logLines, err := lp.appState.Logger.GetLogs(100)
	if err != nil {
		return
	}

	// 解析日志行并添加到缓冲区
	lp.bufferMutex.Lock()
	lp.logBuffer = make([]LogEntry, 0, len(logLines))
	for _, line := range logLines {
		entry := lp.parseLogLine(line)
		if entry != nil {
			lp.logBuffer = append(lp.logBuffer, *entry)
		}
	}
	lp.bufferMutex.Unlock()

	// 刷新显示
	lp.refreshDisplay()
}

// parseLogLine 解析日志行，提取级别、类型和消息
func (lp *LogsPanel) parseLogLine(line string) *LogEntry {
	// 日志格式: timestamp [LEVEL] [type] message
	// 例如: 2025-01-01 12:00:00 [INFO] [app] 这是一条消息

	// 查找第一个 [ 和 ]
	levelStart := strings.Index(line, "[")
	if levelStart == -1 {
		return nil
	}
	levelEnd := strings.Index(line[levelStart:], "]")
	if levelEnd == -1 {
		return nil
	}
	levelEnd += levelStart

	// 提取时间戳和级别
	timestampStr := strings.TrimSpace(line[:levelStart])
	level := line[levelStart+1 : levelEnd]

	// 查找第二个 [ 和 ]
	typeStart := strings.Index(line[levelEnd+1:], "[")
	if typeStart == -1 {
		return nil
	}
	typeStart += levelEnd + 1
	typeEnd := strings.Index(line[typeStart:], "]")
	if typeEnd == -1 {
		return nil
	}
	typeEnd += typeStart

	// 提取类型和消息
	logType := line[typeStart+1 : typeEnd]
	message := strings.TrimSpace(line[typeEnd+1:])

	// 解析时间戳
	timestamp, err := time.Parse("2006-01-02 15:04:05", timestampStr)
	if err != nil {
		timestamp = time.Now()
	}

	return &LogEntry{
		Timestamp: timestamp,
		Level:     level,
		Type:      logType,
		Message:   message,
		Line:      line,
	}
}

// refreshDisplay 根据当前过滤条件刷新显示
func (lp *LogsPanel) refreshDisplay() {
	if lp.logContent == nil || lp.levelSel == nil || lp.typeSel == nil {
		return
	}

	lp.bufferMutex.Lock()
	defer lp.bufferMutex.Unlock()

	levelFilter := lp.levelSel.Selected
	typeFilter := lp.typeSel.Selected

	// 过滤日志
	var filteredEntries []LogEntry
	for _, entry := range lp.logBuffer {
		// 按级别过滤
		if levelFilter != "全部" && entry.Level != levelFilter {
			continue
		}

		// 按类型过滤
		if typeFilter != "全部" && entry.Type != typeFilter {
			continue
		}

		filteredEntries = append(filteredEntries, entry)
	}

	// 构建显示文本
	var segments []widget.RichTextSegment
	for _, entry := range filteredEntries {
		segments = append(segments, &widget.TextSegment{
			Text: entry.Line + "\n",
			Style: widget.RichTextStyle{
				ColorName: theme.ColorNameForeground,
				TextStyle: fyne.TextStyle{Monospace: true},
			},
		})
	}

	// 更新日志内容
	fyne.Do(func() {
		lp.logContent.Segments = segments
		lp.logContent.Refresh()
	})
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

// Refresh 刷新日志显示，重新应用当前过滤条件。
func (lp *LogsPanel) Refresh() {
	lp.refreshDisplay()
}
