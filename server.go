package net

import (
	"context"
	"errors"
	"net"
)

// Service represents a network service object.
type Service interface {
	Error() error

	pipeline() *pipeline
}

// A TCPServer represents a server object using tcp network.
type TCPServer struct {
	address  string
	pl       *pipeline
	listener net.Listener
	ctx      context.Context
	cancel   context.CancelFunc
	err      error
}

// NewTCPServer create a new TCPServer.
func NewTCPServer() *TCPServer {
	server := new(TCPServer)
	server.pl = new(pipeline)
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
	for _, handler := range handlers {
		if err := s.pl.AddHandler(handler); err != nil {
			return err
		}
	}

	return nil
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

// WaitForDone blocks until service stops.
func (s *TCPServer) WaitForDone() {
	<-s.ctx.Done()
}

// Error returns an error that makes service stop.
// For normal stop, returns nil.
func (s *TCPServer) Error() error {
	return s.err
}

func (s *TCPServer) pipeline() *pipeline {
	return s.pl
}

func (s *TCPServer) process(ctx context.Context) {
	defer s.listener.Close()

	var err error

	connCh, errCh := s.startAccept(ctx)

	for {
		select {
		case <-ctx.Done():
			return

		case conn := <-connCh:
			nc := newContext(s, conn)
			go nc.process(s.ctx)

		case err = <-errCh:
			switch {
			case err.(net.Error).Temporary() || err.(net.Error).Timeout():
				connCh, errCh = s.startAccept(ctx) // restart accept routine
				continue
			default:
				s.err = err
				s.Stop()
			}
		}
	}
}

func (s *TCPServer) startAccept(ctx context.Context) (<-chan net.Conn, <-chan error) {
	connCh := make(chan net.Conn)
	errCh := make(chan error)
	go s.accept(ctx, connCh, errCh)
	return connCh, errCh
}

func (s *TCPServer) accept(ctx context.Context, connCh chan<- net.Conn, errCh chan<- error) {
	defer close(connCh)
	defer close(errCh)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
			default:
				errCh <- err
			}
			return
		}
		connCh <- conn
	}
}
