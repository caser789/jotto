package common

import (
	"fmt"

	"git.garena.com/duanzy/motto/motto"
	pb "git.garena.com/duanzy/motto/sample/protocol"
	"git.garena.com/lixh/goorm"
)

func Boot(app interface{}) {
	goorm.Pls_Go_Get_Orm_Lib_V2_40()
	goorm.RegisterLogFunction(func(format string, v ...interface{}) {
		fmt.Printf(format+"\n", v...)
	}, false)

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
}

func ContextFactory(processor motto.Processor, ctx *motto.BaseContext) motto.Context {
	p := processor.(*Processor)

	var orm *goorm.Orm

	if p.Orm != nil {
		orm = goorm.NewOrmWithFlag(p.Orm.Database, p.Orm.Flag)
	}

	return &Context{
		MottoCtx: ctx,
		Orm:      orm,
	}
}
