package net

import (
	"net"
	"time"
)

type tcpConnOptHandler struct {
	noDelay         *bool
	keepAlive       *bool
	keepAlivePeriod time.Duration
	linger          *int
	readBuffersize  *int
	writeBufferSize *int
}

func (h tcpConnOptHandler) OnConnect(ctx *SoContext) error {
	// set no delay option
	if h.noDelay != nil {
		ctx.conn.(*net.TCPConn).SetNoDelay(*h.noDelay)
	}

	// set keep alive option
	if h.keepAlive != nil {
		ctx.conn.(*net.TCPConn).SetKeepAlive(*h.keepAlive)
		if h.keepAlivePeriod != 0 {
			ctx.conn.(*net.TCPConn).SetKeepAlivePeriod(h.keepAlivePeriod)
		}
	}

	// set linger option
	if h.linger != nil {
		ctx.conn.(*net.TCPConn).SetLinger(*h.linger)
	}

	// set read buffer size option
	if h.readBuffersize != nil {
		ctx.conn.(*net.TCPConn).SetReadBuffer(*h.readBuffersize)
	}

	// set write buffer size option
	if h.writeBufferSize != nil {
		ctx.conn.(*net.TCPConn).SetWriteBuffer(*h.writeBufferSize)
	}

	return nil
}
