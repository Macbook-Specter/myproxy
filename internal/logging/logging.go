package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	// LevelDebug 调试级别
	LevelDebug LogLevel = iota
	// LevelInfo 信息级别
	LevelInfo
	// LevelWarn 警告级别
	LevelWarn
	// LevelError 错误级别
	LevelError
	// LevelFatal 致命级别
	LevelFatal
)

var levelNames = map[LogLevel]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
	LevelFatal: "FATAL",
}

// LogType 日志类型
type LogType string

const (
	// LogTypeApp 应用程序日志
	LogTypeApp LogType = "app"
	// LogTypeProxy 代理转发日志
	LogTypeProxy LogType = "proxy"
)

// Logger 日志记录器
type Logger struct {
	level       LogLevel
	files       map[LogType]*os.File
	console     bool
	mutex       sync.Mutex
	logFilePath string
	logDir      string
}

// NewLogger 创建新的日志记录器
func NewLogger(logFilePath string, console bool, level string) (*Logger, error) {
	// 解析日志级别
	logLevel, err := parseLogLevel(level)
	if err != nil {
		return nil, err
	}

	// 获取日志目录
	logDir := filepath.Dir(logFilePath)
	baseName := filepath.Base(logFilePath)
	// 移除扩展名以获取基本名称
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))

	logger := &Logger{
		level:       logLevel,
		console:     console,
		logFilePath: logFilePath,
		logDir:      logDir,
		files:       make(map[LogType]*os.File),
	}

	// 创建目录（如果不存在）
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 打开应用日志文件（追加模式）
	appLogPath := fmt.Sprintf("%s/%s_%s.log", logDir, baseName, LogTypeApp)
	appFile, err := os.OpenFile(appLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("打开应用日志文件失败: %w", err)
	}
	logger.files[LogTypeApp] = appFile

	// 打开代理日志文件（追加模式）
	proxyLogPath := fmt.Sprintf("%s/%s_%s.log", logDir, baseName, LogTypeProxy)
	proxyFile, err := os.OpenFile(proxyLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("打开代理日志文件失败: %w", err)
	}
	logger.files[LogTypeProxy] = proxyFile

	return logger, nil
}

// parseLogLevel 解析日志级别字符串
func parseLogLevel(level string) (LogLevel, error) {
	level = strings.ToLower(level)
	// 如果日志级别为空，返回默认级别
	if level == "" {
		return LevelInfo, nil
	}
	switch level {
	case "debug":
		return LevelDebug, nil
	case "info":
		return LevelInfo, nil
	case "warn":
		return LevelWarn, nil
	case "error":
		return LevelError, nil
	case "fatal":
		return LevelFatal, nil
	default:
		return LevelInfo, fmt.Errorf("无效的日志级别: %s", level)
	}
}

// log 记录日志
func (l *Logger) log(level LogLevel, logType LogType, format string, args ...interface{}) {
	// 检查日志级别
	if level < l.level {
		return
	}

	// 生成日志消息
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelName := levelNames[level]
	message := fmt.Sprintf(format, args...)
	// 在日志中添加类型标识
	logLine := fmt.Sprintf("%s [%s] [%s] %s\n", timestamp, levelName, logType, message)

	l.mutex.Lock()
	defer l.mutex.Unlock()

	// 输出到控制台
	if l.console {
		fmt.Print(logLine)
	}

	// 输出到对应类型的日志文件
	file := l.files[logType]
	if file != nil {
		if _, err := file.WriteString(logLine); err != nil {
			// 如果写入文件失败，尝试重新打开文件
			l.reopenFile(logType)
			// 再次尝试写入
			l.files[logType].WriteString(logLine)
		}
	}

	// 同时写入应用日志作为备份
	if logType != LogTypeApp {
		appFile := l.files[LogTypeApp]
		if appFile != nil {
			appFile.WriteString(logLine)
		}
	}

	// 如果是致命错误，退出程序
	if level == LevelFatal {
		os.Exit(1)
	}
}

// reopenFile 重新打开日志文件
func (l *Logger) reopenFile(logType LogType) {
	file := l.files[logType]
	if file != nil {
		file.Close()
	}

	// 构建对应类型的日志文件路径
	baseName := filepath.Base(l.logFilePath)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
	logPath := fmt.Sprintf("%s/%s_%s.log", l.logDir, baseName, logType)

	newFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		l.files[logType] = newFile
	}
}

// Debug 记录调试日志（默认应用日志）
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, LogTypeApp, format, args...)
}

