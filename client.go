package net

import (
	"context"
	"errors"
	"fmt"
	"net"
)

type TCPClient struct {
	address string
	conn    net.Conn
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewTCPClient() *TCPClient {
	client := new(TCPClient)
	return client
}

func (c *TCPClient) SetAddress(address string) error {
	c.address = address
	return nil
}

func (c *TCPClient) Start() error {
	if c.ctx != nil {
		return errors.New("net: client object is already started")
	}

	c.ctx, c.cancel = context.WithCancel(context.Background())
	go c.connect(c.ctx)

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

func (c *TCPClient) connect(ctx context.Context) {
	var err error

	c.conn, err = net.Dial("tcp", c.address)
	if err != nil {
		return // TODO 에러처리
	}

	data := make([]byte, 1024)
	c.conn.Write([]byte("Hello"))
	n, _ := c.conn.Read(data)
	fmt.Println(string(data[:n]))
	defer c.conn.Close()
}
