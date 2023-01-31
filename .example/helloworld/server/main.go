package main

import (
	"fmt"
	"github.com/ywanbing/spider"
)

func main() {
	tcpX := spider.NewTcpX()

	tcpX.RegisterGlobalMiddle(func(ctx *spider.Context) {
		defer func() {
			if err := recover(); err != nil {
				// do something
			}
		}()

		fmt.Println("global middle")

		ctx.Next()
	})

	tcpX.RegisterHandler(1, 1, func(ctx *spider.Context) {
		// do something
		data := ctx.RawData()
		fmt.Println(string(data))

		ctx.JSON(2, []byte("hello world"))
	})

	tcpX.ListenAndServe("tcp", ":8089")
}
