package spider

const (
	MsgSize   = 4
	MsgIDSize = 4

	ProtoSize    = 1
	MetadataSize = 3

	BodySize = 4
	AllSize  = MsgSize + MsgIDSize + MetadataSize + ProtoSize + BodySize
)

// msgSize = 4
// msgIDSize = 4
// protoType = 1
// headerSize = 3
// metadata
// bodySize = 4
// body
type message struct {
	msgId     uint32
	protoType MarshalType

	metadata map[string]any
	body     []byte
}

var (
	_ MsgID = new(message)
)

func (m *message) GetModelId() int32 {
	if m == nil {
		return 0
	}
	return int32(m.msgId >> 16)
}

func (m *message) GetSubMsgId() int32 {
	if m == nil {
		return 0
	}
	return int32(m.msgId & 0xFFFF)
}

func newMessage(msgId uint32, protoType MarshalType, metadata map[string]any, body []byte) *message {
	return &message{msgId: msgId, protoType: protoType, metadata: metadata, body: body}
}

func newMsgWithMsgID(msgId uint32) *message {
	return &message{
		msgId:    msgId,
		metadata: make(map[string]any),
	}
}

func newMsgWithSubMsgID(modelID, subMsgID int32) *message {
	return &message{
		msgId:    uint32(modelID)<<16 | uint32(subMsgID),
		metadata: make(map[string]any),
	}
}
