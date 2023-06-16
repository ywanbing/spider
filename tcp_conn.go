package spider

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"time"

	"github.com/ywanbing/spider/common"
	"github.com/ywanbing/spider/message"
	"github.com/ywanbing/spider/proto"
)

type TcpConn interface {
	proto.Proto
	net.Conn

	// GetConnId returns the connection id.
	GetConnId() uint64
	SetConnId(uint64)

	Start()

	SendMsg(message.Message) error

	// StopNotifyChan 关闭的时候，需要被通知
	StopNotifyChan() chan struct{}
}

type tcpConn struct {
	*net.TCPConn
	proto.Proto

	// 消息处理函数
	handleFunc func(ctx *Context)

	// 创建连接的配置。
	cfg ConnConfig

	// connId 连接id
	connId uint64

	// byte数组 缓存池
	bufferPool *common.LimitedPool

	// 收发消息的通道
	recvChan chan []byte
	sendChan chan []byte

	stop           bool
	stopNotifyChan chan struct{}
}

var _ TcpConn = new(tcpConn)

func NewTcpConn(conn *net.TCPConn, cfg ConnConfig, handleFunc func(ctx *Context)) TcpConn {
	return &tcpConn{
		TCPConn:        conn,
		Proto:          cfg.p,
		cfg:            cfg,
		handleFunc:     handleFunc,
		bufferPool:     common.NewLimitedPool(cfg.binaryPoolMinSize, cfg.binaryPoolMaxSize),
		recvChan:       make(chan []byte, cfg.maxRecvMsgNum),
		sendChan:       make(chan []byte, cfg.maxSendMsgNum),
		stopNotifyChan: make(chan struct{}),
	}
}

func (t *tcpConn) GetConnId() uint64 {
	if t == nil {
		return 0
	}
	return t.connId
}

func (t *tcpConn) SetConnId(connId uint64) {
	if t == nil {
		return
	}
	t.connId = connId
}

func (t *tcpConn) Start() {
	go t.handFunc()
	go t.send()
}

func (t *tcpConn) Stop() {
	if t.IsStop() {
		return
	}
	t.stop = true
	_ = t.Close()
	close(t.recvChan)
	close(t.sendChan)
	close(t.stopNotifyChan)
	return
}

func (t *tcpConn) StopNotifyChan() chan struct{} {
	return t.stopNotifyChan
}

func (t *tcpConn) IsStop() bool {
	return t.stop
}

func (t *tcpConn) recv() {
	defer t.Stop()
	sizeByte := make([]byte, 4)

	reader := bufio.NewReaderSize(t, int(t.cfg.recvBufferSize))
	for {
		if t.IsStop() {
			return
		}

		_, err := io.ReadFull(reader, sizeByte)
		if err != nil {
			if errors.Is(err, io.EOF) {
				continue
			} else {
				// 关闭连接或者其他错误
				return
			}
		}

		// 读取消息长度
		allSize := binary.BigEndian.Uint32(sizeByte)
		data := t.bufferPool.Get(int(allSize - proto.MsgSize))

		_, err = io.ReadFull(reader, data)
		if err != nil {
			if errors.Is(err, io.EOF) {
				continue
			} else {
				// 关闭连接或者其他错误
				return
			}
		}

		select {
		case t.recvChan <- data:
		default:
			// 一直没有读取，直接关闭连接
			return
		}
	}
}

func (t *tcpConn) SendMsg(data message.Message) error {
	msg, err := t.Pack(data)
	if err != nil {
		return err
	}

	select {
	case t.sendChan <- msg:
	default:
		return errors.New("send chan is full")
	}
	return nil
}

func (t *tcpConn) send() {
	defer t.Stop()
	for msg := range t.sendChan {
		if t.IsStop() {
			return
		}

		_ = t.SetWriteDeadline(time.Now().Add(t.cfg.writeTimeout))
		_, err := t.Write(msg)
		if err != nil {
			// TODO LOG
			return
		}
	}
}

func (t *tcpConn) handFunc() {
	defer t.Stop()
	go t.recv()
	for msg := range t.recvChan {
		if t.IsStop() {
			return
		}

		m, _ := t.Unpack(msg)
		// 回收
		t.bufferPool.Put(msg)

		// 检查消息
		if err := m.Check(); err != nil {
			// 只有请求的消息才会返回错误
			if m.GetHeader()[message.MsgTypeKey] != message.MsgTypeRequest {
				continue
			}
			m.SetHeader(message.MsgErr, err.Error())
			m.SetHeader(message.MsgTypeKey, message.MsgTypeReply)
			m.SetBody(nil)
			t.SendMsg(m)
			continue
		}

		ctx := NewContext(context.Background(), m, t)

		// 消息处理函数
		go t.handleFunc(ctx)
	}
}
