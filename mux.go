package spider

import "errors"

type (
	modelID  = int32
	subMsgID = int32
)

// Mux is a multiplexer for network connections.
// 是一个网络路由复用器
type Mux struct {
	// 用于存储消息ID和消息处理函数的映射关系，key为模块ID，value为消息ID和消息处理函数的映射关系
	Handlers map[modelID]*MsgMiddleHandler

	// AllowAdd of allow routes to be added before starting.
	// 只允许在启动前添加路由
	AllowAdd bool

	// global-middles
	// 全局中间件
	GlobalMiddles []func(ctx *Context)
}

// MsgMiddleHandler 模块处理函数
type MsgMiddleHandler struct {
	// 模块中间件
	ModelMiddles []func(ctx *Context)

	// 消息处理函数
	Handlers map[subMsgID]func(ctx *Context)
	// 消息处理函数的中间件
	HandlerMiddles map[subMsgID][]func(ctx *Context)
}

// newMux returns a new Mux.
func newMux() *Mux {
	return &Mux{
		Handlers:      make(map[modelID]*MsgMiddleHandler),
		AllowAdd:      true,
		GlobalMiddles: make([]func(ctx *Context), 0, 4),
	}
}

// RegisterGlobalMiddle add global routing middle handlers.
func (m *Mux) RegisterGlobalMiddle(middles ...func(ctx *Context)) {
	if !m.AllowAdd {
		panic(errors.New("不允许添加中间件,需要在启动前添加"))
	}
	m.GlobalMiddles = append(m.GlobalMiddles, middles...)
}

// RegisterModelMiddle add routing middle handlers by modelID.
func (m *Mux) RegisterModelMiddle(id modelID, middles ...func(ctx *Context)) {
	if !m.AllowAdd {
		panic(errors.New("不允许添加中间件,需要在启动前添加"))
	}

	if m.Handlers[id] == nil {
		m.Handlers[id] = &MsgMiddleHandler{
			ModelMiddles: make([]func(ctx *Context), 0, 4),
		}
	}
	m.Handlers[id].ModelMiddles = append(m.Handlers[id].ModelMiddles, middles...)
}

// RegisterHandler add routing handlers by modelID and subMsgID.
func (m *Mux) RegisterHandler(id modelID, subID subMsgID, handler func(ctx *Context), middles ...func(ctx *Context)) {
	if !m.AllowAdd {
		panic(errors.New("不允许添加路由,需要在启动前添加"))
	}

	if m.Handlers[id] == nil {
		m.Handlers[id] = &MsgMiddleHandler{
			Handlers:       make(map[subMsgID]func(ctx *Context)),
			HandlerMiddles: make(map[subMsgID][]func(ctx *Context)),
		}
	}

	if m.Handlers[id].Handlers[subID] != nil {
		panic(errors.New("路由已存在"))
	}

	m.Handlers[id].Handlers[subID] = handler
	m.Handlers[id].HandlerMiddles[subID] = append(m.Handlers[id].HandlerMiddles[subID], middles...)
}
