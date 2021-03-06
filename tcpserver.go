package net

import (
	"context"
	"errors"
	"net"
	"time"
)

// A TCPServer represents a server object using tcp network.
type TCPServer struct {
	address    string
	cctx       context.Context
	cancelFunc context.CancelFunc
	doneCh     <-chan struct{}

	listener net.Listener // for Server
	err      error        //

	pl              *pipeline         // for childService
	optHandler      tcpConnOptHandler //
	readTimeoutDur  time.Duration     //
	writeTimeoutDur time.Duration     //
}

// NewTCPServer create a new TCPServer.
func NewTCPServer() *TCPServer {
	server := new(TCPServer)
	server.pl = new(pipeline)
	server.AddHandler(server.optHandler)
	return server
}

// SetAddress sets address for binding. The address has the form "host:port".
func (s *TCPServer) SetAddress(address string) error {
	s.address = address
	return nil
}

// SetTimeout sets the read and write timeout associated with the connection.
func (s *TCPServer) SetTimeout(readTimeout time.Duration, writeTimeout time.Duration) error {
	s.readTimeoutDur = readTimeout
	s.writeTimeoutDur = writeTimeout
	return nil
}

// SetNoDelay controls whether the operating system should delay packet transmission in hopes of sending fewer packets (Nagle's algorithm).
// The default is true (no delay), meaning that data is sent as soon as possible after a Write.
func (s *TCPServer) SetNoDelay(noDelay bool) error {
	s.optHandler.noDelay = new(bool)
	*s.optHandler.noDelay = noDelay
	return nil
}

// SetKeepAlive sets whether the operating system should send keep-alive messages on the connection.
// If period is zero, default value will be used.
func (s *TCPServer) SetKeepAlive(keepAlive bool, period time.Duration) error {
	s.optHandler.keepAlive = new(bool)
	*s.optHandler.keepAlive = keepAlive
	s.optHandler.keepAlivePeriod = period
	return nil
}

// SetLinger sets the behavior of Close on a connection which still has data waiting to be sent or to be acknowledged.
// refer https://golang.org/pkg/net/#TCPConn.SetLinger
func (s *TCPServer) SetLinger(sec int) error {
	s.optHandler.linger = new(int)
	*s.optHandler.linger = sec
	return nil
}

// SetReadBuffer sets the size of the operating system's receive buffer associated with the connection.
func (s *TCPServer) SetReadBuffer(bytes int) error {
	s.optHandler.readBuffersize = new(int)
	*s.optHandler.readBuffersize = bytes
	return nil
}

// SetWriteBuffer sets the size of the operating system's transmit buffer associated with the connection.
func (s *TCPServer) SetWriteBuffer(bytes int) error {
	s.optHandler.writeBufferSize = new(int)
	*s.optHandler.writeBufferSize = bytes
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
	if s.isRunning() {
		return errors.New("net: server object is already started")
	}

	var err error

	s.cctx, s.cancelFunc = context.WithCancel(context.Background())
	s.doneCh = s.cctx.Done()
	s.listener, err = net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	go s.process(s)
	return nil
}

// Stop stops the service. TCPServer closes all connections
func (s *TCPServer) Stop() error {
	if s.isRunning() {
		s.cctx = nil
		s.cancelFunc()
	}

	return nil
}

// WaitForDone blocks until service stops.
func (s *TCPServer) WaitForDone() {
	<-s.doneCh
}

// Error returns an error that makes service stop.
// For normal stop, returns nil.
func (s *TCPServer) Error() error {
	return s.err
}

func (s *TCPServer) pipeline() *pipeline {
	return s.pl
}

func (s *TCPServer) readTimeout() time.Duration {
	return s.readTimeoutDur
}

func (s *TCPServer) writeTimeout() time.Duration {
	return s.writeTimeoutDur
}

func (s *TCPServer) process(acceptor soAcceptor) {
	defer s.listener.Close()

	go acceptor.acceptLoop()

	<-s.doneCh
}

func (s *TCPServer) acceptLoop() {
	for {
		select {
		case <-s.doneCh:
			return

		default:
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-s.doneCh:
				default:
					switch {
					case err.(net.Error).Temporary() || err.(net.Error).Timeout():
						continue
					default:
						s.err = err
						s.Stop()
					}
				}
				return
			}

			child := s.newChildService(context.WithCancel(s.cctx))
			nctx := newContext(child, conn, defaultQueueSize)
			go nctx.process()
		}
	}
}

func (s *TCPServer) newChildService(cctx context.Context, cancelFunc context.CancelFunc) *soChildService {
	child := new(soChildService)
	child.s = s
	child.cancelFunc = cancelFunc
	child.doneCh = cctx.Done()
	return child
}

func (s *TCPServer) isRunning() bool {
	if s.doneCh == nil {
		return false
	}

	select {
	case <-s.doneCh:
		return false
	default:
		return true
	}
}
