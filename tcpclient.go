package net

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"
)

// A TCPClient represents a client object using tcp network.
type TCPClient struct {
	address    string
	cancelFunc context.CancelFunc
	doneCh     <-chan struct{}

	nctx            *SoContext
	pl              *soPipeline
	optHandler      tcpConnOptHandler
	readTimeoutDur  time.Duration
	writeTimeoutDur time.Duration
}

// NewTCPClient create a new TCPClient.
func NewTCPClient() *TCPClient {
	client := new(TCPClient)
	client.pl = new(soPipeline)
	client.AddHandler(client.optHandler)
	return client
}

// SetAddress sets the remote address to connect. The address has the form "host:port".
func (c *TCPClient) SetAddress(address string) error {
	c.address = address
	return nil
}

// SetTimeout sets the read and write timeout associated with the connection.
func (c *TCPClient) SetTimeout(readTimeout time.Duration, writeTimeout time.Duration) error {
	c.readTimeoutDur = readTimeout
	c.writeTimeoutDur = writeTimeout
	return nil
}

// SetNoDelay controls whether the operating system should delay packet transmission in hopes of sending fewer packets (Nagle's algorithm).
// The default is true (no delay), meaning that data is sent as soon as possible after a Write.
func (c *TCPClient) SetNoDelay(noDelay bool) error {
	c.optHandler.noDelay = new(bool)
	*c.optHandler.noDelay = noDelay
	return nil
}

// SetKeepAlive sets whether the operating system should send keep-alive messages on the connection.
// If period is zero, default value will be used.
func (c *TCPClient) SetKeepAlive(keepAlive bool, period time.Duration) error {
	c.optHandler.keepAlive = new(bool)
	*c.optHandler.keepAlive = keepAlive
	c.optHandler.keepAlivePeriod = period
	return nil
}

// AddHandler adds a handler for network events. A handler should implement at least one of interfaces ReadHandler, WriteHandler, ConnectHandler, DisconnectHandler, ErrorHandler.
// Handlers are called in order and only in the case of WriteHandler are called in the reverse direction.
func (c *TCPClient) AddHandler(handlers ...interface{}) error {
	for _, handler := range handlers {
		if err := c.pl.AddHandler(handler); err != nil {
			return err
		}
	}

	return nil
}

// Start starts the service. TCPClient connects to the remote address.
func (c *TCPClient) Start() error {
	if c.isRunning() {
		return errors.New("net: client object is already started")
	}

	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		return fmt.Errorf("net: Dial() failed - %v", err)
	}

	var cctx context.Context
	cctx, c.cancelFunc = context.WithCancel(context.Background())
	c.doneCh = cctx.Done()
	nctx := newContext(c, conn, defaultQueueSize)
	go nctx.process()
	c.nctx = nctx

	return nil
}

// Stop stops the service. TCPClient closes the connection.
func (c *TCPClient) Stop() error {
	if c.isRunning() {
		c.cancel()
	}

	return nil
}

// WaitForDone blocks until service stops.
func (c *TCPClient) WaitForDone() {
	<-c.done()
}

// Write writes parameter out to the peer. This causes the WriteHandler chain to be called.
func (c *TCPClient) Write(out interface{}) error {
	if c.nctx == nil {
		return errors.New("net: connection is not made")
	}

	return c.nctx.Write(out)
}

func (c *TCPClient) pipeline() *soPipeline {
	return c.pl
}

func (c *TCPClient) readTimeout() time.Duration {
	return c.readTimeoutDur
}

func (c *TCPClient) writeTimeout() time.Duration {
	return c.writeTimeoutDur
}

func (c *TCPClient) cancel() {
	c.cancelFunc()
}

func (c *TCPClient) done() <-chan struct{} {
	return c.doneCh
}

func (c *TCPClient) isRunning() bool {
	if c.doneCh == nil {
		return false
	}

	select {
	case <-c.doneCh:
		return false
	default:
		return true
	}
}
