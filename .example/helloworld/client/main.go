package main

import (
	"time"

	"github.com/ywanbing/spider"
	"github.com/ywanbing/spider/codec"
	"github.com/ywanbing/spider/common"
	"github.com/ywanbing/spider/message"
)

func main() {

	// Create a new client
	client := spider.NewTcpClient(":8089")
	client.Start()

	// TODO
	msg := message.NewMessage(common.NewMsgIdWithSubMsgID(1, 1), codec.MarshalType_Json, map[string]any{
		message.MsgTypeKey: message.MsgTypeRequest,
		message.MsgSeq:     1,
	}, []byte("hello world"))
	client.SendMsg(msg)

	time.Sleep(time.Second * 5)
	client.Close()
}
