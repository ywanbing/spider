package message

import "github.com/ywanbing/spider/codec"

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

	metadata map[string]any
	body     []byte
}

var (
	_ Message = new(RawMessage)
)

func (m *RawMessage) GetMarshalType() codec.MarshalType {
	return m.marshalType
}

func (m *RawMessage) GetHeader() map[string]any {
	return m.metadata
}

func (m *RawMessage) SetHeader(k string, v any) {
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

func NewMessage(msgId uint32, protoType codec.MarshalType, metadata map[string]any, body []byte) *RawMessage {
	return &RawMessage{msgId: msgId, marshalType: protoType, metadata: metadata, body: body}
}

func NewMsgWithMsgID(msgId uint32) *RawMessage {
	return &RawMessage{
		msgId:    msgId,
		metadata: make(map[string]any),
	}
}
