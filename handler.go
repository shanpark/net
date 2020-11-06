package net

// ReadHandler is the interface that wraps the Read event handler method.
type ReadHandler interface {
	OnRead(ctx *SoContext, in interface{}) (interface{}, error)
}

// WriteHandler is the interface that wraps the Write event handler method.
type WriteHandler interface {
	OnWrite(ctx *SoContext, out interface{}) (interface{}, error)
}

// ConnectHandler is the interface that wraps the Connect event handler method.
type ConnectHandler interface {
	OnConnect(ctx *SoContext) error
}

// DisconnectHandler is the interface that wraps the Disconnect event handler method.
// OnDisconnect method is the only method can be called after the service is stopped.
type DisconnectHandler interface {
	OnDisconnect(ctx *SoContext)
}

// TimeoutHandler is the interface that wraps the Timeout event handler method.
type TimeoutHandler interface {
	OnTimeout(ctx *SoContext) error
}

// ErrorHandler is the interface that wraps the Error event handler method.
type ErrorHandler interface {
	OnError(ctx *SoContext, err error)
}
