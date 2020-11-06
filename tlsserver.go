package net

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
)

// A TLSServer represents a server object using tls network.
type TLSServer struct {
	TCPServer
	config *tls.Config
}

// NewTLSServer create a new TCPServer.
func NewTLSServer(config *tls.Config) *TLSServer {
	server := new(TLSServer)
	server.config = config
	server.pl = new(pipeline)
	server.AddHandler(server.optHandler)
	return server
}

// Start starts the service. TCPServer binds to the address and can receive connection request.
func (s *TLSServer) Start() error {
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

func (s *TLSServer) acceptLoop() {
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

			conn = tls.Server(conn, s.config) // only difference from TCP

			child := s.newChildService(context.WithCancel(s.cctx))
			nctx := newContext(child, conn, defaultQueueSize)
			go nctx.process()
		}
	}
}
