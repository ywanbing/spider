package spider

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

type TcpConn interface {
	Proto
	net.Conn

	// GetConnId returns the connection id.
	GetConnId() uint64
	SetConnId(uint64)

	Start(t *TcpX)
}

type tcpConn struct {
	*net.TCPConn

	// 保存一个顶级对象的引用
	t *TcpX

	// 创建连接的配置。
	cfg ConnConfig

	// connId 连接id
	connId uint64

	// byte数组 缓存池
	bufferPool *LimitedPool

	// 收发消息的通道
	recvChan chan *[]byte
	sendChan chan *[]byte

	stop bool
}

var _ TcpConn = new(tcpConn)

func NewTcpConn(conn *net.TCPConn, cfg ConnConfig) TcpConn {
	return &tcpConn{
		TCPConn:    conn,
		cfg:        cfg,
		bufferPool: NewLimitedPool(cfg.binaryPoolMinSize, cfg.binaryPoolMaxSize),
		recvChan:   make(chan *[]byte, cfg.maxRecvMsgNum),
		sendChan:   make(chan *[]byte, cfg.maxSendMsgNum),
	}
}

func (t *tcpConn) Pack(m *message) error {
	if t.IsStop() {
		return fmt.Errorf("connection is closed")
	}

	// 元数据默认为 json 序列化
	meatData, _ := json.Marshal(m.metadata)
	meatDataLen := len(meatData)
	if meatDataLen > 0x0fff {
		return fmt.Errorf("metadata is too long")
	}

	bodyData := m.body
	bodyDataLen := len(m.body)

	allSize := AllSize + meatDataLen + bodyDataLen
	data := t.bufferPool.Get(allSize)

	// 1. 写入消息长度
	binary.BigEndian.PutUint32((*data)[:4], uint32(allSize))
	// 2. 写入消息id
	binary.BigEndian.PutUint32((*data)[4:8], m.msgId)
	// 3. 写入序列化类型和头部长度[protoType = 1b, meatDataLen = 3b]
	binary.BigEndian.PutUint32((*data)[8:12], uint32(m.protoType)<<24|uint32(meatDataLen))
	// 4. 写入元数据
	copy((*data)[12:12+meatDataLen], meatData)
	// 5. 写入消息体长度
	binary.BigEndian.PutUint32((*data)[12+meatDataLen:16+meatDataLen], uint32(bodyDataLen))
	// 6. 写入消息体
	copy((*data)[16+meatDataLen:], bodyData)

	select {
	case t.sendChan <- data:
	default:
		t.Stop()
		return fmt.Errorf("sendChan is full")
	}
	return nil
}

func (t *tcpConn) Unpack(msg *[]byte) (*message, error) {
	data := *msg
	msgId := binary.BigEndian.Uint32(data[:4])
	protoTypeAndMeatSize := binary.BigEndian.Uint32(data[4:8])
	protoType := MarshalType(protoTypeAndMeatSize >> 24)
	meatDataLen := protoTypeAndMeatSize & 0x0fff
	meatData := data[8 : 8+meatDataLen]
	bodyDataLen := binary.BigEndian.Uint32(data[8+meatDataLen : 12+meatDataLen])

	// 结束引用
	bodyData := make([]byte, bodyDataLen)
	copy(bodyData, data[12+meatDataLen:])

	// 1. 解析元数据
	meat := make(map[string]any)
	if meatDataLen > 0 {
		_ = json.Unmarshal(meatData, &meat)
	}

	m := newMessage(msgId, protoType, meat, bodyData)
	return m, nil
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

func (t *tcpConn) Start(x *TcpX) {
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
		data := t.bufferPool.Get(int(allSize - 4))

		_, err = io.ReadFull(t, *data)
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

func (t *tcpConn) send() {
	defer t.Stop()
	for msg := range t.sendChan {
		if t.t.IsClosed() || t.IsStop() {
			return
		}

		_ = t.SetWriteDeadline(time.Now().Add(t.cfg.readTimeout))
		_, err := t.Write(*msg)
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
