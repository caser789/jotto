package jotto

import (
	"net/http"

	"github.com/gogo/protobuf/proto"
)

type Context interface {
	Motto() *BaseContext
}

type BaseContext struct {
	// Proto messages
	MessageKind uint32
	ReplyKind   uint32
	Message     proto.Message
	Reply       proto.Message

	// HTTP messages
	Request         *http.Request
	ResponseWritter http.ResponseWriter

	Logger Logger
}

type ContextFactory func(Processor, *BaseContext) Context

func (c *BaseContext) Motto() *BaseContext {
	return c
}
