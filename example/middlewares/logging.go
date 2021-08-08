package middlewares

import (
	"context"

	"git.garena.com/duanzy/motto/motto"
	"git.garena.com/lixh/goorm"
)

func Logging(ctx context.Context, app motto.Application, request, response interface{}, next motto.MiddlewareChainer) (int32, context.Context) {
	logger := motto.GetLogger(ctx)

	logger.Infof("Request: %v", request)

	goorm.RegisterLogFunction(func(format string, v ...interface{}) {
		logger.Debugf(format, v...)
	}, true)

	code, ctx := next(ctx)

	logger.Infof("Response: %v, code: %d", response, code)
	return code, ctx
}
