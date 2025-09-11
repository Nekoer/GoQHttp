package main

import (
	"GoQHttp/config"
	"GoQHttp/logger"
	"GoQHttp/protocol/tencent"
	"GoQHttp/utils"
	"GoQHttp/websocket"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var logFile *os.File
var configuration *config.Config

// initLogger 初始化日志系统
func initLogger(config *config.Config) error {
	// 创建日志记录器
	logConfig := logger.LogConfig{
		Level:     config.Logging.Level,
		FilePath:  config.Logging.FilePath,
		AddSource: config.Logging.AddSource,
		JSON:      config.Logging.JSON,
	}
	logger.Init(logConfig)
	return nil
}

// webhookHandler 处理传入的 webhook 请求
func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/qq" {
		var client tencent.Tencent
		client.Init(w, r, configuration)
	}

}

// healthHandler 提供健康检查端点
func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	_ = json.NewEncoder(w).Encode(response)
}

func main() {
	// 加载配置
	err := config.LoadConfig("config.yml")
	if err != nil {
		fmt.Printf("加载配置失败: %v", err)
	}

	configuration = config.GetConfig()

	// 初始化日志
	if err := initLogger(configuration); err != nil {
		logger.Warnf("初始化日志失败: %v", err)
	}
	defer func() {
		if logFile != nil {
			err := logFile.Close()
			if err != nil {
				logger.Errorf("日志关闭失败, %v", err)
			}
		}
	}()

	// 初始化数据库
	utils.SqLiteInit()

	// 设置 HTTP 路由
	http.HandleFunc("/health", healthHandler)

	if configuration.Bot.QQ.WebhookPath != "" {
		http.HandleFunc(configuration.Bot.QQ.WebhookPath, webhookHandler)
		go tencent.HandlerEvent()
	}
	if len(configuration.Channels) > 0 {
		for _, channel := range configuration.Channels {
			if channel.WSReverse != nil {
				// 反向 WebSocket
				client := websocket.NewWebSocketClient(
					channel.WSReverse.Universal,
					int64(configuration.Bot.QQ.Uid),
					configuration.Bot.QQ.Id,
					configuration.Bot.QQ.Secret,
					"Universal",
					channel.WSReverse.Authorization,
					channel.WSReverse.ReconnectInterval,
					configuration.Bot.QQ.Sandbox,
				)
				websocket.Manager.AddClient(client)
				go client.Connect()
			} else {
				// 正向 WebSocket
			}

		}
	}

	// 创建带超时的 HTTP 服务器
	server := &http.Server{
		Addr:         ":" + configuration.Server.Port,
		ReadTimeout:  time.Duration(configuration.Server.Timeout) * time.Second,
		WriteTimeout: time.Duration(configuration.Server.Timeout) * time.Second,
	}

	// 启动服务器
	logger.Infof("服务器监听端口 %s", configuration.Server.Port)
	logger.Infof("QQ Webhook 监听地址: %s", configuration.Bot.QQ.WebhookPath)
	logger.Infof("日志级别: %s", configuration.Logging.Level)
	if configuration.Logging.FilePath != "" {
		logger.Infof("日志文件: %s", configuration.Logging.FilePath)
	}
	// 等待中断信号
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	if err := server.ListenAndServe(); err != nil {
		logger.Errorf("服务器启动失败: %v", err)
	}

	<-interrupt
	logger.Info("接收到中断信号，关闭所有连接...")

	// 关闭所有连接
	websocket.Manager.CloseAll()
	logger.Info("所有连接已关闭，程序退出")
}
