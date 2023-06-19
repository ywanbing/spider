package main

import (
	"context"

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
	msg := message.NewMessage(common.NewMsgIdWithSubMsgID(1, 1), codec.MarshalType_Raw, map[string]string{
		message.MsgTypeKey: message.MsgTypeRequest.String(),
	}, []byte("hello world"))

	resp, err := client.Call(context.Background(), msg)
	if err != nil {
		panic(err)
	}

	println(string(resp.GetBody()))

	client.Close()
}
