package net

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
)

type Service interface {
	Pipeline() *Pipeline
}

type TCPServer struct {
	address  string
	pipeline *Pipeline
	listener net.Listener
	ctx      context.Context
	cancel   context.CancelFunc
	err      error
}

// NewTCPServer returns a TCPServer object.
func NewTCPServer() *TCPServer {
	server := new(TCPServer)
	return server
}

func (s *TCPServer) SetAddress(address string) error {
	s.address = address
	return nil
}

func (s *TCPServer) SetPipeline(pipeline *Pipeline) {
	s.pipeline = pipeline
}

func (s *TCPServer) Pipeline() *Pipeline {
	return s.pipeline
}

func (s *TCPServer) Start() error {
	if s.ctx != nil {
		return errors.New("net: server object is already started")
	}

	s.ctx, s.cancel = context.WithCancel(context.Background())
	go s.listen(s.ctx)

	return nil
}

func (s *TCPServer) Stop() error {
	if s.ctx != nil {
		s.cancel()
		s.ctx = nil
	}

	return nil
}

func (s *TCPServer) WaitForDone() {
	<-s.ctx.Done()
}

func (s *TCPServer) listen(ctx context.Context) {
	var err error

	s.listener, err = net.Listen("tcp", s.address)
	if err != nil {
		s.handleErr(err)
		return
	}

	defer s.listener.Close()

	connCh := make(chan net.Conn)
	errCh := make(chan error)

	for s.err == nil {
		go s.accept(ctx, connCh, errCh)

		select {
		case <-ctx.Done():
			return
		case conn := <-connCh:
			s.handleConn(conn)
		case err = <-errCh:
			s.handleErr(err)
		}
	}
}

func (s *TCPServer) accept(ctx context.Context, connCh chan net.Conn, errCh chan error) {
	conn, err := s.listener.Accept()
	if err != nil {
		errCh <- err
	}
	connCh <- conn
}

func (s *TCPServer) handleConn(conn net.Conn) {
	nc := NewContext(s, conn)
	go nc.process(s.ctx)
}

func (s *TCPServer) handleErr(err error) {
	switch {
	case err.(net.Error).Temporary():
		fmt.Fprintf(os.Stdout, "temporary error")
	case err.(net.Error).Timeout():
		fmt.Fprintf(os.Stdout, "timeout error")
	default:
		s.err = err
		s.Stop()
	}
}
