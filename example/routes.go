package main

import (
	"fmt"
	"reflect"

	"github.com/caser789/jotto/example/processors"
	pb "github.com/caser789/jotto/example/protocol"
	"github.com/caser789/jotto/jotto"
)

var r = motto.NewRoute

func Logging(app motto.Application, ctx *motto.Context, next func(*motto.Context) error) (err error) {
	fmt.Println("logging", ctx.Request)
	err = next(ctx)
	fmt.Println("logging", ctx.Reply)
	return
}

func RequestId(app motto.Application, ctx *motto.Context, next func(*motto.Context) error) (err error) {
	msg := reflect.ValueOf(ctx.Message)
	f := reflect.Indirect(msg).FieldByName("RequestId")

	next(ctx)

	reply := reflect.ValueOf(ctx.Reply)
	reflect.Indirect(reply).FieldByName("RequestId").Set(f)

	return
}

func Tag(app motto.Application, ctx *motto.Context, next func(*motto.Context) error) (err error) {

	if app.Protocol() == motto.HTTP {
		fmt.Printf("tag: %+v\n", ctx)

		ctx.ResponseWritter.Header().Set("X-UPPER-ID", "Upper - Powered by Motto")
	}

	return next(ctx)
}

var web = []motto.Middleware{
	Logging,
	RequestId,
	Tag,
}

var Routes = map[motto.Route]*motto.Processor{
	r(uint32(pb.MSG_KIND_REQ_ABOUT), "POST", "/v1/about"): &motto.Processor{&pb.ReqAbout{}, &pb.RespAbout{}, processors.About, web},
	r(uint32(pb.MSG_KIND_REQ_TEXT), "POST", "/v1/text"):   &motto.Processor{&pb.ReqText{}, &pb.RespText{}, processors.Text, web},
}
