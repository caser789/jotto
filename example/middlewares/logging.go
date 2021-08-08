package middlewares

import (
	"git.garena.com/duanzy/motto/motto"
	"git.garena.com/lixh/goorm"
)

func Logging(app motto.Application, context motto.Context, next func(motto.Context) error) (err error) {
	ctx := context.Motto()

	ctx.Logger.Infof("Request: %v", ctx.Message)

	goorm.RegisterLogFunction(func(format string, v ...interface{}) {
		ctx.Logger.Debugf(format, v...)
	}, true)

	err = next(context)

	ctx.Logger.Infof("Response: %v", ctx.Reply)
	return
}
