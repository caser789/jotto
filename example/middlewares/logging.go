package middlewares

import (
	"fmt"

	"github.com/caser789/jotto/jotto"
)

func Logging(app motto.Application, context motto.Context, next func(motto.Context) error) (err error) {
	ctx := context.Motto()
	fmt.Println("logging", ctx.Message)
	err = next(context)
	fmt.Println("logging", ctx.Reply)
	return
}
