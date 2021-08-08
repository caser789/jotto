package middlewares

import (
	"context"
	"reflect"

	"git.garena.com/duanzy/motto/motto"
)

func RequestId(ctx context.Context, app motto.Application, request, response interface{}, next motto.MiddlewareChainer) (int32, context.Context) {
	msg := reflect.ValueOf(request)
	f := reflect.Indirect(msg).FieldByName("RequestId")

	code, ctx := next(ctx)

	reply := reflect.ValueOf(response)
	reflect.Indirect(reply).FieldByName("RequestId").Set(f)

	return code, ctx
}
