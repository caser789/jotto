package jotto

import (
	"context"

	"github.com/gogo/protobuf/proto"
)

// ProcessorHandler is the basic logic unit of a Motto app.
type ProcessorHandler func(ctx context.Context, app Application, request, response interface{}) int32

// Processor specifies the logic (Handler) and middlewares (Middlewares) to be executed as a whole, as well as the input (Message) and output (Reply) formats.
// Applications can implement their own processor representations.
type Processor interface {
	Message() proto.Message
	Reply() proto.Message
	Handler() ProcessorHandler
	Middlewares() []Middleware
}

// NewProcessor creates a basic processor
func NewProcessor(message, reply proto.Message, handler ProcessorHandler, middlewares []Middleware) Processor {
	return &BaseProcessor{
		message:     message,
		reply:       reply,
		handler:     handler,
		middlewares: middlewares,
	}
}

// BaseProcessor is the built-in processor format of Motto
type BaseProcessor struct {
	message     proto.Message
	reply       proto.Message
	handler     ProcessorHandler
	middlewares []Middleware
}

// Message returns the input format of a processor
func (p *BaseProcessor) Message() proto.Message {
	return p.message
}

// Reply returns the output format of a processor
func (p *BaseProcessor) Reply() proto.Message {
	return p.reply
}

// Handler returns the ProcessorHandler associated with this processor
func (p *BaseProcessor) Handler() ProcessorHandler {
	return p.handler
}

// Middlewares returns the set of middlewares associated with this processor
func (p *BaseProcessor) Middlewares() []Middleware {
	return p.middlewares
}
