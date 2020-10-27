package net

import "errors"

// ReadHandler is the interface that wraps the Read event handler method.
type SoReadHandler interface {
	OnRead(ctx *SoContext, in interface{}) (interface{}, error)
}

// WriteHandler is the interface that wraps the Write event handler method.
type SoWriteHandler interface {
	OnWrite(ctx *SoContext, out interface{}) (interface{}, error)
}

// ConnectHandler is the interface that wraps the Connect event handler method.
type SoConnectHandler interface {
	OnConnect(ctx *SoContext) error
}

// DisconnectHandler is the interface that wraps the Disconnect event handler method.
// OnDisconnect method is the only method can be called after the service is stopped.
type SoDisconnectHandler interface {
	OnDisconnect(ctx *SoContext)
}

// TimeoutHandler is the interface that wraps the Timeout event handler method.
type SoTimeoutHandler interface {
	OnTimeout(ctx *SoContext) error
}

// ErrorHandler is the interface that wraps the Error event handler method.
type SoErrorHandler interface {
	OnError(ctx *SoContext, err error)
}

type soPipeline struct {
	readHandlers       []SoReadHandler
	writeHandlers      []SoWriteHandler
	connectHandlers    []SoConnectHandler
	disconnectHandlers []SoDisconnectHandler
	timeoutHandlers    []SoTimeoutHandler
	errorHandlers      []SoErrorHandler
}

func (pl *soPipeline) AddHandler(handler interface{}) error {
	var ok bool

	added := false
	if _, ok = handler.(SoReadHandler); ok {
		pl.readHandlers = append(pl.readHandlers, handler.(SoReadHandler))
		added = true
	}
	if _, ok = handler.(SoWriteHandler); ok {
		pl.prependWriteHandler(handler.(SoWriteHandler))
		added = true
	}
	if _, ok = handler.(SoConnectHandler); ok {
		pl.connectHandlers = append(pl.connectHandlers, handler.(SoConnectHandler))
		added = true
	}
	if _, ok = handler.(SoDisconnectHandler); ok {
		pl.disconnectHandlers = append(pl.disconnectHandlers, handler.(SoDisconnectHandler))
		added = true
	}
	if _, ok = handler.(SoTimeoutHandler); ok {
		pl.timeoutHandlers = append(pl.timeoutHandlers, handler.(SoTimeoutHandler))
		added = true
	}
	if _, ok = handler.(SoErrorHandler); ok {
		pl.errorHandlers = append(pl.errorHandlers, handler.(SoErrorHandler))
		added = true
	}

	if !added {
		return errors.New("net: invalid handler - no handler method not found")
	}

	return nil
}

func (pl *soPipeline) prependWriteHandler(writeHandler SoWriteHandler) {
	pl.writeHandlers = append(pl.writeHandlers, nil)
	copy(pl.writeHandlers[1:], pl.writeHandlers)
	pl.writeHandlers[0] = writeHandler
}
