package net

import (
	"context"
	"net"
)

type Context struct {
	service  Service
	conn     net.Conn
	buffer   *Buffer
	rollback bool
}

func NewContext(service Service, conn net.Conn) *Context {
	nc := new(Context)
	nc.service = service
	nc.conn = conn
	nc.buffer = NewBuffer()
	return nc
}

func (nc *Context) Rollback() {
	nc.rollback = true
}

func (nc *Context) IsRollback() bool {
	return nc.rollback
}

func (nc *Context) Commit() {
	nc.rollback = false
	nc.buffer.commit()
}

func (nc *Context) Write(out interface{}) {
	var err error
	var n int

	for {
		for inx := len(nc.service.Pipeline().Handlers()) - 1; inx >= 0; inx-- {
			handler := nc.service.Pipeline().Handlers()[inx]
			if outb, ok := handler.(Outbounder); ok {
				if out, err = outb.Outbound(nc, out); err != nil {
					// Close Handler 또는 Error Handler를 호출하고 process()를 종료하면 어떨까?
				}
				if out == nil {
					break
				}
			}
		}

		buffer := out.(*Buffer).Data()
		written := 0
		for written < len(buffer) {
			n, err = nc.conn.Write(buffer[written:])
			if err != nil {
				// TODO error 처리.
				break
			}
			written += n
		}
	}
}

func (nc *Context) prepareRead() {
	nc.rollback = false
	nc.buffer.Reserve(4096)
}

func (nc *Context) process(ctx context.Context) {
	var err error
	var n int

	defer nc.conn.Close()

Loop:
	for {
		nc.prepareRead()

		if n, err = nc.conn.Read(nc.buffer.Buffer()); err != nil {
			// TODO error 처리.
			return
		}
		nc.buffer.Written(n)

		var out interface{} = nc.buffer
		for _, handler := range nc.service.Pipeline().Handlers() {
			if inb, ok := handler.(Inbounder); ok {
				if out, err = inb.Inbound(nc, out); err != nil {
					// Close Handler 또는 Error Handler를 호출하고 process()를 종료하면 어떨까?
				}
				if nc.IsRollback() {
					nc.buffer.rollback()
					continue Loop
				}
				if out == nil {
					break
				}
			}
		}

		nc.Commit()

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}
