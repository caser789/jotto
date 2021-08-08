package middlewares

import (
	"context"

	"git.garena.com/duanzy/motto/motto"
	"git.garena.com/duanzy/motto/sample/common"
)

func Tag(ctx context.Context, app motto.Application, request, response interface{}, next motto.MiddlewareChainer) (int32, context.Context) {
	httpResponse := motto.GetHTTPResponse(ctx)

	if httpResponse != nil {
		httpResponse.Header().Set("X-APP-COUNTRY", common.Cfg(app).Country)
		httpResponse.Header().Set("X-APP-NAME", "Upper - Powered by Motto")
	}

	return next(ctx)
}
