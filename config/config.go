package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Server struct {
	Port    string `yaml:"port"`
	Timeout int    `yaml:"timeout"`
}
type Logging struct {
	Level     string `yaml:"level"`
	FilePath  string `yaml:"file_path"`
	AddSource bool   `yaml:"add_source"`
	JSON      bool   `yaml:"json"`
}

type QQ struct {
	Enable      bool   `yaml:"enable"`
	Id          int    `yaml:"id"`
	Uid         int    `yaml:"uid"`
	Secret      string `yaml:"secret"`
	Token       string `yaml:"token"`
	ScopeType   string `yaml:"type"`
	Sandbox     bool   `yaml:"sandbox"`
	WebhookPath string `yaml:"webhook_path"`
}

type Telegram struct {
	Enable bool   `yaml:"enable"`
	Token  string `yaml:"token"`
}

type Kook struct {
	Enable bool   `yaml:"enable"`
	Token  string `yaml:"token"`
}
type Bot struct {
	QQ       QQ       `yaml:"qq"`
	Telegram Telegram `yaml:"telegram"`
	Kook     Kook     `yaml:"kook"`
}

// Channel 表示单个服务器配置
type Channel struct {
	WSReverse *WSReverse `yaml:"ws-reverse,omitempty"`
	WS        *WS        `yaml:"ws,omitempty"`
}

// WSReverse 表示反向 WebSocket 连接配置
type WSReverse struct {
	Universal         string `yaml:"universal"`
	ReconnectInterval int64  `yaml:"reconnect-interval"`
	Authorization     string `yaml:"authorization"`
	RetryDelay        int64  `yaml:"retry-delay"`
	MaxRetries        int64  `yaml:"max-retries"`
}

// WS 表示正向 WebSocket 服务器配置
type WS struct {
	Address       string `yaml:"address"`
	Authorization string `yaml:"authorization"`
}

// Config 结构体用于存储服务器配置
type Config struct {
	Server   Server    `yaml:"server"`
	Logging  Logging   `yaml:"logging"`
	Bot      Bot       `yaml:"bot"`
	Channels []Channel `yaml:"channels"`
}

// 全局配置变量
var config Config

func WriteConfig(filename string) {
	config = Config{
		Server: Server{
			Port:    "8080",
			Timeout: 10,
		},
		Logging: Logging{
			Level:     "info",
			FilePath:  "./webhook.log",
			AddSource: true,
			JSON:      false,
		},
		Bot: Bot{
			QQ: QQ{
				Enable:      false,
				Id:          1,
				Uid:         1,
				Secret:      "",
				Token:       "",
				ScopeType:   "public",
				Sandbox:     false,
				WebhookPath: "/qq",
			},
			Telegram: Telegram{
				Enable: false,
				Token:  "",
			},
			Kook: Kook{
				Enable: false,
				Token:  "",
			},
		},
		Channels: []Channel{
			{
				WSReverse: &WSReverse{
					Universal:         "ws-reverse",
					ReconnectInterval: 3000,
					Authorization:     "",
				},
				WS: &WS{
					Address:       "",
					Authorization: "",
				},
			},
		},
	}

	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// 创建编码器
	encoder := yaml.NewEncoder(file)

	// 将配置编码为 YAML 数据
	err = encoder.Encode(&config)
	if err != nil {
		fmt.Println("Error encoding YAML:", err)
		return
	}

	fmt.Println("检测到config.yml不存在")
}

// LoadConfig 从 YAML 文件加载配置
func LoadConfig(filename string) error {
	// 读取文件内容
	file, err := os.ReadFile(filename)
	if err != nil {
		WriteConfig(filename)
		fmt.Println("初始化配置文件完毕,请配置文件后再运行程序")
		fmt.Println("按下回车键退出……")
		var input string
		fmt.Scanln(&input) // 等待用户输入回车
		os.Exit(0)
	}

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return fmt.Errorf("无法解析 YAML 文件: %v", err)
	}

	// 设置默认值
	if config.Server.Port == "" {
		config.Server.Port = "8080"
	}

	if config.Server.Timeout == 0 {
		config.Server.Timeout = 30
	}
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}

	return nil
}

func GetConfig() *Config {
	return &config
}
