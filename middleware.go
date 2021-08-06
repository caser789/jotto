package motto

type Middleware func(ctx *Context, next func(*Context) error) error

func ExecuteProcessor(processor *Processor, ctx *Context, mids []Middleware) (err error) {
	if len(mids) == 0 {
		processor.Handler(ctx)
		return
	}

	return mids[0](ctx, func(c *Context) error {
		return ExecuteProcessor(processor, c, mids[1:])
	})
}
