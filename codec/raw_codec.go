package codec

const (
	MarshalType_Raw MarshalType = 'R'
)

type RawMarshaller struct{}

func (rw RawMarshaller) Marshal(v any) ([]byte, error) {
	return v.([]byte), nil
}
func (rw RawMarshaller) Unmarshal(data []byte, dest any) error {
	dest = data
	return nil
}

func (rw RawMarshaller) MarshalType() MarshalType {
	return MarshalType_Raw
}

func init() {
	_ = RegisterMarshaller(RawMarshaller{})
}
