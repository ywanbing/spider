package spider

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/ywanbing/spider/code"
	"github.com/ywanbing/spider/message"
)

type TcpClient struct {
	TcpConn

	cfg ConnConfig
	// 处理消息的路由
	// 作为客户端，路由应该用来处理推送消息或者广播的消息
	mux *Mux

	// 应该存在一个机制用来处理请求的响应消息
	mutex   sync.Mutex // protects following
	seq     uint64
	pending map[uint64]*Call
	close   chan struct{}
}

// Call represents an active req.
type Call struct {
	req   message.Message
	Reply message.Message
	Error error      // After completion, the error status.
	Done  chan *Call // Strobes when call is complete.
}

func (call *Call) done() {
	select {
	case call.Done <- call:
		// ok
	default:
		// TODO log
		//log.Debug("rpc: discarding Call reply due to insufficient Done chan capacity")
	}
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

func (t *TcpClient) IsClose() bool {
	select {
	case <-t.close:
		return true
	default:
		return false
	}
}

func (t *TcpClient) Close() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	select {
	case <-t.close:
	default:
		close(t.close)
	}
}

// reconnect 断线重连
func (t *TcpClient) reconnect() {
	defer t.Close()
	// 断线重连策略

	reconnectTimes := 0
	reconnectTime := 10 // 重连间隔
	for {
		select {
		case <-t.close:
			return
		case <-t.TcpConn.StopNotifyChan():
			if reconnectTimes > 10 {
				// TODO log
				return
			}
			time.Sleep(time.Duration(reconnectTime) * time.Millisecond)
			reconnectTimes++
			reconnectTime *= 2
			if reconnectTime > 5000 {
				reconnectTime = 5000
			}

			conn, err := net.Dial("tcp", t.cfg.addr)
			if err != nil {
				// TODO log
				continue
			}

			t.TcpConn = NewTcpConn(conn.(*net.TCPConn), t.cfg, t.handleMessage)

			t.TcpConn.Start()
			reconnectTimes = 0
			reconnectTime = 10
		}
	}
}

// RegisterGlobalMiddle add global routing middle handlers.
func (t *TcpClient) RegisterGlobalMiddle(middles ...func(ctx *Context)) {
	t.mux.RegisterGlobalMiddle(middles...)
}

// TcpClient 暂时只提供一个全局的路由中间件

// RegisterModelMiddle add routing middle handlers by modelID.
//func (t *TcpClient) RegisterModelMiddle(id modelID, middles ...func(ctx *Context)) {
//	t.mux.RegisterModelMiddle(id, middles...)
//}

// RegisterHandler add routing handlers by modelID and subMsgID.
//func (t *TcpClient) RegisterHandler(id modelID, subID subMsgID, handler func(ctx *Context), middles ...func(ctx *Context)) {
//	t.mux.RegisterHandler(id, subID, handler, middles...)
//}

