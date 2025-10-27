package client

import (
	"GoQHttp/internal/protocol"
	"GoQHttp/logger"
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	logging "github.com/sacOO7/go-logger"
	"github.com/sacOO7/gowebsocket"
	uuid "github.com/satori/go.uuid"
)

// KookClient 表示一个 WebSocket 客户端连接
type KookClient struct {
	ID                   string
	Domain               string
	WssURL               string
	GatewayURL           string
	Conn                 gowebsocket.Socket
	mu                   sync.RWMutex
	Connected            bool
	MaxRetries           int64
	RetryDelay           time.Duration
	RetryCount           int64
	Authorization        string
	Interval             int64
	LastSuccessMessageId int64
	Resume               bool // 是否需要重连
	Compress             bool // 是否需要Zlib压缩
}
type KookRequestSignal int

const (
	MESSAGE KookRequestSignal = iota
	HANDSHAKE
	PING
	PONG
	RESUME
	RECONNECT
	RESUME_ACK
)

type KookMessage struct {
	Signal               KookRequestSignal `json:"s"`
	Data                 json.RawMessage   `json:"d"`
	LastSuccessMessageId int64             `json:"sn"`
}

// NewKookClient 创建新的 WebSocket 客户端
func NewKookClient(authorization string, compress bool) *KookClient {
	return &KookClient{
		ID:                   generateKookClientID(),
		Domain:               "https://www.kookapp.cn",
		GatewayURL:           "/api/v3/gateway/index",
		Connected:            false,
		MaxRetries:           0,
		RetryCount:           0,
		RetryDelay:           time.Duration(5) * time.Second,
		Authorization:        authorization,
		Interval:             3000,
		LastSuccessMessageId: 0,
		Compress:             compress,
	}
}

func generateKookClientID() string {
	return fmt.Sprintf("KookClient-%d", uuid.NewV4())
}

func (c *KookClient) GetWssURL() error {
	kookGetGatewayUrl := fmt.Sprintf("%s%s", c.Domain, c.GatewayURL)
	logger.Infof("尝试获取Kook GateWay Url: %s", kookGetGatewayUrl)

	if c.Resume && c.Compress {
		kookGetGatewayUrl = fmt.Sprintf("%s%s", kookGetGatewayUrl, "?resume=1&compress=1")
	} else if !c.Resume && c.Compress {
		kookGetGatewayUrl = fmt.Sprintf("%s%s", kookGetGatewayUrl, "?compress=1")
	} else {
		kookGetGatewayUrl = fmt.Sprintf("%s%s", kookGetGatewayUrl, "?compress=0")

	}

	r, err := http.NewRequest("GET", kookGetGatewayUrl, nil)
	if err != nil {
		logger.Errorf("Error creating request: %v", err)
		return err
	}

	r.Header = http.Header{
		"Content-Type":  []string{"application/json"},
		"Authorization": []string{"Bot " + c.Authorization},
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		logger.Errorf("Error sending request: %v", err)
		return err
	}

	defer resp.Body.Close()

	type Response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			URL string `json:"url"`
		} `json:"data"`
	}

	// 读取响应
	body, _ := io.ReadAll(resp.Body)
	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}
	c.GatewayURL = response.Data.URL
	return nil
}

