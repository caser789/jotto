package jotto

import (
	"context"
	"net/http"
)

// ContextKey - motto context key type
type ContextKey int

// Motto context keys
const (
	CtxHTTPRequest ContextKey = iota
	CtxHTTPResponse
	CtxLogger
)

// GetLogger - retrieve a logger from context
func GetLogger(ctx context.Context) (logger Logger) {
	logger, ok := ctx.Value(CtxLogger).(Logger)

	if !ok {
		return NewStdoutLogger(nil)
	}

	return
}

func GetHTTPRequest(ctx context.Context) (request *http.Request) {
	request, ok := ctx.Value(CtxHTTPRequest).(*http.Request)

	if !ok {
		return nil
	}

	return
}

func GetHTTPResponse(ctx context.Context) (response http.ResponseWriter) {
	response, ok := ctx.Value(CtxHTTPResponse).(http.ResponseWriter)

	if !ok {
		return nil
	}

	return
}

// ContextFactory is a function that produces a context that conforms
// to the `Context` interface. The registered context factory (via `Application.SetContextFactory()`)
// will be called once a base context is initialized by Motto. This base
// context will be passed down to the context factory, so that the application
// can create their custom execution context based on it.
type ContextFactory func(Processor, context.Context) context.Context
