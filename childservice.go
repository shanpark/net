package net

import (
	"context"
	"time"
)

type ChildService struct {
	s          *TCPServer
	cctx       context.Context
	cancelFunc context.CancelFunc
	doneCh     <-chan struct{}
}

func (s *TCPServer) newChildService(cctx context.Context, cancelFunc context.CancelFunc) *ChildService {
	child := new(ChildService)
	child.s = s
	child.cctx = cctx
	child.cancelFunc = cancelFunc
	child.doneCh = cctx.Done()
	return child
}

func (cs *ChildService) context() context.Context {
	return cs.cctx
}

func (cs *ChildService) cancel() {
	cs.cctx = nil
	cs.cancelFunc()
}

func (cs *ChildService) pipeline() *pipeline {
	return cs.s.pl
}

func (cs *ChildService) readTimeout() time.Duration {
	return cs.s.readTimeoutDur
}

func (cs *ChildService) writeTimeout() time.Duration {
	return cs.s.writeTimeoutDur
}

func (cs *ChildService) done() <-chan struct{} {
	return cs.doneCh
}
