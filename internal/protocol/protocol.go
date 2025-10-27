package protocol

import (
	"GoQHttp/config"
	"GoQHttp/internal/onebot"
	"GoQHttp/internal/protocol/tencent/dto"
	"net/http"
)

// QQClientChan QQ官方下发频道
var QQClientChan chan *dto.Payload = make(chan *dto.Payload, 100)

var KookChan chan interface{} = make(chan interface{}, 100)

// QQOfficialChan 发送给QQ官方频道
var QQOfficialChan chan *onebot.MessageRequest = make(chan *onebot.MessageRequest, 100)

// BroadcastChan 反向WS 广播频道
var BroadcastChan chan onebot.MessageRequest = make(chan onebot.MessageRequest, 100)

type Protocol interface {
	Init(w http.ResponseWriter, r *http.Request, config *config.Config)
}
