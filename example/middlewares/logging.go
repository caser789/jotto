package middlewares

import (
	"fmt"

	"github.com/caser789/jotto/jotto"
)

func Logging(app motto.Application, ctx *motto.Context, next func(*motto.Context) error) (err error) {
	fmt.Println("logging", ctx.Message)
	err = next(ctx)
	fmt.Println("logging", ctx.Reply)
	return
}
