package net

import (
	"context"
	"time"
)

type tcpChildService struct {
	s          *TCPServer
	cancelFunc context.CancelFunc
	doneCh     <-chan struct{}
}

func (cs *tcpChildService) pipeline() *tcpPipeline {
	return cs.s.pl
}

func (cs *tcpChildService) readTimeout() time.Duration {
	return cs.s.readTimeoutDur
}

func (cs *tcpChildService) writeTimeout() time.Duration {
	return cs.s.writeTimeoutDur
}

func (cs *tcpChildService) cancel() {
	cs.cancelFunc()
}

func (cs *tcpChildService) done() <-chan struct{} {
	return cs.doneCh
}

func (cs *tcpChildService) isRunning() bool {
	if cs.doneCh == nil {
		return false
	}

	select {
	case <-cs.doneCh:
		return false
	default:
		return true
	}
}