// Connect 连接到 WebSocket 服务器
func (c *KookClient) Connect() {
	err := c.GetWssURL()
	if err != nil {
		return
	}

	logger.Infof("尝试连接到: %s", c.GatewayURL)

	// 解析 URL
	u, err := url.Parse(c.GatewayURL)
	if err != nil {
		logger.Warnf("URL 解析错误: %v", err)
		time.Sleep(5 * time.Second)
		return
	}

	header := http.Header{}
	if c.Authorization != "" {
		header.Set("Authorization", "Bot "+c.Authorization)
	}

	// 建立连接
	c.Conn = gowebsocket.New(u.String())
	c.Conn.RequestHeader = header
	c.Conn.WebsocketDialer.WriteBufferSize = 8192
	c.Conn.WebsocketDialer.ReadBufferSize = 8192
	c.Conn.GetLogger().SetLevel(logging.OFF)
	c.Conn.ConnectionOptions = gowebsocket.ConnectionOptions{
		UseSSL:         true,
		UseCompression: true,
		Subprotocols:   []string{},
	}

	c.Conn.OnConnected = func(socket gowebsocket.Socket) {
		c.mu.Lock()
		c.Connected = true
		c.RetryCount = 0 // 重置重试计数
		c.mu.Unlock()

		logger.Infof("已连接 %v", u.String())
	}
	c.Conn.OnConnectError = func(err error, socket gowebsocket.Socket) {
		c.mu.Lock()
		c.Connected = false
		c.mu.Unlock()
		logger.Errorf("%v 连接出错:  %v", c.GatewayURL, err.Error())
	}
	c.Conn.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		c.handleMessage([]byte(message))
	}
	c.Conn.OnBinaryMessage = func(message []byte, socket gowebsocket.Socket) {

		if c.Compress {
			b := bytes.NewBuffer(message)
			r, err := zlib.NewReader(b)
			if err != nil {
				panic(err)
			}
			defer r.Close()

			decompressedData, err := io.ReadAll(r)
			if err != nil {
				panic(err)
			}
			c.handleMessage(decompressedData)
		}
	}
	c.Conn.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		c.mu.Lock()
		c.Connected = false
		c.mu.Unlock()
		logger.Warnf("%v 断开连接", c.GatewayURL)
	}
	c.Conn.Connect()
	for {
		if !c.Connected {
			return
		}
		c.HeartBeat()
		time.Sleep(time.Second * time.Duration(30))
	}
}

func (c *KookClient) Daemon() {
	for {
		if c.MaxRetries != 0 {
			if c.RetryCount >= c.MaxRetries && c.MaxRetries > 0 {
				logger.Warnf("达到最大重试次数 (%d)，停止连接: %s", c.MaxRetries, c.GatewayURL)
				return
			}
		}

		c.Connect()

		c.mu.Lock()
		c.Connected = false
		c.RetryCount++
		c.mu.Unlock()

		if c.RetryDelay > 0 {
			logger.Warnf("%v将在%v后重连", c.GatewayURL, c.RetryDelay)
			time.Sleep(c.RetryDelay)
		}

	}
}

func (c *KookClient) HeartBeat() {
	c.mu.Lock()
	req := KookMessage{
		Signal:               PING,
		Data:                 nil,
		LastSuccessMessageId: c.LastSuccessMessageId,
	}
	msg, _ := json.Marshal(req)
	c.Conn.SendText(string(msg))
	c.mu.Unlock()
}

func (c *KookClient) SendMessage(data string) {
	c.mu.Lock()
	logger.Debug(data)
	if c.Connected {
		c.Conn.SendText(data)
	}
	c.mu.Unlock()
}

// handleMessage 处理接收到的消息
func (c *KookClient) handleMessage(message []byte) {
	logger.Debugf("%s 接收数据 %s", c.GatewayURL, string(message))
	var request KookMessage
	err := json.Unmarshal(message, &request)
	if err != nil {
		logger.Warnf("无法解析消息: %v", err)
		return
	}

	// 根据信令处理
	switch request.Signal {
	case HANDSHAKE:
		logger.Infof("[Kook] 连接事件: %v", string(message))
		type Data struct {
			Code       int    `json:"code"`
			SessionIdB string `json:"sessionId"`
			SessionId  string `json:"session_id"`
		}
		var data Data
		err := json.Unmarshal(request.Data, &data)
		if err != nil {
			logger.Error(err)
			return
		}
		if data.Code != 0 {
			switch data.Code {
			case 40100:
				logger.Warn("缺少参数")
			case 40101:
				logger.Warn("无效的 token")
			case 40102:
				logger.Warn("token 验证失败")
			case 40103:
				//todo 需要重新连接
				logger.Warn("token 过期")
			}
		}

	case PONG:
		logger.Infof("[Kook] 心跳事件: %v", string(message))
	case RECONNECT:
		logger.Infof("[Kook] 服务器要求重连事件: %v", string(message))
	case RESUME_ACK:
		logger.Infof("[Kook] 连接恢复事件: %v", string(message))
	case MESSAGE:
		logger.Infof("[Kook] 消息事件: %s", string(message))
		c.LastSuccessMessageId = request.LastSuccessMessageId
		protocol.KookChan <- request.Data
	default:
		logger.Warnf("不支持处理的事件: %s", string(message))
	}
}

// Close 关闭连接
func (c *KookClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Connected {
		c.Conn.SendBinary([]byte{})
		c.Conn.Close()
		c.Connected = false

	}
}
