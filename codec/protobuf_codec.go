package codec

import (
	"errors"

	"google.golang.org/protobuf/proto"
)

const (
	MarshalType_Proto MarshalType = 'P'
)

type ProtobufMarshaller struct{}

// Marshal v should realize proto.Message
func (pm ProtobufMarshaller) Marshal(v any) ([]byte, error) {
	src, ok := v.(proto.Message)
	if !ok {
		return nil, errors.New("protobuf marshaller requires src realize proto.Message")
	}
	return proto.Marshal(src)
}

// Unmarshal dest should realize proto.Message
func (pm ProtobufMarshaller) Unmarshal(data []byte, dest any) error {
	dst, ok := dest.(proto.Message)
	if !ok {
		return errors.New("protobuf marshaller requires src realize proto.Message")
	}
	return proto.Unmarshal(data, dst)
}

func (pm ProtobufMarshaller) MarshalType() MarshalType {
	return MarshalType_Proto
}

func init() {
	_ = RegisterMarshaller(ProtobufMarshaller{})
}
