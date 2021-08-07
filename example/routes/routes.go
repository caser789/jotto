package routes

import (
	"git.garena.com/duanzy/motto/motto"
	"git.garena.com/duanzy/motto/sample/common"
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
	middlewares.Orm,
}

var (
	MainShared = &common.OrmSetting{"upper", goorm.Trx_ReadSLock}
)

var Routes = map[motto.Route]motto.Processor{
	r(uint32(pb.MSG_KIND_REQ_ABOUT), "POST", "/v1/about"): common.NewProcessor(&pb.ReqAbout{}, &pb.RespAbout{}, processors.About, web, MainShared),
	r(uint32(pb.MSG_KIND_REQ_TEXT), "POST", "/v1/text"):   common.NewProcessor(&pb.ReqText{}, &pb.RespText{}, processors.Text, web, MainShared),
	r(uint32(pb.MSG_KIND_REQ_WAIT), "POST", "/v1/wait"):   common.NewProcessor(&pb.ReqWait{}, &pb.RespWait{}, processors.Wait, web, MainShared),
}
