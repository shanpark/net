package net

import (
	"context"
	"time"
)

type soChildService struct {
	s          soProperty
	cancelFunc context.CancelFunc
	doneCh     <-chan struct{}
}

func (cs *soChildService) pipeline() *pipeline {
	return cs.s.pipeline()
}

func (cs *soChildService) readTimeout() time.Duration {
	return cs.s.readTimeout()
}

func (cs *soChildService) writeTimeout() time.Duration {
	return cs.s.writeTimeout()
}

func (cs *soChildService) cancel() {
	cs.cancelFunc()
}

func (cs *soChildService) done() <-chan struct{} {
	return cs.doneCh
}

func (cs *soChildService) isRunning() bool {
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
