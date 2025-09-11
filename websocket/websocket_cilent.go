package websocket

import (
	"GoQHttp/logger"
	"GoQHttp/onebot"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
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
	TencentClient Tencent
}

// ClientManager 管理多个 WebSocket 客户端连接
type ClientManager struct {
	clients map[string]*Client
	mu      sync.RWMutex
}

type PostType string

const (
	MessagePost   PostType = "message"
	NoticePost    PostType = "notice"
	RequestPost   PostType = "request"
	MetaEventPost PostType = "meta_event"
)

// MessageBase 定义消息结构
type MessageBase struct {
	Time     int64    `json:"time,omitempty"`
	SelfId   int64    `json:"self_id,omitempty"`
	PostType PostType `json:"post_type,omitempty"`
}

type SubType string

const (
	Enable  SubType = "enable"
	Disable SubType = "disable"
	Connect SubType = "connect"
	Friend  SubType = "friend"
	Group   SubType = "group"
	Other   SubType = "other"
	Normal  SubType = "normal"
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
	UserId   int64  `json:"user_id"`
	NickName string `json:"nick_name"`
	Sex      Sex    `json:"sex"`
	Age      int32  `json:"age"`
	Card     string `json:"card"`
	Area     string `json:"area"`
	Level    string `json:"level"`
	Role     Role   `json:"role"`
	Title    string `json:"title"`
}
type Anonymous struct {
	Id   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Flag string `json:"flag,omitempty"`
}
type MessageRequest struct {
	MessageBase
	MessageType     MessageType       `json:"message_type"`
	SubType         SubType           `json:"sub_type,omitempty"`
	MessageId       int32             `json:"message_id,omitempty"`
	GroupId         int32             `json:"group_id"`
	UserId          int64             `json:"user_id"`
	Anonymous       *Anonymous        `json:"anonymous,omitempty"`
	OriginalMessage string            `json:"original_message,omitempty"`
	Message         []*onebot.Element `json:"message"`
	RawMessage      string            `json:"raw_message,omitempty"`
	Font            int32             `json:"font,omitempty"`
	Sender          Sender            `json:"sender,omitempty"`
	AutoEscape      bool              `json:"auto_escape,omitempty"`
}

type Request struct {
	Action string `json:"action"`
	Params any    `json:"params"`
	Echo   string `json:"echo"`
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
func NewWebSocketClient(url string, xSelfID int64, xClientRole string, authorization string, interval int64, sandbox bool, tencentClient Tencent, maxRetries int64, retryDelay int64) *Client {

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
		TencentClient: tencentClient,
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
		if c.MaxRetries == 0 {
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
	c.mu.Lock()
	logger.Debug(data)
	if c.Connected {
		c.Conn.SendText(data)
	}
	c.mu.Unlock()
}

func (c *Client) SendGroupMessage(data MessageRequest) {
	c.mu.Lock()
	c.TencentClient.SendGroupMessage(data)
	c.mu.Unlock()
}

// CQCode 表示一个CQ码
type CQCode struct {
	Type   string            // CQ码类型，如"at", "image"
	Params map[string]string // 参数字典
	Raw    string            // 原始字符串
}

// ParseCQCode 解析CQ码字符串，返回CQCode结构体
func ParseCQCode(s string) (*CQCode, error) {
	// 匹配CQ码的正则表达式
	pattern := `\[CQ:([^,\]]+)(?:,([^,\]]+=[^,\]]+))*\]`
	re := regexp.MustCompile(pattern)

	match := re.FindStringSubmatch(s)
	if match == nil {
		return nil, fmt.Errorf("无效的CQ码格式: %s", s)
	}

	cq := &CQCode{
		Type:   match[1],
		Params: make(map[string]string),
		Raw:    s,
	}

	// 解析参数
	if len(match) > 2 {
		paramStr := match[2]
		// 处理多个参数的情况
		params := strings.Split(paramStr, ",")
		for _, param := range params {
			parts := strings.SplitN(param, "=", 2)
			if len(parts) == 2 {
				key := parts[0]
				value := parts[1]
				cq.Params[key] = value
			}
		}
	}

	return cq, nil
}

// ParseAllCQCodes 从文本中提取并解析所有CQ码
func ParseAllCQCodes(text string) ([]*CQCode, error) {
	pattern := `\[CQ:[^\]]+\]`
	re := regexp.MustCompile(pattern)

	matches := re.FindAllString(text, -1)
	if matches == nil {
		return nil, fmt.Errorf("未找到CQ码")
	}

	var cqCodes []*CQCode
	for _, match := range matches {
		cq, err := ParseCQCode(match)
		if err != nil {
			return nil, err
		}
		cqCodes = append(cqCodes, cq)
	}

	return cqCodes, nil
}

// handleMessage 处理接收到的消息
func (c *Client) handleMessage(message []byte) {
	var request Request
	err := json.Unmarshal(message, &request)
	if err != nil {
		logger.Warnf("无法解析消息: %v", err)
		return
	}

	var messageRequest MessageRequest
	marshal, err := json.Marshal(request.Params)
	if err != nil {
		return
	}

	err = json.Unmarshal(marshal, &messageRequest)
	if err != nil {
		//messageRequest.Message
		var messages []*onebot.Element

		codes, err := ParseAllCQCodes(string(marshal))
		if err != nil {
			return
		}

		for _, code := range codes {
			//logger.Debugf("%+v,%+v,%+v", code.Raw, code.Type, code.Params)
			switch code.Type {
			case "text":
				{
					text := code.Params["text"]
					messages = append(messages, &onebot.Element{
						ElementType: onebot.TextType,
						Data: &onebot.Message{
							Text: text,
						},
					})
				}
			case "image":
				{
					file := code.Params["file"]
					messages = append(messages, &onebot.Element{
						ElementType: onebot.ImageType,
						Data: &onebot.Message{
							Image: onebot.Image{
								File: file,
							},
						},
					})
				}
			}
		}
		messageRequest.Message = messages
	}
	// 根据消息类型处理
	switch messageRequest.PostType {
	case NoticePost:
		logger.Infof("[%s] 通知事件: %v\n", c.URL, string(message))
	case RequestPost:
		logger.Infof("[%s] 请求事件: %v\n", c.URL, string(message))
	case MessagePost:
	default:
		c.SendGroupMessage(messageRequest)
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
