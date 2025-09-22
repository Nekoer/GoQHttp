package tencent

import (
	"GoQHttp/constant"
	"GoQHttp/logger"
	"GoQHttp/onebot"
	"GoQHttp/protocol"
	"GoQHttp/protocol/tencent/dto"
	"GoQHttp/utils"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// FaceTag 定义结构体来存储解析后的标签信息
type FaceTag struct {
	Type int
	ID   string
	Ext  string
	Raw  string
}

// 辅助函数：将字符串转换为整数
func parseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}

// SplitByTags 将字符串按照 <> 标签分割成数组
func SplitByTags(input string) []string {
	input = strings.TrimPrefix(input, " ")
	re := regexp.MustCompile(`(<[^>]+>|[^<]+)`)
	matches := re.FindAllString(input, -1)

	// 过滤空值
	result := make([]string, 0)
	for _, match := range matches {
		if match != "" {
			result = append(result, match)
		}
	}

	return result
}

// 识别并解析 Face 标签
func parseFaceTags(input string) bool {
	// 正则表达式匹配特定的 face 标签格式
	// 匹配模式: <faceType=数字,faceId="值",ext="base64字符串">
	pattern := `<faceType=(\d+),faceId="([^"]+)",ext="([^"]+)">`
	re := regexp.MustCompile(pattern)

	matches := re.FindAllStringSubmatch(input, -1)

	var tags []FaceTag
	for _, match := range matches {
		if len(match) < 4 {
			continue
		}

		tag := FaceTag{
			Type: parseInt(match[1]),
			ID:   match[2],
			Ext:  match[3],
			Raw:  match[0],
		}
		tags = append(tags, tag)
	}

	return len(tags) > 0
}

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

	GroupId, err := utils.DBUtil.GroupInsert(data.GroupId, data.GroupOpenId)
	if err != nil {
		logger.Warnf("群维护失败: %v", err)
		return err
	}

	SenderId, err := utils.DBUtil.SenderInsert(data.Author.UserOpenId, data.Author.UserId)
	if err != nil {
		logger.Warnf("群用户维护失败: %v", err)
		return err
	}

	MessageId, err := utils.DBUtil.GroupMessageInsert(data.MsgId, GroupId, SenderId)
	if err != nil {
		logger.Warnf("群消息维护失败: %v", err)
		return err
	}

	tmpMessage := SplitByTags(data.Content)
	// 构建消息体
	var messages []*onebot.Element
	var rawMessages []string
	var imageIndex = 0
	for _, message := range tmpMessage {
		if parseFaceTags(message) {
			image := onebot.Image{
				File: data.Attachments[imageIndex].URL,
				Url:  data.Attachments[imageIndex].URL,
			}
			messages = append(messages, &onebot.Element{
				ElementType: onebot.ImageType,
				Data:        image,
			})
			imageIndex += 1
			rawMessages = append(rawMessages, image.String())
		} else {
			messages = append(messages, &onebot.Element{
				ElementType: onebot.TextType,
				Data: &onebot.Text{
					Text: message,
				},
			})
			rawMessages = append(rawMessages, message)
		}
	}
	rawMessage := strings.Join(rawMessages, "")

	messageRequest := onebot.MessageRequest{
		MessageBase: onebot.MessageBase{
			Time:     time.Now().Unix(),
			SelfId:   1, // 会在发送的时候根据websocket内部的XSelfId进行覆盖
			PostType: onebot.MessagePost,
		},
		MessageType:     onebot.GroupMessage,
		SubType:         onebot.Normal,
		MessageId:       MessageId,
		GroupId:         GroupId,
		UserId:          SenderId,
		OriginalMessage: messages,
		Message:         messages,
		RawMessage:      rawMessage,
		Font:            1,
		Sender: onebot.Sender{
			UserId:   SenderId,
			NickName: "UNKNOWN",
			Sex:      "未知",
			Age:      0,
			Card:     "",
			Area:     "",
			Level:    "",
			Role:     "",
			Title:    "",
		},
	}

	protocol.BroadcastChan <- messageRequest
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
	guild, err := constant.OpenApi.GetGuild(data.GuildID)
	if err != nil {
		return err
	}
	channel, err := constant.OpenApi.GetChannel(data.ChannelID)
	if err != nil {
		return err
	}
	rawMessage := data.Content
	attachments, err := constant.CQCode.BuildCQCodeFromAttachments(data.Attachments)
	if err != nil {
		return err
	}
	rawMessage += strings.Join(attachments, "")
	logger.Infof("[频道AT][%v][%v]%v:%v", guild.Name, channel.Name, data.Author.Username, rawMessage)
	return nil
}

// DirectMessageEventHandler 私信消息事件 handler
func DirectMessageEventHandler(event *dto.Payload, data *dto.DirectMessageDataEvent) error {
	rawMessage := data.Content
	attachments, err := constant.CQCode.BuildCQCodeFromAttachments(data.Attachments)
	if err != nil {
		return err
	}
	rawMessage += strings.Join(attachments, "")
	logger.Infof("[频道私聊]%v:%v", data.Author.Username, rawMessage)
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
