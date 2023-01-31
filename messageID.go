package spider

type MsgID interface {
	GetModelId() int32
	GetSubMsgId() int32
}
