package jotto

import (
	"context"
)

// Middleware is a function that wraps around the main logic, allowing you to do
// pre/post-processing before/after the main logic is being executed.
//
// Middlwares can be nested, i.e. you can wrap a middleware around another to form
// a middlware stack. The following is an illustration of what a middleware stack of
// two middlewares looks like.
//
//            Request
// 		         ↓
// +----------------------------+
// |          Logging           |
// |  +----------------------+  |
// |  |    Authentication    |  |
// |  |  +----------------+  |  |
// |  |  |   Main Logic   |  |  |
// |  |  +----------------+  |  |
// |  |                      |  |
// |  +----------------------+  |
// |                            |
// +----------------------------+
// 		         ↓
//            Response
//
// The above behavior is achieved by doing the following (in order):
// 1. The middleware executes its pre-processing logic;
// 2. The middleware calls `next`, passing control to the next layer of the onion;
// 3. After the `next` function returns, the middleware executes its post-processing logic.
//
// There're a few things worth noting here:
// * The outmost middleware is called first, but it will be the last to return;
// * A middleware can skip the rest of the middlewares and the main logic by returning early (before calling `next`)
// * When `next` returns, the main logic is either executed, or skipped.
type Middleware func(ctx context.Context, app Application, request, response interface{}, next MiddlewareChainer) (int32, context.Context)

// MiddlewareChainer is a function that chains two middlewares together.
type MiddlewareChainer func(context.Context) (code int32, ctx context.Context)
