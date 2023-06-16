package codec

import "errors"

type MarshalType byte

type Marshaller interface {
	Marshal(any) ([]byte, error)
	Unmarshal([]byte, any) error
	MarshalType() MarshalType
}

var (
	marshallerManager = make(map[MarshalType]Marshaller)
)

func RegisterMarshaller(m Marshaller) error {
	if _, ok := marshallerManager[m.MarshalType()]; ok {
		return errors.New("marshaller already registered")
	}

	marshallerManager[m.MarshalType()] = m
	return nil
}

func GetMarshallerByMarshalType(marshalType MarshalType) Marshaller {
	m, ok := marshallerManager[marshalType]
	if !ok {
		// 提供默认的 json marshaller
		return JsonMarshaller{}
	}
	return m
}
