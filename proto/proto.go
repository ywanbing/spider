package proto

import (
	"github.com/ywanbing/spider/message"
)

type Proto interface {
	// Pack writes the Message into the connection.
	Pack(message.Message) ([]byte, error)
	// Unpack reads bytes from the connection to the Message.
	Unpack([]byte) (message.Message, error)
}
