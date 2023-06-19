package message

import "github.com/ywanbing/spider/codec"

// Message 的接口定义
type Message interface {
	// GetMsgId 获取消息ID
	GetMsgId() uint32

	// GetMarshalType 获取消息的类型
	GetMarshalType() codec.MarshalType

	// GetHeader 获取消息的头部
	GetHeader() map[string]string

	// SetHeader 设置消息的头部
	SetHeader(k string, v string)

	// GetBody 获取消息的内容
	GetBody() []byte

	// SetBody 设置消息的内容
	SetBody([]byte)

	// Check 自我检查
	Check() error
}

type MsgType int32

const (
	MsgTypeUnknown   MsgType = 0
	MsgTypeRequest   MsgType = 1
	MsgTypeReply     MsgType = 2
	MsgTypePush      MsgType = 3
	MsgTypeHeartBeat MsgType = 4
)

// 定义一些默认的消息头的Key
const (
	MsgTypeKey  = "msg_type"
	MsgSeq      = "msg_seq"
	MsgErr      = "msg_err"
	OpenTracing = "open_trace"
)

func (m MsgType) String() string {
	switch m {
	case MsgTypeRequest:
		return "request"
	case MsgTypeReply:
		return "reply"
	case MsgTypePush:
		return "push"
	case MsgTypeHeartBeat:
		return "heartbeat"
	default:
		return "unknown"
	}
}

// MsgTypeFromString 通过string 转换成 MsgType
func MsgTypeFromString(s string) MsgType {
	switch s {
	case "request":
		return MsgTypeRequest
	case "reply":
		return MsgTypeReply
	case "push":
		return MsgTypePush
	case "heartbeat":
		return MsgTypeHeartBeat
	default:
		return MsgTypeUnknown
	}
}
