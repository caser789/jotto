package middlewares

import (
	"github.com/caser789/jotto/example/common"
	"github.com/caser789/jotto/jotto"
)

func Tag(app motto.Application, ctx *motto.Context, next func(*motto.Context) error) (err error) {

	if app.Protocol() == motto.HTTP {
		ctx.ResponseWritter.Header().Set("X-APP-COUNTRY", common.Cfg(app).Country)
		ctx.ResponseWritter.Header().Set("X-APP-NAME", "Upper - Powered by Motto")
	}

	return next(ctx)
}
