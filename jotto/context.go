package jotto

import (
	"context"
	"net/http"
	"time"
)

// ContextKey - motto context key type
type ContextKey int

// Motto context keys
const (
	CtxHTTPRequest ContextKey = iota
	CtxHTTPResponse
	CtxHTTPRequestBody
	CtxHTTPResponseBody
	CtxHTTPStatus
	CtxHTTPResponseHeaders
	CtxLogger
	CtxTime
)

// GetLogger - retrieve a logger from context
func GetLogger(ctx context.Context) (logger Logger) {
	logger, ok := ctx.Value(CtxLogger).(Logger)

	if !ok {
		return NewStdoutLogger(nil)
	}

	return
}

// GetTime - get the time when the context is created
func GetTime(ctx context.Context) (timestamp uint32) {
	timestamp, ok := ctx.Value(CtxTime).(uint32)

	if !ok {
		return uint32(time.Now().Unix())
	}

	return timestamp
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

func GetHTTPBody(ctx context.Context) (body []byte) {
	body, ok := ctx.Value(CtxHTTPRequestBody).([]byte)

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
type ContextFactory func(context.Context, Processor) context.Context
