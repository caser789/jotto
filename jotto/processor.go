package jotto

import "github.com/gogo/protobuf/proto"

type ProcessorHandler func(Application, Context)

type Processor interface {
	Message() proto.Message
	Reply() proto.Message
	Handler() ProcessorHandler
	Middlewares() []Middleware
}

func NewProcessor(message, reply proto.Message, handler ProcessorHandler, middlewares []Middleware) Processor {
	return &BaseProcessor{
		message:     message,
		reply:       reply,
		handler:     handler,
		middlewares: middlewares,
	}
}

type BaseProcessor struct {
	message     proto.Message
	reply       proto.Message
	handler     ProcessorHandler
	middlewares []Middleware
}

func (p *BaseProcessor) Message() proto.Message {
	return p.message
}
func (p *BaseProcessor) Reply() proto.Message {
	return p.reply
}
func (p *BaseProcessor) Handler() ProcessorHandler {
	return p.handler
}
func (p *BaseProcessor) Middlewares() []Middleware {
	return p.middlewares
}
