package net

import "time"

// SoService represents a stream oriented network service object.
type SoService interface {
	SetAddress(address string) error
	AddHandler(handlers ...interface{}) error
	Start() error
	Stop() error
	WaitForDone()
}

type soAcceptor interface {
	acceptLoop()
}

type soProperty interface {
	pipeline() *soPipeline
	readTimeout() time.Duration
	writeTimeout() time.Duration
}

// Stream oriented service
type soObject interface {
	soProperty
	cancel()
	done() <-chan struct{}
	isRunning() bool
}
