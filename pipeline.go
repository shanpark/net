package net

import "errors"

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

func (pl *pipeline) prependWriteHandler(writeHandler WriteHandler) {
	pl.writeHandlers = append(pl.writeHandlers, nil)
	copy(pl.writeHandlers[1:], pl.writeHandlers)
	pl.writeHandlers[0] = writeHandler
}
