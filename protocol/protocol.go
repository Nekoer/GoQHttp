package protocol

import (
	"GoQHttp/config"
	"GoQHttp/onebot"
	"GoQHttp/protocol/tencent/dto"
	"net/http"
)

// ClientChan QQ官方下发频道
var ClientChan chan *dto.Payload = make(chan *dto.Payload, 100)

// OfficialChan 发送给QQ官方频道
var OfficialChan chan *onebot.MessageRequest = make(chan *onebot.MessageRequest, 100)

// BroadcastChan 反向WS 广播频道
var BroadcastChan chan onebot.MessageRequest = make(chan onebot.MessageRequest, 100)

type Protocol interface {
	Init(w http.ResponseWriter, r *http.Request, config *config.Config)
}
