package net

import (
	"context"
	"time"
)

// Service represents a network service object.
type Service interface {
	context() context.Context
	cancel()
	pipeline() *pipeline
	readTimeout() time.Duration
	writeTimeout() time.Duration
}
