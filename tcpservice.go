package net

import (
	"time"
)

type tcpService interface {
	pipeline() *tcpPipeline
	readTimeout() time.Duration
	writeTimeout() time.Duration
	cancel()
	done() <-chan struct{}
	isRunning() bool
}
