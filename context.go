package net

import (
	"context"
	"net"
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
func (nc *Context) Rollback() {
	nc.rollback = true
}

// IsRollback returns whether a rollback request has been made.
func (nc *Context) IsRollback() bool {
	return nc.rollback
}

// Commit commits the current state of the read operation.
func (nc *Context) Commit() {
	nc.rollback = false
	nc.buffer.Commit()
}

// Write writes 'out' to the peer. This causes the WriteHandler chain to be called.
func (nc *Context) Write(out interface{}) error {
	var err error
	var n int

	for _, handler := range nc.service.pipeline().writeHandlers {
		if out, err = handler.OnWrite(nc, out); err != nil {
			return err
		}
		if out == nil {
			break
		}
	}

	buffer := out.(*Buffer).Data()
	written := 0
	for written < len(buffer) {
		n, err = nc.conn.Write(buffer[written:])
		if err != nil {
			return err
		}
		written += n
	}

	return nil
}

// Close requests context to close the connection of the context.
func (nc *Context) Close() {
	nc.service.cancelFunc()()
}

func newContext(service Service, conn net.Conn) *Context {
	nc := new(Context)
	nc.service = service
	nc.conn = conn
	nc.buffer = NewBuffer()
	return nc
}

func (nc *Context) startRead(ctx context.Context, readCmdCh <-chan int) (<-chan int, <-chan error) {
	dataReadyCh := make(chan int)
	errCh := make(chan error)
	go nc.read(ctx, readCmdCh, dataReadyCh, errCh)
	return dataReadyCh, errCh
}

func (nc *Context) read(ctx context.Context, readCmdCh <-chan int, dataReadyCh chan<- int, errCh chan<- error) {
	defer close(dataReadyCh)
	defer close(errCh)

	var err error
	var n int

	for {
		if 0 == <-readCmdCh { // wait read command
			return // readCmdCh closed. just return.
		}

		nc.prepareRead()
		if n, err = nc.conn.Read(nc.buffer.Buffer()); err != nil {
			select {
			case <-ctx.Done():
			default:
				errCh <- err
			}
			return
		}
		nc.buffer.BufferConsume(n)

		dataReadyCh <- n
	}
}

func (nc *Context) prepareRead() {
	nc.rollback = false
	nc.buffer.Reserve(4096)
}

func (nc *Context) process(ctx context.Context) {
	defer nc.conn.Close()

	var err error

	for _, handler := range nc.service.pipeline().connectHandlers {
		if err = handler.OnConnect(nc); err != nil {
			nc.handleError(err)
			return
		}
	}
	defer nc.handleDisconnect()

	readCmdCh := make(chan int) // read command channel
	defer close(readCmdCh)

	dataReadyCh, errCh := nc.startRead(ctx, readCmdCh)
Loop:
	for {
		readCmdCh <- 1 // send read command

		select {
		case <-ctx.Done():
			return

		case <-dataReadyCh:
			var out interface{} = nc.buffer
			for _, handler := range nc.service.pipeline().readHandlers {
				if out, err = handler.OnRead(nc, out); err != nil {
					nc.handleError(err)
					return
				}
				if nc.IsRollback() {
					nc.buffer.Rollback()
					continue Loop
				}
				if out == nil {
					break
				}
			}
			nc.Commit()

		case err = <-errCh:
			nc.handleError(err)
			if nc.service.context() != nil {
				dataReadyCh, errCh = nc.startRead(ctx, readCmdCh)
			}
		}
	}
}

func (nc *Context) handleDisconnect() {
	for _, handler := range nc.service.pipeline().disconnectHandlers {
		handler.OnDisconnect(nc)
	}
}

func (nc *Context) handleError(err error) {
	for _, handler := range nc.service.pipeline().errorHandlers {
		handler.OnError(nc, err)
	}
}
