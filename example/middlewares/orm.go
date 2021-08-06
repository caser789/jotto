package middlewares

import (
	"fmt"

	"git.garena.com/duanzy/motto/sample/common"

	"git.garena.com/duanzy/motto/motto"
)

func Orm(app motto.Application, context motto.Context, next func(motto.Context) error) (err error) {
	ctx := common.Ctx(context)

	if ctx.Orm != nil && ctx.Orm.NeedTrx() {
		ctx.Orm.Begin()
	}

	err = next(context)

	if ctx.Orm != nil && ctx.Orm.NeedTrx() {
		if err = ctx.Orm.Commit(); err != nil {
			fmt.Println("Txn commit error", err)
		}
	}
	return
}
