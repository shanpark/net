package net

import (
	"context"
	"errors"
	"net"
)

type Service interface {
	Pipeline() *Pipeline
	Stop() error
}

// A TCPServer represents a server object using tcp network.
type TCPServer struct {
	address  string
	pipeline *Pipeline
	listener net.Listener
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewTCPServer create a new TCPServer.
func NewTCPServer() *TCPServer {
	server := new(TCPServer)
	server.pipeline = NewPipeline()
	return server
}

// SetAddress sets address for binding. The address has the form "host:port".
func (s *TCPServer) SetAddress(address string) error {
	s.address = address
	return nil
}

// AddHandler adds a handler for network events. A handler should implement at least one of interfaces ReadHandler, WriteHandler, ConnectHandler, DisconnectHandler, ErrorHandler.
// Handlers are called in order and only in the case of WriteHandler are called in the reverse direction.
func (s *TCPServer) AddHandler(handlers ...interface{}) error {
	for handler := range handlers {
		if err := s.pipeline.AddHandler(handler); err != nil {
			return err
		}
	}

	return nil
}

func (s *TCPServer) Pipeline() *Pipeline {
	return s.pipeline
}

// Start starts the service. TCPServer binds to the address and can receive connection request.
func (s *TCPServer) Start() error {
	if s.ctx != nil {
		return errors.New("net: server object is already started")
	}

	var err error

	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.listener, err = net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	go s.process(s.ctx)
	return nil
}

// Stop stops the service. TCPServer closes all connections
func (s *TCPServer) Stop() error {
	if s.ctx != nil {
		s.cancel()
		s.ctx = nil
	}

	return nil
}

// WaitForDone blocks until service ends.
func (s *TCPServer) WaitForDone() {
	<-s.ctx.Done()
}

func (s *TCPServer) process(ctx context.Context) {
	defer s.listener.Close()

	var err error

	connCh := make(chan net.Conn)
	errCh := make(chan error)
	go s.accept(ctx, connCh, errCh)

	for {
		select {
		case <-ctx.Done():
			return

		case conn := <-connCh:
			nc := NewContext(s, conn)
			go nc.process(s.ctx)

		case err = <-errCh:
			switch {
			case err.(net.Error).Temporary():
				continue
			case err.(net.Error).Timeout():
				continue
			default:
				s.Stop()
			}
		}
	}
}

func (s *TCPServer) accept(ctx context.Context, connCh chan net.Conn, errCh chan error) {
	defer close(connCh)
	defer close(errCh)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				errCh <- err
			}
		}
		connCh <- conn
	}
}
