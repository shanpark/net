package net

import (
	"context"
	"io"
	"net"
	"time"
)

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
	nctx.buffer = NewBuffer()
	return nctx
}

func (nctx *Context) startRead(ctx context.Context, readCmdCh <-chan int) (<-chan int, <-chan error) {
	dataReadyCh := make(chan int)
	errCh := make(chan error)
	go nctx.read(ctx, readCmdCh, dataReadyCh, errCh)
	return dataReadyCh, errCh
}

func (nctx *Context) read(ctx context.Context, readCmdCh <-chan int, dataReadyCh chan<- int, errCh chan<- error) {
	defer close(dataReadyCh)
	defer close(errCh)

	var err error
	var n int

	for {
		if 0 == <-readCmdCh { // wait read command
			return // readCmdCh closed. just return.
		}

		if nctx.service.readTimeout() > 0 {
			nctx.conn.SetReadDeadline(time.Now().Add(nctx.service.readTimeout())) // set timeout
		}
		nctx.prepareRead()
		if n, err = nctx.conn.Read(nctx.buffer.Buffer()); err != nil {
			select {
			case <-ctx.Done():
			default:
				errCh <- err
			}
			return
		}
		nctx.buffer.BufferConsume(n)

		dataReadyCh <- n
	}
}

func (nctx *Context) prepareRead() {
	nctx.rollback = false
	nctx.buffer.Reserve(4096)
}

func (nctx *Context) process(ctx context.Context) {
	defer nctx.conn.Close()

	var err error

	for _, handler := range nctx.service.pipeline().connectHandlers {
		if err = handler.OnConnect(nctx); err != nil {
			nctx.handleError(err)
			return
		}
	}
	defer nctx.handleDisconnect()

	readCmdCh := make(chan int) // read command channel
	defer close(readCmdCh)

	dataReadyCh, errCh := nctx.startRead(ctx, readCmdCh)
Loop:
	for {
		readCmdCh <- 1 // send read command

		select {
		case <-ctx.Done():
			return

		case <-dataReadyCh:
			for {
				var out interface{} = nctx.buffer
				var remain = nctx.buffer.Readable()
				for _, handler := range nctx.service.pipeline().readHandlers {
					if nctx.IsRollback() {
						nctx.buffer.Rollback()
					}
					if out, err = handler.OnRead(nctx, out); err != nil {
						nctx.handleError(err)
						return
					}
					if out == nil {
						break
					}
				}
				nctx.Commit()
				if (nctx.buffer.Readable() == 0) || (nctx.buffer.Readable() == remain) {
					break
				}
			}

		case err = <-errCh:
			if err == io.EOF {
				nctx.Close()
			} else if err.(net.Error).Timeout() {
				nctx.handleTimeout()
			} else {
				nctx.handleError(err)
				if nctx.service.context() != nil {
					dataReadyCh, errCh = nctx.startRead(ctx, readCmdCh)
				}
			}
		}
	}
}

func (nctx *Context) handleDisconnect() {
	for _, handler := range nctx.service.pipeline().disconnectHandlers {
		handler.OnDisconnect(nctx)
	}
}

func (nctx *Context) handleTimeout() {
	var err error
	for _, handler := range nctx.service.pipeline().timeoutHandlers {
		if err = handler.OnTimeout(nctx); err != nil {
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
