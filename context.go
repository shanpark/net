package net

import (
	"io"
	"net"
	"time"
)

const (
	eventNone = iota
	eventConnect
	eventDisconnect
	eventRead
	eventWrite
	eventError
)

type event struct {
	id  int
	err error
}

// Context represents the current states of the network session.
// And many requests are made through Context.
type Context struct {
	service  Service
	conn     net.Conn
	buffer   *Buffer
	rollback bool
}

// Rollback requests that the status of the read operation be rolled back to its last commit state.
func (nctx *Context) Rollback() {
	nctx.rollback = true
}

// IsRollback returns whether a rollback request has been made.
func (nctx *Context) IsRollback() bool {
	return nctx.rollback
}

// Commit commits the current state of the read operation.
func (nctx *Context) Commit() {
	nctx.rollback = false
	nctx.buffer.Commit()
}

// Write writes parameter out to the peer. This causes the WriteHandler chain to be called.
func (nctx *Context) Write(out interface{}) error {
	var err error
	var n int

	for _, handler := range nctx.service.pipeline().writeHandlers {
		if out, err = handler.OnWrite(nctx, out); err != nil {
			return err
		}
		if out == nil {
			break
		}
	}

	buffer := out.(*Buffer).Data()
	written := 0
	for written < len(buffer) {
		if nctx.service.writeTimeout() > 0 {
			nctx.conn.SetWriteDeadline(time.Now().Add(nctx.service.writeTimeout())) // set timeout
		}

		n, err = nctx.conn.Write(buffer[written:])
		if err != nil {
			return err
		}
		written += n
	}

	return nil
}

// Close requests context to close the connection of the context.
func (nctx *Context) Close() {
	nctx.service.cancel()
}

func newContext(service Service, conn net.Conn) *Context {
	nctx := new(Context)
	nctx.service = service
	nctx.conn = conn
	nctx.buffer = NewBuffer(4096)
	return nctx
}

func (nctx *Context) process() {
	defer nctx.conn.Close()

	if !nctx.handleConnect() {
		return
	}
	defer nctx.handleDisconnect()

	go nctx.readLoop()

	<-nctx.service.done()
}

func (nctx *Context) readLoop() {
	for {
		select {
		case <-nctx.service.done():
			return

		default:
			if nctx.service.readTimeout() > 0 {
				nctx.conn.SetReadDeadline(time.Now().Add(nctx.service.readTimeout())) // set timeout
			}
			nctx.prepareRead()
			n, err := nctx.conn.Read(nctx.buffer.Buffer())
			if err != nil {
				select {
				case <-nctx.service.done():
				default:
					if err == io.EOF {
						nctx.Close()
					} else if err.(net.Error).Timeout() {
						nctx.handleTimeout()
					} else {
						nctx.handleError(err)
					}
				}
				return
			}
			nctx.buffer.BufferConsume(n)

			nctx.handleRead()
		}
	}
}

func (nctx *Context) prepareRead() {
	nctx.rollback = false
	nctx.buffer.Reserve(4096)
}

func (nctx *Context) handleConnect() bool {
	for _, handler := range nctx.service.pipeline().connectHandlers {
		if err := handler.OnConnect(nctx); err != nil {
			nctx.handleError(err)
			return false
		}
	}
	return true
}

func (nctx *Context) handleDisconnect() {
	for _, handler := range nctx.service.pipeline().disconnectHandlers {
		handler.OnDisconnect(nctx)
	}
}

func (nctx *Context) handleRead() {
ReadLoop:
	for {
		var err error
		var out interface{} = nctx.buffer
		var remain = nctx.buffer.Readable()
		for _, handler := range nctx.service.pipeline().readHandlers {
			out, err = handler.OnRead(nctx, out)
			if nctx.IsRollback() || (err != nil) {
				if nctx.IsRollback() {
					nctx.buffer.Rollback()
				}
				if err != nil {
					nctx.handleError(err)
				}
				nctx.Commit() // Commit() initialize rollback flag.
				break ReadLoop
			}
		}
		nctx.Commit()
		if (nctx.buffer.Readable() == 0) || (nctx.buffer.Readable() == remain) {
			break
		}
	}
}

func (nctx *Context) handleTimeout() {
	for _, handler := range nctx.service.pipeline().timeoutHandlers {
		if err := handler.OnTimeout(nctx); err != nil {
			nctx.handleError(err)
			break
		}
	}
}

func (nctx *Context) handleError(err error) {
	for _, handler := range nctx.service.pipeline().errorHandlers {
		handler.OnError(nctx, err)
	}
}
