package middlewares

import (
	"git.garena.com/lixh/goorm"
	"github.com/caser789/jotto/jotto"
)

func Logging(app motto.Application, context motto.Context, next func(motto.Context) error) (err error) {
	ctx := context.Motto()
	ctx.Logger.Info("Request: %v", ctx.Message)

	goorm.RegisterLogFunction(func(format string, v ...interface{}) {
		ctx.Logger.Debug(format, v...)
	}, true)

	err = next(context)

	ctx.Logger.Info("Response: %v", ctx.Reply)
	return
}
