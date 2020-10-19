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

const defaultQueueSize = 32

type event struct {
	id    int
	param interface{}
}

// Context represents the current states of the network session.
// And many requests are made through Context.
type Context struct {
	service    Service
	conn       net.Conn
	eventQueue chan event
	buffer     *Buffer
	rollback   bool
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
	nctx.eventQueue <- event{eventWrite, out}
	return nil
}

// Close requests context to close the connection of the context.
func (nctx *Context) Close() {
	nctx.service.cancel()
}

func newContext(service Service, conn net.Conn, queueSize int) *Context {
	nctx := new(Context)
	nctx.service = service
	nctx.conn = conn
	nctx.eventQueue = make(chan event, queueSize)
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

	var err error
EventLoop:
	for {
		select {
		case <-nctx.service.done():
			return
		case evt := <-nctx.eventQueue:
			switch evt.id {
			case eventRead:
				nctx.handleRead()

			case eventWrite:
				out := evt.param
				for _, handler := range nctx.service.pipeline().writeHandlers {
					if out, err = handler.OnWrite(nctx, out); err != nil {
						nctx.handleError(err)
						continue EventLoop
					}
				}

				buffer, ok := out.(*Buffer)
				if !ok {
					nctx.handleError(err)
					continue EventLoop
				}
				bytes := buffer.Data()
				written := 0
				for written < len(bytes) {
					if nctx.service.writeTimeout() > 0 {
						nctx.conn.SetWriteDeadline(time.Now().Add(nctx.service.writeTimeout())) // set timeout
					}

					n, err := nctx.conn.Write(bytes[written:])
					if err != nil {
						nctx.handleError(err)
						continue EventLoop
					}
					written += n
				}
			}
		}
	}
}

func (nctx *Context) readLoop() {
	readBuf := make([]byte, 4096)
	for {
		select {
		case <-nctx.service.done():
			return

		default:
			if nctx.service.readTimeout() > 0 {
				nctx.conn.SetReadDeadline(time.Now().Add(nctx.service.readTimeout())) // set timeout
			}
			n, err := nctx.conn.Read(readBuf)
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
				continue
			}
			nctx.buffer.Write(readBuf[:n])

			nctx.eventQueue <- event{eventRead, nil}
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
