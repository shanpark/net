package net

import "net"

type Context struct {
	conn   net.Conn
	buffer *Buffer
}

func NewContext(conn net.Conn) *Context {
	ctx := new(Context)
	ctx.conn = conn
	ctx.buffer = NewBuffer()
	return ctx
}

func (ctx *Context) Process() {
	n, err := ctx.conn.Read(ctx.buffer.Buffer())
	ctx.buffer.Written(n)
}
