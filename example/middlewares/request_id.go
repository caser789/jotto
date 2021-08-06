package middlewares

import (
	"reflect"

	"github.com/caser789/jotto/jotto"
)

func RequestId(app motto.Application, context motto.Context, next func(motto.Context) error) (err error) {
	ctx := context.Motto()
	msg := reflect.ValueOf(ctx.Message)
	f := reflect.Indirect(msg).FieldByName("RequestId")

	next(context)

	reply := reflect.ValueOf(ctx.Reply)
	reflect.Indirect(reply).FieldByName("RequestId").Set(f)

	return
}
