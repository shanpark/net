package net

import (
	"context"
	"errors"
	"net"
)

type TCPClient struct {
	address  string
	pipeline *Pipeline
	conn     net.Conn
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewTCPClient() *TCPClient {
	client := new(TCPClient)
	client.pipeline = new(Pipeline)
	return client
}

func (c *TCPClient) SetAddress(address string) error {
	c.address = address
	return nil
}

func (c *TCPClient) SetPipeline(pipeline *Pipeline) {
	c.pipeline = pipeline
}

func (c *TCPClient) Pipeline() *Pipeline {
	return c.pipeline
}

func (c *TCPClient) Start() error {
	var err error

	if c.ctx != nil {
		return errors.New("net: client object is already started")
	}

	c.conn, err = net.Dial("tcp", c.address)
	if err != nil {
		return err
	}

	c.ctx, c.cancel = context.WithCancel(context.Background())
	nc := NewContext(c, c.conn)
	go nc.process(c.ctx)

	return nil
}

func (c *TCPClient) Stop() error {
	if c.ctx != nil {
		c.cancel()
		c.ctx = nil
	}

	return nil
}

func (c *TCPClient) WaitForDone() {
	<-c.ctx.Done()
}
