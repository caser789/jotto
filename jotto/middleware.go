package jotto

type Middleware func(app Application, ctx *Context, next func(*Context) error) error
