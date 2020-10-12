package net

type Inbounder interface {
	Inbound(ctx *Context, in interface{}) (interface{}, error)
}

type Outbounder interface {
	Outbound(ctx *Context, in interface{}) (interface{}, error)
}

type Handler interface {
	Inbounder
	Outbounder
}

type Pipeline struct {
	handlers []Handler
}

func (pl *Pipeline) AddHandler(handler Handler) {
	pl.handlers = append(pl.handlers, handler)
}

func (pl *Pipeline) Handlers() []Handler {
	return pl.handlers
}
