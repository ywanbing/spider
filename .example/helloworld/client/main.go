package main

import "github.com/ywanbing/spider"

func main() {

	// Create a new client
	client := spider.NewTcpClient(":8089")
	client.Start()

	// TODO

}
