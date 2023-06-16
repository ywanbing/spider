package code

import "errors"

// 统一错误信息

var (
	ErrNilMessage  = Error("message is nil")
	ErrNilMetadata = Error("metadata is nil")
	ErrNilMsgType  = Error("msg type is empty")
	ErrNilMsgSeq   = Error("msg seq is empty")
)

func Error(s string) error {
	return errors.New(s)
}
