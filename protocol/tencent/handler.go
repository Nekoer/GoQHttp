package tencent

import (
	"GoQHttp/logger"
	"GoQHttp/protocol/tencent/dto"
	"GoQHttp/websocket"
	"encoding/json"
	"time"
)

func GroupAddRobotEventHandler(event *dto.Payload, data *dto.GroupAddRobotDataEvent) error {
	logger.Infof("收到群添加机器人事件, 群号:%v", data.GroupOpenId)
	return nil
}

func GroupDelRobotEventHandler(event *dto.Payload, data *dto.GroupDelRobotDataEvent) error {
	logger.Infof("收到群删除机器人事件, 群号:%v", data.GroupOpenId)
	return nil
}

func GroupMsgRejectEventHandler(event *dto.Payload, data *dto.GroupMsgRejectDataEvent) error {
	logger.Infof("收到群聊拒绝机器人主动消息事件, 群号:%v", data.GroupOpenId)
	return nil
}

func GroupMsgReceiveEventHandler(event *dto.Payload, data *dto.GroupMsgReceiveDataEvent) error {
	logger.Infof("收到群聊接受机器人主动消息事件, 群号:%v", data.GroupOpenId)
	return nil
}

func GroupAtMessageEventHandler(event *dto.Payload, data *dto.GroupATMessageDataEvent) error {

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		logger.Infof("群消息: %+v", data.Content)
	} else {
		logger.Infof("群消息: %+v", string(jsonBytes))
	}

	for _, client := range websocket.Manager.GetAllClients() {
		messageResponse := websocket.MessageResponse{
			MessageBase: websocket.MessageBase{
				Time:     time.Now().Unix(),
				SelfId:   client.XSelfID,
				PostType: websocket.MessagePost,
			},
			MessageType: websocket.GroupMessage,
			SubType:     websocket.Group,
			MessageId:   0,
			GroupId:     0,
			UserId:      0,
			Message:     []any{},
			RawMessage:  data.Content,
			Font:        0,
			Sender: websocket.Sender{
				UserId:   0,
				NickName: "",
				Sex:      "",
				Age:      0,
				Card:     "",
				Area:     "",
				Level:    "",
				Role:     "",
				Title:    "",
			},
		}

		jsonData, err := json.Marshal(messageResponse)
		if err != nil {
			return err
		}

		client.SendMessage(string(jsonData))
	}
	return nil
}

func GroupMessageEventHandler(event *dto.Payload, data *dto.GroupMessageDataEvent) error {
	return nil
}

// GuildEventHandler 频道事件handler
func GuildEventHandler(event *dto.Payload, data *dto.GuildDataEvent) error {

	return nil
}

// GuildMemberEventHandler 频道成员事件 handler
func GuildMemberEventHandler(event *dto.Payload, data *dto.GuildMemberDataEvent) error {
	return nil
}

// ChannelEventHandler 子频道事件 handler
func ChannelEventHandler(event *dto.Payload, data *dto.ChannelDataEvent) error {
	return nil
}

// MessageEventHandler 消息事件 handler
func MessageEventHandler(event *dto.Payload, data *dto.MessageDataEvent) error {
	return nil
}

// MessageDeleteEventHandler 消息事件 handler
func MessageDeleteEventHandler(event *dto.Payload, data *dto.MessageDeleteDataEvent) error {
	return nil
}

// PublicMessageDeleteEventHandler 消息事件 handler
func PublicMessageDeleteEventHandler(event *dto.Payload, data *dto.PublicMessageDeleteDataEvent) error {
	return nil
}

// DirectMessageDeleteEventHandler 消息事件 handler
func DirectMessageDeleteEventHandler(event *dto.Payload, data *dto.DirectMessageDeleteDataEvent) error {
	return nil
}

// MessageReactionEventHandler 表情表态事件 handler
func MessageReactionEventHandler(event *dto.Payload, data *dto.MessageReactionDataEvent) error {
	return nil
}

// ATMessageEventHandler at 机器人消息事件 handler
func ATMessageEventHandler(event *dto.Payload, data *dto.ATMessageDataEvent) error {
	return nil
}

// DirectMessageEventHandler 私信消息事件 handler
func DirectMessageEventHandler(event *dto.Payload, data *dto.DirectMessageDataEvent) error {
	return nil
}

// AudioEventHandler 音频机器人事件 handler
func AudioEventHandler(event *dto.Payload, data *dto.AudioDataEvent) error {
	return nil
}

// MessageAuditEventHandler 消息审核事件 handler
func MessageAuditEventHandler(event *dto.Payload, data *dto.MessageAuditDataEvent) error {
	return nil
}

// ThreadEventHandler 论坛主题事件 handler
func ThreadEventHandler(event *dto.Payload, data *dto.ThreadDataEvent) error {
	return nil
}

// PostEventHandler 论坛回帖事件 handler
func PostEventHandler(event *dto.Payload, data *dto.PostDataEvent) error {
	return nil
}

// ReplyEventHandler 论坛帖子回复事件 handler
func ReplyEventHandler(event *dto.Payload, data *dto.ReplyDataEvent) error {
	return nil
}

// ForumAuditEventHandler 论坛帖子审核事件 handler
func ForumAuditEventHandler(event *dto.Payload, data *dto.ForumAuditDataEvent) error {
	return nil
}

// InteractionEventHandler 互动事件 handler
func InteractionEventHandler(event *dto.Payload, data *dto.InteractionDataEvent) error {
	return nil
}

func C2CMessageEventHandler(event *dto.Payload, data *dto.C2CMessageDataEvent) error {
	return nil
}

func FriendAddEventHandler(event *dto.Payload, data *dto.FriendAddDataEvent) error {
	return nil
}

func FriendDelEventHandler(event *dto.Payload, data *dto.FriendDelDataEvent) error {
	return nil
}

func C2CMsgRejectHandler(event *dto.Payload, data *dto.FriendMsgRejectDataEvent) error {
	return nil
}

func C2CMsgReceiveHandler(event *dto.Payload, data *dto.FriendMsgReceiveDataEvent) error {
	return nil
}
