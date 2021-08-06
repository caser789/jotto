package main

import (
	"reflect"

	"github.com/caser789/jotto/example/processors"
	pb "github.com/caser789/jotto/example/protocol"
	"github.com/caser789/jotto/jotto"
)

var r = jotto.NewRoute

func RequestId(ctx *jotto.Context, next func(*jotto.Context) error) (err error) {
	msg := reflect.ValueOf(ctx.Message)
	f := reflect.Indirect(msg).FieldByName("RequestId")

	next(ctx)

	reply := reflect.ValueOf(ctx.Reply)
	reflect.Indirect(reply).FieldByName("RequestId").Set(f)

	return
}

func Tag(ctx *jotto.Context, next func(*jotto.Context) error) (err error) {

	next(ctx)

	ctx.ResponseWritter.Header().Set("X-UPPER-ID", "Upper - Powered by Jotto")

	return
}

var web = []jotto.Middleware{
	RequestId,
	Tag,
}

var Routes = map[jotto.Route]*jotto.Processor{
	r(uint32(pb.MSG_KIND_REQ_ABOUT), "POST", "/v1/about"): &jotto.Processor{&pb.ReqAbout{}, &pb.RespAbout{}, processors.About, web},
	r(uint32(pb.MSG_KIND_REQ_TEXT), "POST", "/v1/text"):   &jotto.Processor{&pb.ReqText{}, &pb.RespText{}, processors.Text, web},
}
