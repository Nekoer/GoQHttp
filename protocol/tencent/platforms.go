package tencent

import "GoQHttp/protocol/tencent/dto"

var ClientChan chan *dto.Payload = make(chan *dto.Payload, 100)
