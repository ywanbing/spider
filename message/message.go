package message

import "github.com/ywanbing/spider/codec"

// Message 的接口定义
type Message interface {
	// GetMsgId 获取消息ID
	GetMsgId() uint32

	// GetMarshalType 获取消息的类型
	GetMarshalType() codec.MarshalType

	// GetHeader 获取消息的头部
	GetHeader() map[string]any

	// SetHeader 设置消息的头部
	SetHeader(k string, v any)

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
	MsgTypeKey = "msg_type"
	MsgSeq     = "msg_seq"
	MsgErr     = "msg_err"
	OpenTrance = "open_trance"
)
