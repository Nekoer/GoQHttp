package logger

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
)

type LogConfig struct {
	Level     string
	FilePath  string
	AddSource bool
	JSON      bool
	Write     bool
}

var level slog.Leveler
var Logger *slog.Logger

func Init(config LogConfig) {
	switch config.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error", "fatal":
		level = slog.LevelError
	}

	// 输出目标
	var writers []io.Writer
	writers = append(writers, os.Stdout) // 总是输出到控制台

	if config.Write {
		// ✅ 仅当启用写入文件时才创建文件输出
		logfile, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			// 文件打开失败时，仍然继续运行，只打印警告
			slog.Warn("failed to open log file", "error", err)
		} else {
			writers = append(writers, bufio.NewWriter(logfile))
		}
	}

	outputIO := io.MultiWriter(writers...)
	opts := PrettyHandlerOptions{
		slog.HandlerOptions{
			Level:     level,
			AddSource: config.AddSource,
		},
	}
	prettyHandler := NewPrettyHandler(outputIO, opts)
	Logger = slog.New(prettyHandler)
}

func Infof(format string, args ...any) {
	Logger.Info(fmt.Sprintf(format, args...))
}
func Warnf(format string, args ...any) {
	Logger.Warn(fmt.Sprintf(format, args...))
}
func Debugf(format string, args ...any) {
	Logger.Debug(fmt.Sprintf(format, args...))
}
func Errorf(format string, args ...any) {
	Logger.Error(fmt.Sprintf(format, args...))
}

func Info(args ...any) {
	Logger.Info(fmt.Sprint(args...))
}

func Warn(args ...any) {
	Logger.Warn(fmt.Sprint(args...))
}

func Debug(args ...any) {
	Logger.Debug(fmt.Sprint(args...))
}
func Error(args ...any) {
	Logger.Error(fmt.Sprint(args...))
}
