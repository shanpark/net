package net

import (
	"io"
	"net"
	"time"
)

const (
	eventNone = iota
	eventRead
	eventWrite
)

const defaultQueueSize = 32

type event struct {
	id    int
	param interface{}
}

// SoContext represents the current states of the TCP network session.
// And some requests are made through TCPContext.
type SoContext struct {
	svc        soObject
	conn       net.Conn
	eventQueue chan event
	buffer     *Buffer
	rollback   bool
}

// Conn returns an underlying net.Conn
func (nctx *SoContext) Conn() net.Conn {
	return nctx.conn
}

// Rollback requests that the status of the read operation be rolled back to its last commit state.
func (nctx *SoContext) Rollback() {
	nctx.rollback = true
}

// IsRollback returns whether a rollback request has been made.
func (nctx *SoContext) IsRollback() bool {
	return nctx.rollback
}

// Commit commits the current state of the read operation.
func (nctx *SoContext) Commit() {
	nctx.rollback = false
	nctx.buffer.Commit()
}

// Write writes parameter out to the peer. This causes the WriteHandler chain to be called.
func (nctx *SoContext) Write(out interface{}) error {
	nctx.eventQueue <- event{eventWrite, out}
	return nil
}

// Close requests context to close the connection of the context.
func (nctx *SoContext) Close() {
	nctx.svc.cancel()
}

func newContext(svc soObject, conn net.Conn, queueSize int) *SoContext {
	nctx := new(SoContext)
	nctx.svc = svc
	nctx.conn = conn
	nctx.eventQueue = make(chan event, queueSize)
	nctx.buffer = NewBuffer(4096)
	return nctx
}

func (nctx *SoContext) process() {
	defer nctx.conn.Close()
	defer nctx.handleDisconnect()

	if !nctx.handleConnect() {
		return
	}

	go nctx.readLoop()

	for {
		select {
		case <-nctx.svc.done():
			return
		case evt := <-nctx.eventQueue:
			switch evt.id {
			case eventRead:
				nctx.handleRead()
			case eventWrite:
				nctx.handleWrite(evt.param)
			}
		}
	}
}

func (nctx *SoContext) readLoop() {
	readBuf := make([]byte, 4096)
	for {
		select {
		case <-nctx.svc.done():
			return

		default:
			if nctx.svc.readTimeout() > 0 {
				nctx.conn.SetReadDeadline(time.Now().Add(nctx.svc.readTimeout())) // set timeout
			}
			n, err := nctx.conn.Read(readBuf)
			if err != nil {
				select {
				case <-nctx.svc.done():
				default:
					if err == io.EOF {
						nctx.Close()
					} else {
						nerr, ok := err.(net.Error)
						if ok {
							if nerr.Timeout() {
								nctx.handleTimeout()
							} else {
								nctx.handleError(err)
							}
						} else {
							nctx.handleError(err)
						}
					}
				}
				continue
			}
			nctx.buffer.Write(readBuf[:n])

			nctx.eventQueue <- event{eventRead, nil}
		}
	}
}

func (nctx *SoContext) prepareRead() {
	nctx.rollback = false
	nctx.buffer.Reserve(4096)
}

func (nctx *SoContext) handleConnect() bool {
	for _, handler := range nctx.svc.pipeline().connectHandlers {
		if nctx.svc.isRunning() {
			if err := handler.OnConnect(nctx); err != nil {
				nctx.handleError(err)
				return false // stop process
			}
		} else {
			return false // stop process
		}
	}
	return true // continue process
}

func (nctx *SoContext) handleDisconnect() {
	for _, handler := range nctx.svc.pipeline().disconnectHandlers {
		handler.OnDisconnect(nctx)
	}
}

func (nctx *SoContext) handleRead() {
ReadLoop:
	for {
		var err error
		var out interface{} = nctx.buffer
		var remain = nctx.buffer.Readable()
		for _, handler := range nctx.svc.pipeline().readHandlers {
			if nctx.svc.isRunning() {
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
		}
		nctx.Commit()
		if (nctx.buffer.Readable() == 0) || (nctx.buffer.Readable() == remain) {
			break
		}
	}
}

func (nctx *SoContext) handleWrite(out interface{}) {
	var err error
	for _, handler := range nctx.svc.pipeline().writeHandlers {
		if nctx.svc.isRunning() {
			if out, err = handler.OnWrite(nctx, out); err != nil {
				nctx.handleError(err)
				return
			}
		}
	}

	buffer, ok := out.(*Buffer)
	if !ok {
		nctx.handleError(err)
		return
	}
	bytes := buffer.Data()
	written := 0
	for written < len(bytes) {
		if nctx.svc.writeTimeout() > 0 {
			nctx.conn.SetWriteDeadline(time.Now().Add(nctx.svc.writeTimeout())) // set timeout
		}

		n, err := nctx.conn.Write(bytes[written:])
		if err != nil {
			nctx.handleError(err)
			return
		}
		written += n
	}
}

func (nctx *SoContext) handleTimeout() {
	for _, handler := range nctx.svc.pipeline().timeoutHandlers {
		if nctx.svc.isRunning() {
			if err := handler.OnTimeout(nctx); err != nil {
				nctx.handleError(err)
				break
			}
		}
	}
}

func (nctx *SoContext) handleError(err error) {
	for _, handler := range nctx.svc.pipeline().errorHandlers {
		if nctx.svc.isRunning() {
			handler.OnError(nctx, err)
		}
	}
}
