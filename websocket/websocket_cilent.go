package websocket

import (
	"GoQHttp/logger"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	logging "github.com/sacOO7/go-logger"
	"github.com/sacOO7/gowebsocket"
)

// Client 表示一个 WebSocket 客户端连接
type Client struct {
	ID            string
	URL           string
	Conn          gowebsocket.Socket
	mu            sync.Mutex
	Connected     bool
	Reconnect     bool
	MaxRetries    int
	RetryCount    int
	XSelfID       int64
	XClientRole   string
	Authorization string
	Interval      int64
}

// ClientManager 管理多个 WebSocket 客户端连接
type ClientManager struct {
	clients map[string]*Client
	mu      sync.RWMutex
}

//	meta_event_type: 'heartbeat',
//	interval: 5000,
//	post_type: 'meta_event',
//	self_id: this.bot.selfId,
//	time: Date.now(),
//	status: {
//	  app_initialized: true,
//	  app_enabled: true,
//	  plugins_good: true,
//	  app_good: true,
//	  online: true,
//	  good: true,
//	  stat: {
//	    packet_received: 0,
//	    packet_sent: 0,
//	    packet_lost: 0,
//	    message_received: 0,
//	    message_sent: 0,
//	    disconnect_times: 0,
//	    lost_times: 0,
//	    last_message_time: 0,
//	  },
//	}
//

type PostType string

const (
	MessagePost   PostType = "message"
	NoticePost    PostType = "notice"
	RequestPost   PostType = "request"
	MetaEventPost PostType = "meta_event"
)

// MessageBase 定义消息结构
type MessageBase struct {
	Time     int64    `json:"time"`
	SelfId   int64    `json:"self_id"`
	PostType PostType `json:"post_type"`
}

type SubType string

const (
	Enable  SubType = "enable"
	Disable SubType = "disable"
	Connect SubType = "connect"
	Friend  SubType = "friend"
	Group   SubType = "group"
	Other   SubType = "other"
)

type MetaEventType string

const (
	LifecycleType MetaEventType = "lifecycle"
	HeartbeatType MetaEventType = "heartbeat"
)

type LifeCycle struct {
	MessageBase
	MetaEventType MetaEventType `json:"meta_event_type"`
	SubType       SubType       `json:"sub_type"`
}

type Status struct {
	Online bool `json:"online"`
	Good   bool `json:"good"`
}

type Heartbeat struct {
	MessageBase
	MetaEventType MetaEventType `json:"meta_event_type"`
	Interval      int64         `json:"interval"`
	Status        Status        `json:"status"`
}

type MessageType string

const (
	PrivateMessage MessageType = "private"
	GroupMessage   MessageType = "group"
)

type Sex string

const (
	Male    Sex = "male"
	Female  Sex = "female"
	Unknown Sex = "unknown"
)

type Role string

const (
	Owner  Role = "Owner"
	Admin  Role = "Admin"
	Member Role = "Member"
)

