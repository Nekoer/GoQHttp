package main

import (
	"GoQHttp/config"
	"GoQHttp/internal"
	"GoQHttp/internal/constant"
	protocol "GoQHttp/internal/protocol/tencent"
	"GoQHttp/logger"
	"GoQHttp/utils"
	"GoQHttp/websocket/client"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var (
	QQClient protocol.Tencent
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
		QQClient.Init(w, r, constant.Configuration)
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

func displayBanner() {
	fmt.Println(" ██████   ██████   ██████  ██   ██ ████████ ████████ ██████  ")
	fmt.Println("██       ██    ██ ██    ██ ██   ██    ██       ██    ██   ██ ")
	fmt.Println("██   ███ ██    ██ ██    ██ ███████    ██       ██    ██████  ")
	fmt.Println("██    ██ ██    ██ ██ ▄▄ ██ ██   ██    ██       ██    ██      ")
	fmt.Println(" ██████   ██████   ██████  ██   ██    ██       ██    ██      ")
	fmt.Println("                      ▀▀                                     ")
	fmt.Println("                                                             ")
	fmt.Println(fmt.Sprintf("Project Version: %v", internal.Version))
}

func main() {
	displayBanner()
	// 加载配置
	err := config.LoadConfig("config.yml")
	if err != nil {
		return
	}

	constant.Configuration = config.GetConfig()

	// 初始化日志
	if err := initLogger(constant.Configuration); err != nil {
		fmt.Println(fmt.Errorf("初始化日志失败: %v", err))
	}
	defer func() {
		if constant.LogFile != nil {
			err := constant.LogFile.Close()
			if err != nil {
				logger.Errorf("日志关闭失败, %v", err)
			}
		}
	}()

	// 初始化数据库
	utils.SqLiteInit()

	// 设置 HTTP 路由
	http.HandleFunc("/health", healthHandler)

	if constant.Configuration.Bot.QQ.Enable {
		// 启动 QQ官方
		if constant.Configuration.Bot.QQ.WebhookPath != "" && constant.Configuration.Bot.QQ.Secret != "" && constant.Configuration.Bot.QQ.Id != 0 && constant.Configuration.Bot.QQ.Uid != 0 && constant.Configuration.Bot.QQ.Token != "" && constant.Configuration.Bot.QQ.ScopeType != "" {
			err := constant.OpenApi.Init(constant.Configuration.Bot.QQ.Id, constant.Configuration.Bot.QQ.Secret, constant.Configuration.Bot.QQ.Sandbox)
			if err != nil {
				logger.Errorf("GetAppAccessToken err: %v", err)
				return
			}

			http.HandleFunc(constant.Configuration.Bot.QQ.WebhookPath, webhookHandler)
			go QQClient.HandlerEvent()
			go constant.OpenApi.SendPacket()
		} else {
			logger.Warnf("QQ机器人启动失败,请检查设置")
		}
	}

	if constant.Configuration.Bot.Kook.Enable {
		if constant.Configuration.Bot.Kook.Token == "" {
			logger.Warnf("Kook机器人启动失败,请检查设置")
			return
		}
		kookClient := client.NewKookClient(constant.Configuration.Bot.Kook.Token, true)
		go kookClient.Connect()
	}

	// 功能端对接
	if len(constant.Configuration.Channels) > 0 {

		for _, channel := range constant.Configuration.Channels {
			if channel.WSReverse != nil {
				// 反向 WebSocket
				NoneBotClient := client.NewNoneBotClient(
					channel.WSReverse.Universal,
					int64(constant.Configuration.Bot.QQ.Uid),
					"Universal",
					channel.WSReverse.Authorization,
					channel.WSReverse.ReconnectInterval,
					constant.Configuration.Bot.QQ.Sandbox,
					channel.WSReverse.MaxRetries,
					channel.WSReverse.RetryDelay,
				)
				client.NoneBotManager.AddClient(NoneBotClient)
			} else {
				//hub := websocket.NewHub()
				//go hub.Run()
				//
				//http.HandleFunc(channel.WS.Address, func(w http.ResponseWriter, r *http.Request) {
				//	websocket.ServeWs(hub, w, r)
				//})
				//TODO 正向 WebSocket
			}

		}
		client.NoneBotManager.StartAll()
		go client.NoneBotManager.Broadcast()
	}

	// 创建带超时的 HTTP 服务器
	server := &http.Server{
		Addr:         ":" + constant.Configuration.Server.Port,
		ReadTimeout:  time.Duration(constant.Configuration.Server.Timeout) * time.Second,
		WriteTimeout: time.Duration(constant.Configuration.Server.Timeout) * time.Second,
	}

	// 启动服务器
	logger.Infof("服务器监听端口 %s", constant.Configuration.Server.Port)
	logger.Infof("QQ Webhook 监听地址: %s", constant.Configuration.Bot.QQ.WebhookPath)
	logger.Infof("日志级别: %s", constant.Configuration.Logging.Level)
	if constant.Configuration.Logging.FilePath != "" {
		logger.Infof("日志文件: %s", constant.Configuration.Logging.FilePath)
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Errorf("服务器启动失败: %v", err)
	}
}
