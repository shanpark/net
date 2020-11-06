package net

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
)

// A TLSClient represents a client object using tcp network.
type TLSClient struct {
	TCPClient
	config *tls.Config
}

// NewTLSClient create a new TLSClient.
func NewTLSClient(config *tls.Config) *TLSClient {
	client := new(TLSClient)
	client.config = config
	client.pl = new(pipeline)
	client.AddHandler(client.optHandler)
	return client
}

// Start starts the service. TCPClient connects to the remote address.
func (c *TLSClient) Start() error {
	if c.isRunning() {
		return errors.New("net: tls client object is already started")
	}

	conn, err := net.Dial("tcp", c.address)
	if err != nil {
		return fmt.Errorf("net: Dial() failed - %v", err)
	}

	conn = tls.Client(conn, c.config) // only difference from TCP

	var cctx context.Context
	cctx, c.cancelFunc = context.WithCancel(context.Background())
	c.doneCh = cctx.Done()
	nctx := newContext(c, conn, defaultQueueSize)
	go nctx.process()
	c.nctx = nctx

	return nil
}
