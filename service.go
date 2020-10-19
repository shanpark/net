package net

import (
	"time"
)

// Service represents a network service object.
type Service interface {
	pipeline() *pipeline
	readTimeout() time.Duration
	writeTimeout() time.Duration
	done() <-chan struct{}
	cancel()
}
