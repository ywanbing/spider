package proto

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/ywanbing/spider/codec"
	"github.com/ywanbing/spider/message"
)

const (
	MsgSize   = 4
	MsgIDSize = 4

	ProtoSize    = 1
	MetadataSize = 3

	BodySize = 4
	AllSize  = MsgSize + MsgIDSize + MetadataSize + ProtoSize + BodySize
)

// RawProto 提供默认的Proto实现
type RawProto struct{}

// NewRawProto 创建一个 RawProto, 并且初始化缓存池
func NewRawProto() *RawProto {
	return &RawProto{}
}

func (r RawProto) Pack(m message.Message) ([]byte, error) {
	// 元数据默认为 json 序列化
	meatData, _ := json.Marshal(m.GetHeader())
	meatDataLen := len(meatData)
	if meatDataLen > 0x0fff {
		return nil, fmt.Errorf("metadata is too long")
	}

	body := m.GetBody()
	bodyLen := len(body)

	allSize := AllSize + meatDataLen + bodyLen
	data := make([]byte, allSize)

	// 1. 写入消息长度
	binary.BigEndian.PutUint32(data[:4], uint32(allSize))
	// 2. 写入消息id
	binary.BigEndian.PutUint32(data[4:8], m.GetMsgId())
	// 3. 写入序列化类型和头部长度[protoType = 1b, meatDataLen = 3b]
	binary.BigEndian.PutUint32(data[8:12], uint32(m.GetMarshalType())<<24|uint32(meatDataLen))
	// 4. 写入元数据
	copy(data[12:12+meatDataLen], meatData)
	// 5. 写入消息体长度
	binary.BigEndian.PutUint32(data[12+meatDataLen:16+meatDataLen], uint32(bodyLen))
	// 6. 写入消息体
	copy(data[16+meatDataLen:], body)

	return data, nil
}

func (r RawProto) Unpack(data []byte) (message.Message, error) {
	msgId := binary.BigEndian.Uint32(data[:4])
	protoTypeAndMeatSize := binary.BigEndian.Uint32(data[4:8])
	protoType := codec.MarshalType(protoTypeAndMeatSize >> 24)
	meatDataLen := protoTypeAndMeatSize & 0x0fff
	meatData := data[8 : 8+meatDataLen]
	bodyDataLen := binary.BigEndian.Uint32(data[8+meatDataLen : 12+meatDataLen])

	// 结束引用
	bodyData := make([]byte, bodyDataLen)
	copy(bodyData, data[12+meatDataLen:])

	// 1. 解析元数据
	meat := make(map[string]string)
	if meatDataLen > 0 {
		err := json.Unmarshal(meatData, &meat)
		if err != nil {
			return nil, err
		}
	}

	m := message.NewMessage(msgId, protoType, meat, bodyData)
	return m, nil
}
