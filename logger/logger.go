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

	logfile, _ := os.Open(config.FilePath)
	outputIO := io.MultiWriter(os.Stdout, bufio.NewWriter(logfile))
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
