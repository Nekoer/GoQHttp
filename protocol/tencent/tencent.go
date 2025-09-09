package tencent

import (
	"GoQHttp/config"
	"GoQHttp/logger"
	"GoQHttp/protocol/tencent/dto"
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Tencent struct {
}

type ValidationRequest struct {
	PlainToken string `json:"plain_token"`
	EventTs    string `json:"event_ts"`
}
type ValidationResponse struct {
	PlainToken string `json:"plain_token"`
	Signature  string `json:"signature"`
}

func handleValidation(payload *dto.Payload, config *config.Config) ([]byte, error) {
	validationPayload := &ValidationRequest{}
	b, _ := json.Marshal(payload.Data)
	if err := json.Unmarshal(b, validationPayload); err != nil {
		logger.Warnf("parse http payload failed: %s", err)
		return nil, err
	}

	seed := config.Bot.QQ.Secret
	for len(seed) < ed25519.SeedSize {
		seed = strings.Repeat(seed, 2)
	}
	seed = seed[:ed25519.SeedSize]
	reader := strings.NewReader(seed)
	// GenerateKey 方法会返回公钥、私钥，这里只需要私钥进行签名生成不需要返回公钥
	_, privateKey, err := ed25519.GenerateKey(reader)
	if err != nil {
		logger.Warnf("ed25519 generate key failed: %s", err)
		return nil, err
	}
	var msg bytes.Buffer
	msg.WriteString(validationPayload.EventTs)
	msg.WriteString(validationPayload.PlainToken)
	signature := hex.EncodeToString(ed25519.Sign(privateKey, msg.Bytes()))

	rspBytes, err := json.Marshal(
		&ValidationResponse{
			PlainToken: validationPayload.PlainToken,
			Signature:  signature,
		})
	if err != nil {
		logger.Warnf("handle validation failed: %s", err)
		return nil, err
	}
	return rspBytes, nil
}

func verifySignature(signature string, timestamp string, body []byte, config *config.Config) bool {
	seed := config.Bot.QQ.Secret
	for len(seed) < ed25519.SeedSize {
		seed = strings.Repeat(seed, 2)
	}
	rand := strings.NewReader(seed[:ed25519.SeedSize])
	publicKey, _, err := ed25519.GenerateKey(rand)

	if err != nil {
		logger.Warnf("ed25519 generate key failed: %s", err)
		return false
	}

	if signature == "" {
		logger.Warn("signature: NULL")
		return false
	}
	sig, err := hex.DecodeString(signature)
	if err != nil {
		logger.Warnf("signature decode failed: %s", err)
		return false
	}
	if len(sig) != ed25519.SignatureSize || sig[63]&224 != 0 {
		logger.Warnf("check signature failed: signature invalid")
		return false
	}

	if timestamp == "" {
		logger.Warnf("signature timestamp: NULL")
		return false
	}

	// 按照timestamp+Body顺序组成签名体
	var msg bytes.Buffer
	msg.Write([]byte(timestamp))
	msg.Write(body)

	return ed25519.Verify(publicKey, msg.Bytes(), sig)
}

func (qq *Tencent) Init(w http.ResponseWriter, r *http.Request, config *config.Config) {
	startTime := time.Now()

	// 验证内容类型
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		sendErrorResponse(w, "不支持的媒体类型", http.StatusUnsupportedMediaType)
		return
	}

	//userAgent := r.Header.Get("User-Agent")
	signature := r.Header.Get("X-Signature-Ed25519")
	timestamp := r.Header.Get("X-Signature-Timestamp")
	appid := r.Header.Get("X-Bot-Appid")
	if appid != strconv.Itoa(config.Bot.QQ.Id) {
		sendErrorResponse(w, "appid不相符", http.StatusUnauthorized)
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendErrorResponse(w, "无法读取请求体", http.StatusBadRequest)
		return
	}

	defer func() {
		if r.Body != nil {
			_ = r.Body.Close()
		}
	}()

	payload := &dto.Payload{}
	if err = json.Unmarshal(body, payload); err != nil {
		logger.Warnf("parse http payload err %s", err)
		return
	}

	// 记录接收到的 webhook
	logger.Debugf("收到 Webhook 事件: %s", payload.Type)
	switch payload.OPCode {
	case 0:
		{
			isValid := verifySignature(signature, timestamp, body, config)
			logger.Debugf("消息验签: %v", isValid)
			// 根据事件类型处理
			if !isValid {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("true"))
			} else {
				ClientChan <- payload
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
			}
		}
	case 13:
		{
			//	验证签名
			result, err := handleValidation(payload, config)
			if err != nil {
				sendErrorResponse(w, err.Error(), http.StatusBadRequest)
			} else {
				_, _ = w.Write(result)
			}

		}
	default:
		logger.Warnf("未处理操作: %s", string(body))
	}

	// 记录处理时间
	duration := time.Since(startTime)
	logger.Debugf("请求处理完成: %s %s - %v", r.Method, r.URL.Path, duration)
}

