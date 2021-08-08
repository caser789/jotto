package routes

import (
	"git.garena.com/duanzy/motto/motto"
	"git.garena.com/duanzy/motto/sample/common"
	"git.garena.com/duanzy/motto/sample/middlewares"
	"git.garena.com/duanzy/motto/sample/processors"
	pb "git.garena.com/duanzy/motto/sample/protocol"
	"git.garena.com/lixh/goorm"
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
	r(uint32(pb.MSG_KIND_REQ_ABOUT), "POST", "/v1/about", "main"): common.NewProcessor(&pb.ReqAbout{}, &pb.RespAbout{}, processors.About, web, MainShared),
	r(uint32(pb.MSG_KIND_REQ_TEXT), "POST", "/v1/text", "main"):   common.NewProcessor(&pb.ReqText{}, &pb.RespText{}, processors.Text, web, MainShared),
	r(uint32(pb.MSG_KIND_REQ_WAIT), "POST", "/v1/wait", "main"):   common.NewProcessor(&pb.ReqWait{}, &pb.RespWait{}, processors.Wait, web, MainShared),
}
