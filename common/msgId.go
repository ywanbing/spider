package common

func GetModelId(msgId uint32) int32 {
	if msgId == 0 {
		return 0
	}
	return int32(msgId >> 16)
}

func GetSubMsgId(msgId uint32) int32 {
	if msgId == 0 {
		return 0
	}
	return int32(msgId & 0xFFFF)
}

func NewMsgIdWithSubMsgID(modelID, subMsgID int32) uint32 {
	return uint32(modelID<<16 | subMsgID)
}
