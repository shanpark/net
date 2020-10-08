package net

type Handler func(in interface{}) interface{}

type Pipeline struct {
	handlers []Handler
}

func (pl *Pipeline) AddHandler(handler Handler) {
	pl.handlers = append(pl.handlers, handler)
}

func Inbound(in interface{}) (interface{}, bool) {

}

func Outbound(out interface{}) (interface{}, bool) {

}
