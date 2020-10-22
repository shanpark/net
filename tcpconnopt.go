package net

import (
	"net"
	"time"
)

type tcpConnOptHandler struct {
	noDelay         *bool
	keepAlive       *bool
	keepAlivePeriod time.Duration
}

func (h tcpConnOptHandler) OnConnect(ctx *TCPContext) error {
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

	return nil
}
