package routes

import (
	"github.com/caser789/jotto/example/middlewares"
	"github.com/caser789/jotto/example/processors"
	pb "github.com/caser789/jotto/example/protocol"
	"github.com/caser789/jotto/jotto"
)

var r = motto.NewRoute

var web = []motto.Middleware{
	middlewares.Logging,
	middlewares.RequestId,
	middlewares.Tag,
}

var Routes = map[motto.Route]*motto.Processor{
	r(uint32(pb.MSG_KIND_REQ_ABOUT), "POST", "/v1/about"): &motto.Processor{&pb.ReqAbout{}, &pb.RespAbout{}, processors.About, web},
	r(uint32(pb.MSG_KIND_REQ_TEXT), "POST", "/v1/text"):   &motto.Processor{&pb.ReqText{}, &pb.RespText{}, processors.Text, web},
}