// sendErrorResponse 发送错误响应
func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	logger.Infof("错误: %s (状态码: %d)", message, statusCode)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]interface{}{
		"status":  "error",
		"message": message,
	}
	_ = json.NewEncoder(w).Encode(response)
}

// HandlerEvent 根据事件类型处理 webhook
func HandlerEvent() {
	for payload := range ClientChan {
		message, _ := json.Marshal(payload.Data)
		//{
		//"op":0,
		//"id":"GROUP_AT_MESSAGE_CREATE:yit7pdisg1ugi2ejzlc2ftxchdbcvuwyprx6jusrlcb3ifop7ft9skolij8yha",
		//"d":
		//{
		//"id":"ROBOT1.0_yiT7pd.iSG1ugi2eJZLC-4hjLnBZLEexTLIZboQEsNgdqRNv9EV2xw0aYNn76KurHB.MGL6iFrTqL73j.uv-O8U1msiTe5JeJMfw9nzdrjc!",
		//"content":" lssv",
		//"timestamp":"2025-09-08T23:13:34+08:00",
		//"author":{"id":"0BA13D42931777D3922F226B6B5D7F13","member_openid":"0BA13D42931777D3922F226B6B5D7F13","union_openid":"0BA13D42931777D3922F226B6B5D7F13"},
		//"group_id":"82B3406CB35C23E1071C8D6EC61C6064",
		//"group_openid":"82B3406CB35C23E1071C8D6EC61C6064",
		//"message_scene":{"source":"default"},
		//"message_type":0
		//},
		//"t":"GROUP_AT_MESSAGE_CREATE"}

		//logger.Infof("处理事件: %+v", payload)

		switch payload.Type {
		case dto.EventGroupATMessageCreate:
			{
				gm := &dto.GroupATMessageDataEvent{}
				err := json.Unmarshal(message, gm)
				if err == nil {
					_ = GroupAtMessageEventHandler(payload, gm)
				}
			}
		case dto.EventGroupAddRobbot:
			{
				gar := &dto.GroupAddRobotDataEvent{}
				err := json.Unmarshal(message, gar)
				if err == nil {
					_ = GroupAddRobotEventHandler(payload, gar)
				}
			}
		case dto.EventGroupDelRobbot:
			{
				gdr := &dto.GroupDelRobotDataEvent{}
				err := json.Unmarshal(message, gdr)
				if err == nil {
					_ = GroupDelRobotEventHandler(payload, gdr)
				}
			}
		case dto.EventGroupMsgReceive:
			{
				gmr := &dto.GroupMsgReceiveDataEvent{}
				err := json.Unmarshal(message, gmr)
				if err == nil {
					_ = GroupMsgReceiveEventHandler(payload, gmr)
				}
			}
		case dto.EventGroupMsgReject:
			{
				gmr := &dto.GroupMsgRejectDataEvent{}
				err := json.Unmarshal(message, gmr)
				if err == nil {
					_ = GroupMsgRejectEventHandler(payload, gmr)
				}
			}
		case dto.EventC2CMessageCreate:
			{
				cmc := &dto.C2CMessageDataEvent{}
				err := json.Unmarshal(message, cmc)
				if err == nil {
					_ = C2CMessageEventHandler(payload, cmc)
				}
			}
		case dto.EventC2CMsgReceive:
			{
				fmr := &dto.FriendMsgReceiveDataEvent{}
				err := json.Unmarshal(message, fmr)
				if err == nil {
					_ = C2CMsgReceiveHandler(payload, fmr)
				}
			}
		case dto.EventC2CMsgReject:
			{
				fmr := &dto.FriendMsgRejectDataEvent{}
				err := json.Unmarshal(message, fmr)
				if err == nil {
					_ = C2CMsgRejectHandler(payload, fmr)
				}
			}
		case dto.EventFriendAdd:
			{
				fad := &dto.FriendAddDataEvent{}
				err := json.Unmarshal(message, fad)
				if err == nil {
					_ = FriendAddEventHandler(payload, fad)
				}
			}
		case dto.EventFriendDel:
			{
				fad := &dto.FriendDelDataEvent{}
				err := json.Unmarshal(message, fad)
				if err == nil {
					_ = FriendDelEventHandler(payload, fad)
				}
			}
		case dto.EventAtMessageCreate:
			{
				am := &dto.ATMessageDataEvent{}
				err := json.Unmarshal(message, am)
				if err == nil {
					_ = ATMessageEventHandler(payload, am)
				}
			}
		case dto.EventMessageCreate:
			{
				m := &dto.MessageDataEvent{}
				err := json.Unmarshal(message, m)
				if err == nil {
					_ = MessageEventHandler(payload, m)
				}
			}
		case dto.EventInteractionCreate:
			{
				i := &dto.InteractionDataEvent{}
				err := json.Unmarshal(message, i)
				if err == nil {
					i.ID = payload.ID
					_ = InteractionEventHandler(payload, i)
				}
			}
		case dto.EventDirectMessageCreate:
			{
				i := &dto.DirectMessageDataEvent{}
				err := json.Unmarshal(message, i)
				if err == nil {
					_ = DirectMessageEventHandler(payload, i)
				}
			}
		case dto.EventMessageReactionRemove:
		case dto.EventMessageReactionAdd:
			{
				mr := &dto.MessageReactionDataEvent{}
				err := json.Unmarshal(message, mr)
				if err == nil {
					_ = MessageReactionEventHandler(payload, mr)
				}
			}
		case dto.EventMessageAuditReject:
		case dto.EventMessageAuditPass:
			{
				mr := &dto.MessageAuditDataEvent{}
				err := json.Unmarshal(message, mr)
				if err == nil {
					_ = MessageAuditEventHandler(payload, mr)
				}
			}
		case dto.EventForumReplyDelete:
		case dto.EventForumThreadDelete:
		case dto.EventForumPostDelete:
		case dto.EventForumThreadUpdate:
		case dto.EventForumReplyCreate:
		case dto.EventForumPostCreate:
		case dto.EventForumThreadCreate:
			{
				ft := &dto.ForumAuditDataEvent{}
				err := json.Unmarshal(message, ft)
				if err == nil {
					_ = ForumAuditEventHandler(payload, ft)
				}
			}
		case dto.EventGuildDelete:
		case dto.EventGuildUpdate:
		case dto.EventGuildCreate:
			{
				g := &dto.GuildDataEvent{}
				err := json.Unmarshal(message, g)
				if err == nil {
					_ = GuildEventHandler(payload, g)
				}
			}
		case dto.EventChannelDelete:
		case dto.EventChannelUpdate:
		case dto.EventChannelCreate:
			{
				c := &dto.ChannelDataEvent{}
				err := json.Unmarshal(message, c)
				if err == nil {
					_ = ChannelEventHandler(payload, c)
				}
			}
		case dto.EventGuildMemberUpdate:
		case dto.EventGuildMemberRemove:
		case dto.EventGuildMemberAdd:
			{
				gm := &dto.GuildMemberDataEvent{}
				err := json.Unmarshal(message, gm)
				if err == nil {
					_ = GuildMemberEventHandler(payload, gm)
				}
			}
		}
	}
}
