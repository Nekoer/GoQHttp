package websocket

import (
	"GoQHttp/constant"
	"GoQHttp/logger"
	"GoQHttp/onebot"
	"GoQHttp/protocol"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	logging "github.com/sacOO7/go-logger"
	"github.com/sacOO7/gowebsocket"
	uuid "github.com/satori/go.uuid"
)

// Client 表示一个 WebSocket 客户端连接
type Client struct {
	ID            string
	AppId         int
	Secret        string
	URL           string
	Conn          gowebsocket.Socket
	mu            sync.RWMutex
	Connected     bool
	MaxRetries    int64
	RetryDelay    time.Duration
	RetryCount    int64
	XSelfID       int64
	XClientRole   string
	Authorization string
	Interval      int64
	Sandbox       bool
}

// ClientManager 管理多个 WebSocket 客户端连接
type ClientManager struct {
	clients map[string]*Client
	mu      sync.RWMutex
}

type Request struct {
	Action string      `json:"action"`
	Params any         `json:"params"`
	Echo   interface{} `json:"echo"`
}
type Response struct {
	Status  string `json:"status"`  // 执行状态，必须是 ok、failed 中的一个
	Code    int64  `json:"retcode"` // 返回码
	Data    any    `json:"data"`    // 响应数据
	Message string `json:"message"` // 错误信息
	Echo    any    `json:"echo"`    // 动作请求中的 echo 字段值
}

var (
	Manager *ClientManager
)

func init() {
	Manager = &ClientManager{
		clients: make(map[string]*Client),
	}
}

// NewWebSocketClient 创建新的 WebSocket 客户端
func NewWebSocketClient(url string, xSelfID int64, xClientRole string, authorization string, interval int64, sandbox bool, maxRetries int64, retryDelay int64) *Client {

	return &Client{
		ID:            generateID(),
		URL:           url,
		Connected:     false,
		MaxRetries:    maxRetries,
		RetryCount:    0,
		RetryDelay:    time.Duration(retryDelay) * time.Second,
		XSelfID:       xSelfID,
		XClientRole:   xClientRole,
		Authorization: authorization,
		Interval:      interval,
		Sandbox:       sandbox,
	}
}

// generateID 生成唯一客户端 ID
func generateID() string {
	return fmt.Sprintf("client-%d", uuid.NewV4())
}

// Connect 连接到 WebSocket 服务器
func (c *Client) Connect() {
	logger.Infof("尝试连接到: %s", c.URL)

	// 解析 URL
	u, err := url.Parse(c.URL)
	if err != nil {
		logger.Warnf("URL 解析错误: %v", err)
		time.Sleep(5 * time.Second)
		return
	}

	header := http.Header{}
	header.Set("X-Self-ID", strconv.FormatInt(c.XSelfID, 10))
	header.Set("X-Client-Role", c.XClientRole)
	if c.Authorization != "" {
		header.Set("Authorization", "Bearer "+c.Authorization)
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

		c.LifeCycle()

		logger.Infof("已连接 %v", u.String())
	}
	c.Conn.OnConnectError = func(err error, socket gowebsocket.Socket) {
		c.mu.Lock()
		c.Connected = false
		c.mu.Unlock()
		logger.Errorf("%v 连接出错:  %v", c.URL, err.Error())
	}
	c.Conn.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		c.handleMessage([]byte(message))
	}
	c.Conn.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		c.mu.Lock()
		c.Connected = false
		c.mu.Unlock()
		logger.Warnf("%v 断开连接", c.URL)
	}
	c.Conn.Connect()
	for {
		if !c.Connected {
			return
		}
		c.HeartBeat()
		time.Sleep(time.Second * time.Duration(5))
	}
}

func (c *Client) Daemon() {
	for {
		if c.MaxRetries != 0 {
			if c.RetryCount >= c.MaxRetries && c.MaxRetries > 0 {
				logger.Warnf("达到最大重试次数 (%d)，停止连接: %s", c.MaxRetries, c.URL)
				return
			}
		}

		c.Connect()

		c.mu.Lock()
		c.Connected = false
		c.RetryCount++
		c.mu.Unlock()

		if c.RetryDelay > 0 {
			logger.Warnf("%v将在%v后重连", c.URL, c.RetryDelay)
			time.Sleep(c.RetryDelay)
		}

	}
}

