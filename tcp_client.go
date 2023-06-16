package spider

import (
	"fmt"
	"net"
	"time"

	"github.com/ywanbing/spider/message"
)

type TcpClient struct {
	TcpConn

	cfg ConnConfig
	// 处理消息的路由
	// 作为客户端，路由应该用来处理推送消息或者广播的消息
	mux *Mux

	// 应该存在一个机制用来处理请求的响应消息

	close chan struct{}
}

func NewTcpClient(addr string, cfgOptions ...ConnConfigOption) *TcpClient {
	cfg := defaultConnConfig
	cfg.addr = addr
	for _, opt := range cfgOptions {
		cfg = opt(cfg)
	}

	return &TcpClient{
		cfg:   cfg,
		mux:   newMux(),
		close: make(chan struct{}),
	}
}

// Start connects to the address on the named network.
func (t *TcpClient) Start() error {
	conn, err := net.Dial("tcp", t.cfg.addr)
	if err != nil {
		return err
	}

	tcpConn := NewTcpConn(conn.(*net.TCPConn), t.cfg, t.handleMessage)
	if !t.cfg.onConnHandle(tcpConn) {
		tcpConn.Close()
		return fmt.Errorf("onConnHandle error")
	}

	t.TcpConn = tcpConn
	t.TcpConn.Start()

	// 开启一个协程用来处理断线重连
	go t.reconnect()

	return nil
}

// reconnect 断线重连
func (t *TcpClient) reconnect() {
	defer close(t.close)
	// 断线重连策略

	reconnectTimes := 0
	reconnectTime := 1 // 重连间隔
	for {
		select {
		case <-t.close:
			return
		case <-t.TcpConn.StopNotifyChan():
			if reconnectTimes > 10 {
				// TODO log
				return
			}
			time.Sleep(time.Duration(reconnectTime) * time.Second)
			reconnectTimes++
			reconnectTime *= 2
			if reconnectTime > 10 {
				reconnectTime = 10
			}

			conn, err := net.Dial("tcp", t.cfg.addr)
			if err != nil {
				// TODO log
				continue
			}

			t.TcpConn = NewTcpConn(conn.(*net.TCPConn), t.cfg, t.handleMessage)

			t.TcpConn.Start()
			reconnectTimes = 0
			reconnectTime = 1
		}
	}
}

// RegisterGlobalMiddle add global routing middle handlers.
func (t *TcpClient) RegisterGlobalMiddle(middles ...func(ctx *Context)) {
	t.mux.RegisterGlobalMiddle(middles...)
}

// RegisterModelMiddle add routing middle handlers by modelID.
func (t *TcpClient) RegisterModelMiddle(id modelID, middles ...func(ctx *Context)) {
	t.mux.RegisterModelMiddle(id, middles...)
}

// RegisterHandler add routing handlers by modelID and subMsgID.
func (t *TcpClient) RegisterHandler(id modelID, subID subMsgID, handler func(ctx *Context), middles ...func(ctx *Context)) {
	t.mux.RegisterHandler(id, subID, handler, middles...)
}

// handleMessage 服务器处理消息
func (t *TcpClient) handleMessage(ctx *Context) {
	header := ctx.reqMsg.GetHeader()
	switch message.MsgType(header[message.MsgTypeKey].(float64)) {
	case message.MsgTypeReply:
		// 响应消息
		t.HandleReply(ctx)
	case message.MsgTypePush:
		// 推送消息
		t.HandlePush(ctx)
	case message.MsgTypeHeartBeat:
		// 心跳消息
		t.HandleHeartBeat(ctx)
	default:
		//	TODO log
	}
}

// HandleReply 处理响应消息
func (t *TcpClient) HandleReply(ctx *Context) {
	//	print msg
	body := ctx.reqMsg.GetBody()
	fmt.Printf("recv reply msg: %s", string(body))
}

// HandlePush 处理推送消息
func (t *TcpClient) HandlePush(ctx *Context) {

}

// HandleHeartBeat 处理心跳消息
func (t *TcpClient) HandleHeartBeat(ctx *Context) {

}
