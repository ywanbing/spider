package proto

import (
	"fmt"

	"github.com/ywanbing/spider/message"
)

type GzipProto struct {
	*RawProto
}

func NewGzipProto() *GzipProto {
	return &GzipProto{
		RawProto: NewRawProto(),
	}
}

// Pack 采用gzip压缩body得数据
func (g GzipProto) Pack(m message.Message) ([]byte, error) {
	body := m.GetBody()
	bodyLen := len(body)
	if bodyLen == 0 {
		return nil, fmt.Errorf("body is empty")
	}

	// TODO  gzip 压缩

	m.SetBody(body)
	return g.RawProto.Pack(m)
}

func (g GzipProto) Unpack(data []byte) (message.Message, error) {
	m, err := g.RawProto.Unpack(data)
	if err != nil {
		return nil, err
	}
	body := m.GetBody()

	// TODO gzip 解压缩

	m.SetBody(body)

	return m, nil
}
