package net

import "errors"

type ReadHandler interface {
	OnRead(ctx *Context, in interface{}) (interface{}, error)
}

type WriteHandler interface {
	OnWrite(ctx *Context, out interface{}) (interface{}, error)
}

type ConnectHandler interface {
	OnConnect(ctx *Context) error
}

type DisconnectHandler interface {
	OnDisconnect(ctx *Context) error
}

type ErrorHandler interface {
	OnError(ctx *Context) error
}

type Pipeline struct {
	readHandlers       []ReadHandler
	writeHandlers      []WriteHandler
	connectHandlers    []ConnectHandler
	disconnectHandlers []DisconnectHandler
	errorHandlers      []ErrorHandler
}

func (pl *Pipeline) AddHandler(handler interface{}) error {
	var ok bool

	added := false
	if _, ok = handler.(ReadHandler); ok {
		pl.readHandlers = append(pl.readHandlers, handler.(ReadHandler))
		added = true
	}
	if _, ok = handler.(WriteHandler); ok {
		pl.writeHandlers = append(pl.writeHandlers, handler.(WriteHandler))
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
	if _, ok = handler.(ErrorHandler); ok {
		pl.errorHandlers = append(pl.errorHandlers, handler.(ErrorHandler))
		added = true
	}

	if !added {
		return errors.New("net: invalid handler - no handler method not found")
	}

	return nil
}

func (pl *Pipeline) ReadHandlers() []ReadHandler {
	return pl.readHandlers
}

func (pl *Pipeline) WriteHandlers() []WriteHandler {
	return pl.writeHandlers
}

func (pl *Pipeline) ConnectHandlers() []ConnectHandler {
	return pl.connectHandlers
}

func (pl *Pipeline) DisconnectHandlers() []DisconnectHandler {
	return pl.disconnectHandlers
}

func (pl *Pipeline) ErrorHandlers() []ErrorHandler {
	return pl.errorHandlers
}