type Sender struct {
	UserId   int64  `json:"user_id,omitempty"`
	NickName string `json:"nick_name,omitempty"`
	Sex      Sex    `json:"sex,omitempty"`
	Age      int32  `json:"age,omitempty"`
	Card     string `json:"card,omitempty"`
	Area     string `json:"area,omitempty"`
	Level    string `json:"level,omitempty"`
	Role     Role   `json:"role,omitempty"`
	Title    string `json:"title,omitempty"`
}
type Anonymous struct {
	Id   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Flag string `json:"flag,omitempty"`
}
type MessageResponse struct {
	MessageBase
	MessageType MessageType `json:"message_type"`
	SubType     SubType     `json:"sub_type"`
	MessageId   int32       `json:"message_id"`
	GroupId     int32       `json:"group_id"`
	UserId      int64       `json:"user_id"`
	Anonymous   Anonymous   `json:"anonymous,omitempty"`
	Message     []any       `json:"message"`
	RawMessage  string      `json:"raw_message"`
	Font        int32       `json:"font"`
	Sender      Sender      `json:"sender"`
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
func NewWebSocketClient(url string, xSelfID int64, xClientRole string, authorization string, interval int64) *Client {
	return &Client{
		ID:            generateID(),
		URL:           url,
		Connected:     false,
		Reconnect:     true,
		MaxRetries:    10,
		RetryCount:    0,
		XSelfID:       xSelfID,
		XClientRole:   xClientRole,
		Authorization: authorization,
		Interval:      interval,
	}
}

// generateID 生成唯一客户端 ID
func generateID() string {
	return fmt.Sprintf("client-%d", time.Now().UnixNano())
}

// Connect 连接到 WebSocket 服务器
func (c *Client) Connect() {
	for {
		if c.RetryCount >= c.MaxRetries && c.MaxRetries > 0 {
			logger.Warnf("达到最大重试次数 (%d)，停止连接: %s", c.MaxRetries, c.URL)
			return
		}

		logger.Infof("尝试连接到: %s", c.URL)

		// 解析 URL
		u, err := url.Parse(c.URL)
		if err != nil {
			logger.Warnf("URL 解析错误: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		header := http.Header{}
		header.Set("X-Self-ID", strconv.FormatInt(c.XSelfID, 10))
		header.Set("X-Client-Role", c.XClientRole)
		if c.Authorization != "" {
			header.Set("Authorization", "Bearer "+c.Authorization)
		}

		exit := make(chan struct{}, 1)

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
			c.Connected = true
			logger.Infof("已连接 %v", u.String())

		}
		c.Conn.OnConnectError = func(err error, socket gowebsocket.Socket) {
			logger.Errorf("connectErr:  %v", err.Error())
			exit <- struct{}{}
		}
		c.Conn.OnTextMessage = func(message string, socket gowebsocket.Socket) {
			c.handleMessage([]byte(message))
		}
		c.Conn.OnDisconnected = func(err error, socket gowebsocket.Socket) {
			c.Connected = false
			exit <- struct{}{}
			return
		}
		logger.Logger.Info("Connecting to platform server: " +
			c.URL + "...")
		c.Conn.Connect()
		c.LifeCycle()
		for {
			if !c.Connected {
				return
			}
			c.HeartBeat()
			time.Sleep(time.Second * time.Duration(5))
		}
	}
}

func (c *Client) LifeCycle() {
	c.mu.Lock()
	lifeCycle := LifeCycle{
		MessageBase: MessageBase{
			Time:     time.Now().Unix(),
			SelfId:   c.XSelfID,
			PostType: MetaEventPost,
		},
		MetaEventType: LifecycleType,
		SubType:       Connect,
	}
	msg, _ := json.Marshal(lifeCycle)
	c.Conn.SendText(string(msg))
	c.mu.Unlock()
}

func (c *Client) HeartBeat() {
	c.mu.Lock()
	heartbeat := Heartbeat{
		MessageBase: MessageBase{
			Time:     time.Now().Unix(),
			SelfId:   c.XSelfID,
			PostType: MetaEventPost,
		},
		MetaEventType: LifecycleType,
		Interval:      c.Interval,
		Status: Status{
			Online: true,
			Good:   false,
		},
	}
	msg, _ := json.Marshal(heartbeat)
	c.Conn.SendText(string(msg))
	c.mu.Unlock()
}

func (c *Client) SendMessage(data string) {
	c.Conn.SendText(data)
}

// handleMessage 处理接收到的消息
func (c *Client) handleMessage(message []byte) {
	var msg MessageResponse
	err := json.Unmarshal(message, &msg)
	if err != nil {
		logger.Warnf("无法解析消息: %v", err)
		return
	}

	// 根据消息类型处理
	switch msg.PostType {
	case MessagePost:
		logger.Infof("[%s] 消息事件: %s\n", c.URL, string(message))
	case NoticePost:
		logger.Infof("[%s] 通知事件: %v\n", c.URL, string(message))
	case RequestPost:
		logger.Infof("[%s] 请求事件: %v\n", c.URL, string(message))
	default:
		logger.Infof("[%s] 元事件: %v\n", c.URL, string(message))
	}
}

// SendMessage 发送消息到服务器
//func (c *Client) SendMessage(msgType string, content interface{}) error {
//	message := Message{
//		Type:      msgType,
//		Content:   content,
//		Timestamp: time.Now(),
//		Source:    c.ID,
//	}
//
//	msgBytes, err := json.Marshal(message)
//	if err != nil {
//		return err
//	}
//
//	select {
//	case c.SendChan <- msgBytes:
//		return nil
//	default:
//		return fmt.Errorf("发送通道已满")
//	}
//}

// Close 关闭连接
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Reconnect = false

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
