package spider

import (
	"context"

	"github.com/ywanbing/spider/codec"
	"github.com/ywanbing/spider/message"
)

const ABORT int8 = 100

type Context struct {
	// context.Context
	ctx context.Context

	reqMsg message.Message

	// 当前的连接对象
	conn TcpConn
	t    *TcpServer

	// used to control middleware abort or next
	// offset == ABORT, abort
	// else next
	offset   int8
	handlers []func(*Context)
}

func NewContext(ctx context.Context, reqMsg message.Message, conn TcpConn) *Context {
	return &Context{
		ctx:    ctx,
		reqMsg: reqMsg,
		conn:   conn,
		offset: -1,
	}
}

// Next Since middlewares are divided into 3 kinds: global, messageIDSelfRelated, anchorType,
// offset can't be used straightly to control middlewares like  middlewares[offset]().
// Thus, c.Next() means actually do nothing.
func (c *Context) Next() {
	c.offset++
	s := len(c.handlers)
	for ; int(c.offset) < s; c.offset++ {
		if !c.isAbort() {
			c.handlers[c.offset](c)
		} else {
			return
		}
	}
}

func (c *Context) isAbort() bool {
	if c.offset >= ABORT {
		return true
	}
	return false
}

// Abort stop middleware chain
func (c *Context) Abort() {
	c.offset = ABORT
}

// JSON Reply to client using json marshaller.
// Whatever ctx.Packx.Marshaller.MarshalName is 'json' or not , message block will marshal its header and body by json marshaller.
func (c *Context) JSON(msgId uint32, src interface{}, meatData ...map[string]any) error {
	return c.commonReplyWithMarshaller(codec.JsonMarshaller{}, msgId, src, meatData...)
}

// ProtoBuf Reply to client using protobuf marshaller.
// Whatever ctx.Packx.Marshaller.MarshalName is 'protobuf' or not , message block will marshal its header and body by protobuf marshaller.
func (c *Context) ProtoBuf(msgId uint32, src interface{}, meatData ...map[string]any) error {
	return c.commonReplyWithMarshaller(codec.ProtobufMarshaller{}, msgId, src, meatData...)
}

// Raw Reply to client using protobuf marshaller.
// Whatever ctx.Packx.Marshaller.MarshalName is 'protobuf' or not , message block will marshal its header and body by protobuf marshaller.
func (c *Context) Raw(msgId uint32, src interface{}, meatData ...map[string]any) error {
	return c.commonReplyWithMarshaller(codec.RawMarshaller{}, msgId, src, meatData...)
}

func (c *Context) commonReplyWithMarshaller(marshaller codec.Marshaller, msgId uint32, src any, meatData ...map[string]any) error {
	bytes, err := marshaller.Marshal(src)
	if err != nil {
		return err
	}

	// 默认第一个为metadata
	md := c.reqMsg.GetHeader()
	if len(meatData) > 0 {
		for k, v := range meatData[0] {
			md[k] = v
		}
	}

	md[message.MsgTypeKey] = message.MsgTypeReply
	return c.conn.SendMsg(message.NewMessage(msgId, marshaller.MarshalType(), md, bytes))
}

// Bind 自动反序列化
func (c *Context) Bind(dest any) error {
	return codec.GetMarshallerByMarshalType(c.reqMsg.GetMarshalType()).Unmarshal(c.reqMsg.GetBody(), dest)
}

// RawData 获取原始数据，不做任何解析，请根据MarshallerType 配合使用
func (c *Context) RawData() []byte {
	return c.reqMsg.GetBody()
}

// MarshallerType 获取解析类型，请配合RawData使用
func (c *Context) MarshallerType() codec.MarshalType {
	return c.reqMsg.GetMarshalType()
}
