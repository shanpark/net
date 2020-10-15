package net

import "errors"

// ReadHandler is the interface that wraps the Read event handler method.
type ReadHandler interface {
	OnRead(ctx *Context, in interface{}) (interface{}, error)
}

// WriteHandler is the interface that wraps the Write event handler method.
type WriteHandler interface {
	OnWrite(ctx *Context, out interface{}) (interface{}, error)
}

// ConnectHandler is the interface that wraps the Connect event handler method.
type ConnectHandler interface {
	OnConnect(ctx *Context) error
}

// DisconnectHandler is the interface that wraps the Disconnect event handler method.
type DisconnectHandler interface {
	OnDisconnect(ctx *Context)
}

// TimeoutHandler is the interface that wraps the Timeout event handler method.
type TimeoutHandler interface {
	OnTimeout(ctx *Context) error
}

// ErrorHandler is the interface that wraps the Error event handler method.
type ErrorHandler interface {
	OnError(ctx *Context, err error)
}

type pipeline struct {
	readHandlers       []ReadHandler
	writeHandlers      []WriteHandler
	connectHandlers    []ConnectHandler
	disconnectHandlers []DisconnectHandler
	timeoutHandlers    []TimeoutHandler
	errorHandlers      []ErrorHandler
}

func (pl *pipeline) AddHandler(handler interface{}) error {
	var ok bool

	added := false
	if _, ok = handler.(ReadHandler); ok {
		pl.readHandlers = append(pl.readHandlers, handler.(ReadHandler))
		added = true
	}
	if _, ok = handler.(WriteHandler); ok {
		pl.writeHandlers = prependWriteHandler(pl.writeHandlers, handler.(WriteHandler))
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

func prependWriteHandler(writeHandlers []WriteHandler, writeHandler WriteHandler) []WriteHandler {
	writeHandlers = append(writeHandlers, nil)
	copy(writeHandlers[1:], writeHandlers)
	writeHandlers[0] = writeHandler
	return writeHandlers
}
