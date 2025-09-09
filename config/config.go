package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 结构体用于存储服务器配置
type Config struct {
	Server struct {
		Port    string `yaml:"port"`
		Timeout int    `yaml:"timeout"`
	} `yaml:"server"`

	Logging struct {
		Level     string `yaml:"level"`
		FilePath  string `yaml:"file_path"`
		AddSource bool   `yaml:"add_source"`
		JSON      bool   `yaml:"json"`
	} `yaml:"logging"`

	Bot struct {
		QQ struct {
			Id          int    `yaml:"id"`
			Uid         int    `yaml:"uid"`
			Secret      string `yaml:"secret"`
			Token       string `yaml:"token"`
			ScopeType   string `yaml:"type"`
			Sandbox     bool   `yaml:"sandbox"`
			WebhookPath string `yaml:"webhook_path"`
		} `yaml:"qq"`
		Telegram struct {
			Token string `yaml:"token"`
		} `yaml:"telegram"`
	} `yaml:"bot"`

	Channels []Channel `yaml:"channels"`
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
}

// WS 表示正向 WebSocket 服务器配置
type WS struct {
	Address       string `yaml:"address"`
	Authorization string `yaml:"authorization"`
}

// 全局配置变量
var config Config

// LoadConfig 从 YAML 文件加载配置
func LoadConfig(filename string) error {
	// 读取文件内容
	file, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("打开配置文件失败: %v", err)
	}

	err = yaml.Unmarshal(file, &config)
	if err != nil {
		panic(fmt.Errorf("无法解析 YAML 文件: %v", err))
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
