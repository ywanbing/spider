package spider

import (
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

	Start(t *TcpServer)

	SendMsg(message.Message) error
}

type tcpConn struct {
	*net.TCPConn
	proto.Proto

	// 保存一个顶级对象的引用
	t *TcpServer

	// 创建连接的配置。
	cfg ConnConfig

	// connId 连接id
	connId uint64

	// byte数组 缓存池
	bufferPool *common.LimitedPool

	// 收发消息的通道
	recvChan chan []byte
	sendChan chan []byte

	stop bool
}

var _ TcpConn = new(tcpConn)

func NewTcpConn(conn *net.TCPConn, cfg ConnConfig) TcpConn {
	return &tcpConn{
		TCPConn:    conn,
		Proto:      cfg.p,
		cfg:        cfg,
		bufferPool: common.NewLimitedPool(cfg.binaryPoolMinSize, cfg.binaryPoolMaxSize),
		recvChan:   make(chan []byte, cfg.maxRecvMsgNum),
		sendChan:   make(chan []byte, cfg.maxSendMsgNum),
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

func (t *tcpConn) Start(x *TcpServer) {
	t.t = x

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
	return
}

func (t *tcpConn) IsStop() bool {
	return t.stop
}

func (t *tcpConn) recv() {
	defer t.Stop()
	sizeByte := make([]byte, 4)
	for {
		if t.t.IsClosed() || t.IsStop() {
			return
		}

		_, err := io.ReadFull(t, sizeByte)
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
		data := t.bufferPool.Get(int(allSize))

		_, err = io.ReadFull(t, data)
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
		if t.t.IsClosed() || t.IsStop() {
			return
		}

		_ = t.SetWriteDeadline(time.Now().Add(t.cfg.writeTimeout))
		_, err := t.Write(msg)
		if err != nil {
			// TODO LOG
			return
		}
		t.bufferPool.Put(msg)
	}
}

func (t *tcpConn) handFunc() {
	defer t.Stop()
	go t.recv()
	for msg := range t.recvChan {
		if t.t.IsClosed() || t.IsStop() {
			return
		}

		m, _ := t.Unpack(msg)

		// 回收
		t.bufferPool.Put(msg)

		ctx := NewContext(context.Background(), m, t, t.t)

		// 消息处理函数
		go handleMessage(ctx)
	}
}
