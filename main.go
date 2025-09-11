package main

import (
	"GoQHttp/config"
	"GoQHttp/logger"
	"GoQHttp/protocol/tencent"
	"GoQHttp/utils"
	"GoQHttp/websocket"
	"encoding/json"
	"net/http"
	"os"
	"time"
)

var (
	logFile       *os.File
	configuration *config.Config
	TencentClient websocket.Tencent
)

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
		return
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
		err := TencentClient.Init(configuration.Bot.QQ.Id, configuration.Bot.QQ.Secret, false)
		if err != nil {
			logger.Errorf("GetAppAccessToken err: %v", err)
			return
		}
		for _, channel := range configuration.Channels {
			logger.Infof("%+v", channel)
			if channel.WSReverse != nil {
				// 反向 WebSocket
				client := websocket.NewWebSocketClient(
					channel.WSReverse.Universal,
					int64(configuration.Bot.QQ.Uid),
					"Universal",
					channel.WSReverse.Authorization,
					channel.WSReverse.ReconnectInterval,
					configuration.Bot.QQ.Sandbox,
					TencentClient,
					channel.WSReverse.MaxRetries,
					channel.WSReverse.RetryDelay,
				)
				websocket.Manager.AddClient(client)
			} else {
				//TODO 正向 WebSocket
			}

		}
		websocket.Manager.StartAll()
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

	if err := server.ListenAndServe(); err != nil {
		logger.Errorf("服务器启动失败: %v", err)
	}
}
