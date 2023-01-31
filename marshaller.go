package spider

import (
	"encoding/json"
	"errors"

	"google.golang.org/protobuf/proto"
)

type MarshalType byte

const (
	MarshalType_Json  MarshalType = 1
	MarshalType_Proto MarshalType = 2
)

type Marshaller interface {
	Marshal(any) ([]byte, error)
	Unmarshal([]byte, any) error
	MarshalType() MarshalType
}

func GetMarshallerByMarshalType(marshalType MarshalType) Marshaller {
	switch marshalType {
	case MarshalType_Json:
		return JsonMarshaller{}
	case MarshalType_Proto:
		return ProtobufMarshaller{}
	default:
		return JsonMarshaller{}
	}
}

type JsonMarshaller struct{}

func (js JsonMarshaller) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}
func (js JsonMarshaller) Unmarshal(data []byte, dest any) error {
	return json.Unmarshal(data, dest)
}

func (js JsonMarshaller) MarshalType() MarshalType {
	return MarshalType_Json
}

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
