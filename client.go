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
	address         string
	pl              *pipeline
	cctx            context.Context
	cancelFunc      context.CancelFunc
	optHandler      tcpConnOptHandler
	readTimeoutDur  time.Duration
	writeTimeoutDur time.Duration
}

// NewTCPClient create a new TCPClient.
func NewTCPClient() *TCPClient {
	client := new(TCPClient)
	client.pl = new(pipeline)
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
	if c.cctx != nil {
		return errors.New("net: client object is already started")
	}

	var err error
	var conn net.Conn

	conn, err = net.Dial("tcp", c.address)
	if err != nil {
		return fmt.Errorf("net: Dial() failed - %v", err)
	}

	c.cctx, c.cancelFunc = context.WithCancel(context.Background())
	nc := newContext(c, conn)
	go nc.process(c.cctx)

	return nil
}

// Stop stops the service. TCPClient closes the connection.
func (c *TCPClient) Stop() error {
	if c.cctx != nil {
		c.cancel()
	}

	return nil
}

// WaitForDone blocks until service stops.
func (c *TCPClient) WaitForDone() {
	<-c.cctx.Done()
}

func (c *TCPClient) context() context.Context {
	return c.cctx
}

func (c *TCPClient) cancel() {
	c.cctx = nil
	c.cancelFunc()
}

func (c *TCPClient) pipeline() *pipeline {
	return c.pl
}

func (c *TCPClient) readTimeout() time.Duration {
	return c.readTimeoutDur
}

func (c *TCPClient) writeTimeout() time.Duration {
	return c.writeTimeoutDur
}
