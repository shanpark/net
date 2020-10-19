package net

import (
	"context"
	"time"
)

type childService struct {
	s          *TCPServer
	cctx       context.Context
	cancelFunc context.CancelFunc
	doneCh     <-chan struct{}
}

func (s *TCPServer) newChildService(cctx context.Context, cancelFunc context.CancelFunc) *childService {
	child := new(childService)
	child.s = s
	child.cctx = cctx
	child.cancelFunc = cancelFunc
	child.doneCh = cctx.Done()
	return child
}

func (cs *childService) context() context.Context {
	return cs.cctx
}

func (cs *childService) cancel() {
	cs.cctx = nil
	cs.cancelFunc()
}

func (cs *childService) pipeline() *pipeline {
	return cs.s.pl
}

func (cs *childService) readTimeout() time.Duration {
	return cs.s.readTimeoutDur
}

func (cs *childService) writeTimeout() time.Duration {
	return cs.s.writeTimeoutDur
}

func (cs *childService) done() <-chan struct{} {
	return cs.doneCh
}
