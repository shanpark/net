package net

import (
	"context"
	"errors"
	"fmt"
	"net"
)

// A TCPClient represents a client object using tcp network.
type TCPClient struct {
	address string
	pl      *pipeline
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewTCPClient create a new TCPClient.
func NewTCPClient() *TCPClient {
	client := new(TCPClient)
	client.pl = new(pipeline)
	return client
}

// SetAddress sets the remote address to connect. The address has the form "host:port".
func (c *TCPClient) SetAddress(address string) error {
	c.address = address
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
	if c.ctx != nil {
		return errors.New("net: client object is already started")
	}

	var err error
	var conn net.Conn

	conn, err = net.Dial("tcp", c.address)
	if err != nil {
		return fmt.Errorf("net: Dial() failed - %v", err)
	}

	c.ctx, c.cancel = context.WithCancel(context.Background())
	nc := newContext(c, conn)
	go nc.process(c.ctx)

	return nil
}

// Stop stops the service. TCPClient closes the connection.
func (c *TCPClient) Stop() error {
	if c.ctx != nil {
		c.cancel()
		c.ctx = nil
	}

	return nil
}

// WaitForDone blocks until service stops.
func (c *TCPClient) WaitForDone() {
	<-c.ctx.Done()
}

func (c *TCPClient) context() context.Context {
	return c.ctx
}

func (c *TCPClient) cancelFunc() context.CancelFunc {
	return c.cancel
}

func (c *TCPClient) pipeline() *pipeline {
	return c.pl
}
