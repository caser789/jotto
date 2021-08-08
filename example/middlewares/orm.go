package middlewares

import (
	"context"

	"git.garena.com/duanzy/motto/sample/common"

	"git.garena.com/duanzy/motto/motto"
)

func Orm(ctx context.Context, app motto.Application, request, response interface{}, next motto.MiddlewareChainer) (int32, context.Context) {
	orm := common.GetContextOrm(ctx)

	if orm == nil {
		return next(ctx)
	}

	logger := motto.GetLogger(ctx)

	if orm != nil && orm.NeedTrx() {
		orm.Begin()
	}

	code, ctx := next(ctx)

	if orm != nil && orm.NeedTrx() {
		if code == 0 {
			if err := orm.Commit(); err != nil {
				logger.Errorf("Txn commit error", err)
			}
		} else {
			orm.Rollback()
		}
	}
	return code, ctx
}
