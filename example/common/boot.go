package common

import (
	"context"
	"fmt"
	"reflect"

	"git.garena.com/common/gocommon"
	"git.garena.com/duanzy/motto/motto"
	pb "git.garena.com/duanzy/motto/sample/protocol"
	"git.garena.com/lixh/goorm"
	"github.com/golang/protobuf/proto"
)

func Boot(payloads ...interface{}) {
	// Initialise the logger
	gocommon.LoggerInit("log/upper.log", 86400, 1000*1000*1000, 30, 3)

	app := payloads[0].(motto.Application)

	logger := app.MakeLogger(nil)

	goorm.Pls_Go_Get_Orm_Lib_V2_40()
	goorm.RegisterLogFunction(func(format string, v ...interface{}) {
		logger.Debugf(format, v...)
	}, true)

	mi := goorm.RegisterModel(&pb.Quote{})
	mi.SetPrimayColumn("id", false)

	cfg := []goorm.OrmDBConfig{
		goorm.OrmDBConfig{
			MasterDSN: "root:@tcp(127.0.0.1:3306)/upper",
			SlaveDSN:  []string{"root:@tcp(127.0.0.1:3306)/upper"},
			DBName:    "upper",
			MaxConn:   100,
			MaxIdle:   100,
		},
	}

	if err := goorm.RegisterOrmFromConfig(cfg, "upper"); err != nil {
		fmt.Println("Error setting up ORM: ", err)
		return
	}

	app.SetPanicHandler(Panic)

}

func Panic(ctx context.Context, app motto.Application, recover, request, response interface{}) {
	reply := reflect.ValueOf(response)
	reflect.Indirect(reply).FieldByName("RequestId").Set(reflect.ValueOf(proto.String(fmt.Sprintf("error:%v", recover))))
}

func Reload(payloads ...interface{}) {
	fmt.Println("Application is reloaded, do something")
}

func Terminate(payloads ...interface{}) {
	fmt.Println("Application is terminating")
}

func ContextFactory(ctx context.Context, processor motto.Processor) context.Context {
	p := processor.(*Processor)

	var orm *goorm.Orm

	if p.Orm != nil {
		orm = goorm.NewOrmWithFlag(p.Orm.Database, p.Orm.Flag)
	}

	return context.WithValue(ctx, CtxOrm, orm)
}
