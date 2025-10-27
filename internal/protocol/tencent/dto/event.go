package dto

type EventType string

const (
	EventGuildCreate           EventType = "GUILD_CREATE"
	EventGuildUpdate           EventType = "GUILD_UPDATE"
	EventGuildDelete           EventType = "GUILD_DELETE"
	EventChannelCreate         EventType = "CHANNEL_CREATE"
	EventChannelUpdate         EventType = "CHANNEL_UPDATE"
	EventChannelDelete         EventType = "CHANNEL_DELETE"
	EventGuildMemberAdd        EventType = "GUILD_MEMBER_ADD"
	EventGuildMemberUpdate     EventType = "GUILD_MEMBER_UPDATE"
	EventGuildMemberRemove     EventType = "GUILD_MEMBER_REMOVE"
	EventMessageCreate         EventType = "MESSAGE_CREATE"
	EventMessageReactionAdd    EventType = "MESSAGE_REACTION_ADD"
	EventMessageReactionRemove EventType = "MESSAGE_REACTION_REMOVE"
	EventAtMessageCreate       EventType = "AT_MESSAGE_CREATE"
	EventPublicMessageDelete   EventType = "PUBLIC_MESSAGE_DELETE"
	EventDirectMessageCreate   EventType = "DIRECT_MESSAGE_CREATE"
	EventDirectMessageDelete   EventType = "DIRECT_MESSAGE_DELETE"
	EventAudioStart            EventType = "AUDIO_START"
	EventAudioFinish           EventType = "AUDIO_FINISH"
	EventAudioOnMic            EventType = "AUDIO_ON_MIC"
	EventAudioOffMic           EventType = "AUDIO_OFF_MIC"
	EventMessageAuditPass      EventType = "MESSAGE_AUDIT_PASS"
	EventMessageAuditReject    EventType = "MESSAGE_AUDIT_REJECT"
	EventMessageDelete         EventType = "MESSAGE_DELETE"
	EventForumThreadCreate     EventType = "FORUM_THREAD_CREATE"
	EventForumThreadUpdate     EventType = "FORUM_THREAD_UPDATE"
	EventForumThreadDelete     EventType = "FORUM_THREAD_DELETE"
	EventForumPostCreate       EventType = "FORUM_POST_CREATE"
	EventForumPostDelete       EventType = "FORUM_POST_DELETE"
	EventForumReplyCreate      EventType = "FORUM_REPLY_CREATE"
	EventForumReplyDelete      EventType = "FORUM_REPLY_DELETE"
	EventForumAuditResult      EventType = "FORUM_PUBLISH_AUDIT_RESULT"
	EventInteractionCreate     EventType = "INTERACTION_CREATE"
	EventC2CMessageCreate      EventType = "C2C_MESSAGE_CREATE"
	EventGroupATMessageCreate  EventType = "GROUP_AT_MESSAGE_CREATE"
	EventGroupMessageCreate    EventType = "GROUP_MESSAGE_CREATE"
	EventGroupAddRobbot        EventType = "GROUP_ADD_ROBBOT"
	EventGroupDelRobbot        EventType = "GROUP_DEL_ROBBOT"
	EventGroupMsgReject        EventType = "GROUP_MSG_REJECT"
	EventGroupMsgReceive       EventType = "GROUP_MSG_RECEIVE"
	EventFriendAdd             EventType = "FRIEND_ADD"
	EventFriendDel             EventType = "FRIEND_DEL"
	EventC2CMsgReject          EventType = "C2C_MSG_REJECT"
	EventC2CMsgReceive         EventType = "C2C_MSG_RECEIVE"
)

// Payload websocket 消息结构
type Payload struct {
	PayloadBase
	Data       interface{} `json:"d,omitempty"`
	RawMessage []byte      `json:"-"` // 原始的 message 数据
}

// PayloadBase 基础消息结构，排除了 data
type PayloadBase struct {
	OPCode OPCode    `json:"op"`
	ID     string    `json:"id,omitempty"`
	Seq    uint32    `json:"s,omitempty"`
	Type   EventType `json:"t,omitempty"`
}

// GuildDataEvent 频道 payload
type GuildDataEvent Guild

// GuildMemberDataEvent 频道成员 payload
type GuildMemberDataEvent Member

// ChannelDataEvent 子频道 payload
type ChannelDataEvent Channel

// MessageDataEvent 消息 payload
type MessageDataEvent Message

// ATMessageDataEvent only at 机器人的消息 payload
type ATMessageDataEvent Message

// DirectMessageDataEvent 私信消息 payload
type DirectMessageDataEvent Message

type C2CMessageDataEvent C2CMessage

type GroupATMessageDataEvent GroupMessage

type GroupMessageDataEvent GroupMessage

// MessageDeleteDataEvent 消息 payload
type MessageDeleteDataEvent MessageDelete

// PublicMessageDeleteDataEvent 公域机器人的消息删除 payload
type PublicMessageDeleteDataEvent MessageDelete

// DirectMessageDeleteDataEvent 私信消息 payload
type DirectMessageDeleteDataEvent MessageDelete

// AudioDataEvent 音频机器人的音频流事件
type AudioDataEvent AudioAction

// MessageReactionDataEvent 表情表态事件
type MessageReactionDataEvent MessageReaction

// MessageAuditDataEvent 消息审核事件
type MessageAuditDataEvent MessageAudit

// ThreadDataEvent 主题事件
type ThreadDataEvent Thread

// PostDataEvent 帖子事件
type PostDataEvent Post

// ReplyDataEvent 帖子回复事件
type ReplyDataEvent Reply

// ForumAuditDataEvent 帖子审核事件
type ForumAuditDataEvent ForumAuditResult

// InteractionDataEvent 互动事件
type InteractionDataEvent Interaction

type GroupAddRobotDataEvent GroupAddRobotEvent

type GroupDelRobotDataEvent GroupDelRobotEvent

type GroupMsgRejectDataEvent GroupMsgRejectEvent

type GroupMsgReceiveDataEvent GroupMsgReceiveEvent

type FriendAddDataEvent FriendAddEvent

type FriendDelDataEvent FriendDelEvent

type FriendMsgRejectDataEvent FriendMsgRejectEvent

type FriendMsgReceiveDataEvent FriendMsgReceiveEvent