// handleMessage 服务器处理消息
func (t *TcpClient) handleMessage(ctx *Context) {
	header := ctx.reqMsg.GetHeader()
	switch message.MsgTypeFromString(header[message.MsgTypeKey]) {
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
	// 回复的消息
	respMsg := ctx.reqMsg
	seq := respMsg.GetHeader()[message.MsgSeq]
	errStr := respMsg.GetHeader()[message.MsgErr]
	// 转成数字
	seqNum, err := strconv.ParseUint(seq, 10, 64)
	if err != nil {
		// TODO log
		return
	}

	t.mutex.Lock()
	call := t.pending[seqNum]
	delete(t.pending, seqNum)
	t.mutex.Unlock()
	if call == nil {
		// TODO log
		return
	}

	call.Reply = respMsg
	if errStr != "" {
		call.Error = errors.New(errStr)
	}
	call.done()
}

// HandlePush 处理推送消息
func (t *TcpClient) HandlePush(ctx *Context) {

}

// HandleHeartBeat 处理心跳消息
func (t *TcpClient) HandleHeartBeat(ctx *Context) {

}

/*
func (t *TcpClient) Go(c context.Context, req message.Message) *Call {
	call := new(Call)
	call.req = req
	call.Done = make(chan *Call, 10) // buffered. 依据 rpcx
	go t.send(c, call)
	return call
}

// send 发送消息，由客户端进行中间件的处理
func (t *TcpClient) send(c context.Context, call *Call) {
	defer func() {
		if r := recover(); r != nil {
			// TODO log
			//log.Errorf("client send error is %v", r)
		}
		// write channel ,
		if call != nil {
			call.done()
		}
	}()

	t.mutex.Lock()
	// 检查客户端状态
	if t.IsClose() {
		t.mutex.Unlock()
		call.Error = code.ErrConnClosed
		call.done()
		return
	}

	if t.pending == nil {
		t.pending = make(map[uint64]*Call)
	}

	seq := t.seq
	t.seq++
	t.pending[seq] = call
	t.mutex.Unlock()

	// 设置消息头
	call.req.SetHeader(message.MsgSeq, strconv.FormatUint(seq, 10))
	call.req.SetHeader(message.MsgTypeKey, message.MsgTypeRequest.String())

	// 创建自己的上下文
	ctx := NewContext(c, call.req, t.TcpConn)
	if ctx.handlers == nil {
		ctx.handlers = make([]func(c *Context), 0, len(t.mux.GlobalMiddles)+1)
	}

	// global middleware
	ctx.handlers = append(ctx.handlers, t.mux.GlobalMiddles...)

	// handler
	ctx.handlers = append(ctx.handlers, func(c *Context) {
		err := t.TcpConn.SendMsg(c.reqMsg)
		if err != nil {
			call.Error = err
			call.done()
		}
	})

	// 执行
	if len(ctx.handlers) > 0 {
		ctx.Next()
	}
}*/

// Call 发送消息，由客户端进行中间件的处理
func (t *TcpClient) Call(c context.Context, req message.Message) (resp message.Message, err error) {
	// 检查客户端状态
	if t.IsClose() {
		t.mutex.Unlock()
		return nil, code.ErrConnClosed
	}

	call := new(Call)
	call.req = req
	call.Done = make(chan *Call, 10) // buffered. 依据 rpcx
	t.mutex.Lock()
	if t.pending == nil {
		t.pending = make(map[uint64]*Call)
	}

	seq := t.seq
	t.seq++
	t.pending[seq] = call
	t.mutex.Unlock()

	// 设置消息头
	call.req.SetHeader(message.MsgSeq, strconv.FormatUint(seq, 10))
	call.req.SetHeader(message.MsgTypeKey, message.MsgTypeRequest.String())

	// 创建自己的上下文
	ctx := NewContext(c, req, t.TcpConn)
	if ctx.handlers == nil {
		ctx.handlers = make([]func(c *Context), 0, len(t.mux.GlobalMiddles)+1)
	}

	// global middleware
	ctx.handlers = append(ctx.handlers, t.mux.GlobalMiddles...)

	var send bool

	// handler
	ctx.handlers = append(ctx.handlers, func(c *Context) {
		err := t.TcpConn.SendMsg(c.reqMsg)
		if err != nil {
			t.mutex.Lock()
			delete(t.pending, seq)
			t.mutex.Unlock()
			call.Error = err
			call.done()
		}

		send = true

		select {
		case call = <-call.Done:
		case <-c.ctx.Done():
			t.mutex.Lock()
			delete(t.pending, seq)
			t.mutex.Unlock()
			call.Error = c.ctx.Err()
			call.done()
		}
	})

	// 执行
	defer func() {
		if !send {
			t.mutex.Lock()
			delete(t.pending, seq)
			t.mutex.Unlock()
			call.Error = code.ErrMessageNotSent
			call.done()
		}
	}()
	ctx.Next()

	return call.Reply, call.Error
}
