package codec

import "encoding/json"

const (
	MarshalType_Json MarshalType = 'J'
)

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

func init() {
	_ = RegisterMarshaller(JsonMarshaller{})
}