// DebugWithType 记录指定类型的调试日志
func (l *Logger) DebugWithType(logType LogType, format string, args ...interface{}) {
	l.log(LevelDebug, logType, format, args...)
}

// Info 记录信息日志（默认应用日志）
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, LogTypeApp, format, args...)
}

// InfoWithType 记录指定类型的信息日志
func (l *Logger) InfoWithType(logType LogType, format string, args ...interface{}) {
	l.log(LevelInfo, logType, format, args...)
}

// Warn 记录警告日志（默认应用日志）
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, LogTypeApp, format, args...)
}

// WarnWithType 记录指定类型的警告日志
func (l *Logger) WarnWithType(logType LogType, format string, args ...interface{}) {
	l.log(LevelWarn, logType, format, args...)
}

// Error 记录错误日志（默认应用日志）
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, LogTypeApp, format, args...)
}

// ErrorWithType 记录指定类型的错误日志
func (l *Logger) ErrorWithType(logType LogType, format string, args ...interface{}) {
	l.log(LevelError, logType, format, args...)
}

// Fatal 记录致命日志（默认应用日志）
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(LevelFatal, LogTypeApp, format, args...)
}

// FatalWithType 记录指定类型的致命日志
func (l *Logger) FatalWithType(logType LogType, format string, args ...interface{}) {
	l.log(LevelFatal, logType, format, args...)
}

// Close 关闭日志记录器
func (l *Logger) Close() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// 关闭所有日志文件
	for logType, file := range l.files {
		if file != nil {
			file.Close()
			l.files[logType] = nil
		}
	}
}

// Rotate 日志轮转
func (l *Logger) Rotate() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	baseName := filepath.Base(l.logFilePath)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
	timestamp := time.Now().Format("20060102150405")

	// 对每个日志文件进行轮转
	for logType, file := range l.files {
		if file != nil {
			file.Close()
		}

		// 构建当前日志文件路径
		logPath := fmt.Sprintf("%s/%s_%s.log", l.logDir, baseName, logType)
		
		// 备份当前日志文件
		if _, err := os.Stat(logPath); err == nil {
			backupPath := fmt.Sprintf("%s.%s", logPath, timestamp)
			if err := os.Rename(logPath, backupPath); err != nil {
				return fmt.Errorf("备份日志文件失败: %w", err)
			}
		}

		// 重新打开日志文件
		newFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("重新打开日志文件失败: %w", err)
		}

		l.files[logType] = newFile
	}

	return nil
}

// GetLogs 获取所有日志内容，合并所有日志文件的内容
func (l *Logger) GetLogs(lines int) ([]string, error) {
	var allLines []string

	// 获取所有日志文件的内容
	baseName := filepath.Base(l.logFilePath)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))

	// 按日志类型顺序获取日志（应用日志优先）
	logTypes := []LogType{LogTypeApp, LogTypeProxy}
	for _, logType := range logTypes {
		logPath := fmt.Sprintf("%s/%s_%s.log", l.logDir, baseName, logType)
		
		// 打开日志文件
		file, err := os.Open(logPath)
		if err != nil {
			continue // 忽略不存在的文件
		}
		
		// 读取文件内容
		content, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			continue // 忽略读取错误
		}

		// 按行分割
		logLines := strings.Split(string(content), "\n")
		// 移除最后一个空行
		if len(logLines) > 0 && logLines[len(logLines)-1] == "" {
			logLines = logLines[:len(logLines)-1]
		}

		// 添加到所有行
		allLines = append(allLines, logLines...)
	}

	// 返回最后 N 行
	start := 0
	if len(allLines) > lines {
		start = len(allLines) - lines
	}

	return allLines[start:], nil
}

// GetLogsByType 获取指定类型的日志内容
func (l *Logger) GetLogsByType(logType LogType, lines int) ([]string, error) {
	baseName := filepath.Base(l.logFilePath)
	baseName = strings.TrimSuffix(baseName, filepath.Ext(baseName))
	logPath := fmt.Sprintf("%s/%s_%s.log", l.logDir, baseName, logType)

	// 打开日志文件
	file, err := os.Open(logPath)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %w", err)
	}
	defer file.Close()

	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("读取日志文件失败: %w", err)
	}

	// 按行分割
	logLines := strings.Split(string(content), "\n")
	// 移除最后一个空行
	if len(logLines) > 0 && logLines[len(logLines)-1] == "" {
		logLines = logLines[:len(logLines)-1]
	}

	// 返回最后 N 行
	start := 0
	if len(logLines) > lines {
		start = len(logLines) - lines
	}

	return logLines[start:], nil
}
