package net

import "context"

// Service represents a network service object.
type Service interface {
	pipeline() *pipeline
	context() context.Context
	cancelFunc() context.CancelFunc
}
