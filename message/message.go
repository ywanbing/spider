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
}
