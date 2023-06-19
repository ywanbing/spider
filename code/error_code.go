package code

import "errors"

// 统一错误信息

var (
	ErrNilMessage     = Error("message is nil")
	ErrNilMetadata    = Error("metadata is nil")
	ErrNilMsgType     = Error("msg type is empty")
	ErrNilMsgSeq      = Error("msg seq is empty")
	ErrConnClosed     = Error("connection is closed")
	ErrMessageNotSent = Error("message not sent")
)

func Error(s string) error {
	return errors.New(s)
}
