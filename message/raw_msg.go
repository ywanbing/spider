package message

import (
	"github.com/ywanbing/spider/code"
	"github.com/ywanbing/spider/codec"
)

// RawMessage msg 实现
// msgSize = 4
// msgIDSize = 4
// protoType = 1
// headerSize = 3
// metadata
// bodySize = 4
// body
type RawMessage struct {
	msgId       uint32
	marshalType codec.MarshalType

	metadata map[string]string
	body     []byte
}

var (
	_ Message = new(RawMessage)
)

func (m *RawMessage) GetMarshalType() codec.MarshalType {
	return m.marshalType
}

func (m *RawMessage) GetHeader() map[string]string {
	return m.metadata
}

func (m *RawMessage) SetHeader(k string, v string) {
	m.metadata[k] = v
}

func (m *RawMessage) GetBody() []byte {
	return m.body
}

func (m *RawMessage) SetBody(body []byte) {
	m.body = body
}

func (m *RawMessage) GetMsgId() uint32 {
	if m == nil {
		return 0
	}
	return m.msgId
}

func (m *RawMessage) Check() error {
	if m == nil {
		return code.ErrNilMessage
	}

	if m.metadata == nil {
		return code.ErrNilMetadata
	}

	if m.metadata[MsgTypeKey] == "" {
		return code.ErrNilMsgType
	}

	// 如果是请求消息，那么必须要有Seq
	if m.metadata[MsgTypeKey] == MsgTypeRequest.String() && m.metadata[MsgSeq] == "" {
		return code.ErrNilMsgSeq
	}

	return nil
}

func NewMessage(msgId uint32, protoType codec.MarshalType, metadata map[string]string, body []byte) *RawMessage {
	return &RawMessage{msgId: msgId, marshalType: protoType, metadata: metadata, body: body}
}

func NewMsgWithMsgID(msgId uint32) *RawMessage {
	return &RawMessage{
		msgId:    msgId,
		metadata: make(map[string]string),
	}
}
