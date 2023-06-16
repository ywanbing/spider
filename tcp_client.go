package spider

type TcpClient struct {
	TcpConn

	cfg ConnConfig
	mux *Mux

	close chan struct{}
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
