package net

import "errors"

// ReadHandler is the interface that wraps the Read event handler method.
type ReadHandler interface {
	OnRead(ctx *TCPContext, in interface{}) (interface{}, error)
}

// WriteHandler is the interface that wraps the Write event handler method.
type WriteHandler interface {
	OnWrite(ctx *TCPContext, out interface{}) (interface{}, error)
}

// ConnectHandler is the interface that wraps the Connect event handler method.
type ConnectHandler interface {
	OnConnect(ctx *TCPContext) error
}

// DisconnectHandler is the interface that wraps the Disconnect event handler method.
// OnDisconnect method is the only method can be called after the service is stopped.
type DisconnectHandler interface {
	OnDisconnect(ctx *TCPContext)
}

// TimeoutHandler is the interface that wraps the Timeout event handler method.
type TimeoutHandler interface {
	OnTimeout(ctx *TCPContext) error
}

// ErrorHandler is the interface that wraps the Error event handler method.
type ErrorHandler interface {
	OnError(ctx *TCPContext, err error)
}

type tcpPipeline struct {
	readHandlers       []ReadHandler
	writeHandlers      []WriteHandler
	connectHandlers    []ConnectHandler
	disconnectHandlers []DisconnectHandler
	timeoutHandlers    []TimeoutHandler
	errorHandlers      []ErrorHandler
}

func (pl *tcpPipeline) AddHandler(handler interface{}) error {
	var ok bool

	added := false
	if _, ok = handler.(ReadHandler); ok {
		pl.readHandlers = append(pl.readHandlers, handler.(ReadHandler))
		added = true
	}
	if _, ok = handler.(WriteHandler); ok {
		pl.prependWriteHandler(handler.(WriteHandler))
		added = true
	}
	if _, ok = handler.(ConnectHandler); ok {
		pl.connectHandlers = append(pl.connectHandlers, handler.(ConnectHandler))
		added = true
	}
	if _, ok = handler.(DisconnectHandler); ok {
		pl.disconnectHandlers = append(pl.disconnectHandlers, handler.(DisconnectHandler))
		added = true
	}
	if _, ok = handler.(TimeoutHandler); ok {
		pl.timeoutHandlers = append(pl.timeoutHandlers, handler.(TimeoutHandler))
		added = true
	}
	if _, ok = handler.(ErrorHandler); ok {
		pl.errorHandlers = append(pl.errorHandlers, handler.(ErrorHandler))
		added = true
	}

	if !added {
		return errors.New("net: invalid handler - no handler method not found")
	}

	return nil
}

func (pl *tcpPipeline) prependWriteHandler(writeHandler WriteHandler) {
	pl.writeHandlers = append(pl.writeHandlers, nil)
	copy(pl.writeHandlers[1:], pl.writeHandlers)
	pl.writeHandlers[0] = writeHandler
}