func (c *Client) LifeCycle() {
	c.mu.Lock()
	lifeCycle := onebot.LifeCycle{
		MessageBase: onebot.MessageBase{
			Time:     time.Now().Unix(),
			SelfId:   c.XSelfID,
			PostType: onebot.MetaEventPost,
		},
		MetaEventType: onebot.LifecycleType,
		SubType:       onebot.Connect,
	}
	msg, _ := json.Marshal(lifeCycle)
	c.Conn.SendText(string(msg))
	c.mu.Unlock()
}

func (c *Client) HeartBeat() {
	c.mu.Lock()
	heartbeat := onebot.Heartbeat{
		MessageBase: onebot.MessageBase{
			Time:     time.Now().Unix(),
			SelfId:   c.XSelfID,
			PostType: onebot.MetaEventPost,
		},
		MetaEventType: onebot.LifecycleType,
		Interval:      c.Interval,
		Status: onebot.Status{
			Online: true,
			Good:   false,
		},
	}
	msg, _ := json.Marshal(heartbeat)
	c.Conn.SendText(string(msg))
	c.mu.Unlock()
}

func (c *Client) SendMessage(data string) {
	c.mu.Lock()
	logger.Debug(data)
	if c.Connected {
		c.Conn.SendText(data)
	}
	c.mu.Unlock()
}

// handleMessage 处理接收到的消息
func (c *Client) handleMessage(message []byte) {
	logger.Debugf("%s 接收数据 %s", c.URL, string(message))
	var request Request
	err := json.Unmarshal(message, &request)
	if err != nil {
		logger.Warnf("无法解析消息: %v", err)
		return
	}

	var messageRequest *onebot.MessageRequest
	marshal, err := json.Marshal(request.Params)

	if err != nil {
		return
	}
	// onebot v11
	err = json.Unmarshal(marshal, &messageRequest)

	if err != nil {
		messages, err := constant.CQCode.ParseAllCQCodes(string(marshal))
		if err != nil {
			return
		}

		messageRequest.Message = messages
	}

	// 根据消息类型处理
	switch messageRequest.PostType {
	case onebot.NoticePost:
		logger.Infof("[%s] 通知事件: %v\n", c.URL, string(message))
	case onebot.RequestPost:
		logger.Infof("[%s] 请求事件: %v\n", c.URL, string(message))
	case onebot.MessagePost:
	default:
		protocol.OfficialChan <- messageRequest
		logger.Infof("[%s] 消息事件: %s\n", c.URL, string(message))
	}
}

// Close 关闭连接
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Connected {
		c.Conn.SendBinary([]byte{})
		c.Conn.Close()
		c.Connected = false

	}
}

// AddClient 添加客户端到管理器
func (m *ClientManager) AddClient(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[client.ID] = client
}

// RemoveClient 从管理器移除客户端
func (m *ClientManager) RemoveClient(clientID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if client, exists := m.clients[clientID]; exists {
		client.Close()
		delete(m.clients, clientID)
	}
}

// GetClient 获取指定客户端
func (m *ClientManager) GetClient(clientID string) (*Client, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	client, exists := m.clients[clientID]
	return client, exists
}

// GetAllClients 获取所有客户端
func (m *ClientManager) GetAllClients() []*Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	clients := make([]*Client, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	return clients
}

// CloseAll 关闭所有客户端连接
func (m *ClientManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, client := range m.clients {
		client.Close()
	}
	m.clients = make(map[string]*Client)
}

func (m *ClientManager) StartAll() {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, client := range m.clients {
		go client.Daemon()
	}
}

// Broadcast 向所有连接的客户端广播消息
func (m *ClientManager) Broadcast() {
	for payload := range protocol.BroadcastChan {
		for _, client := range m.clients {
			if client.Connected {
				payload.MessageBase.SelfId = client.XSelfID
				jsonData, err := json.Marshal(payload)
				if err != nil {
					continue
				}
				client.Conn.SendText(string(jsonData))
			}
		}
	}
}
