package processors

import (
	"time"

	"git.garena.com/duanzy/motto/motto"
	"git.garena.com/duanzy/motto/sample/common"
	pb "git.garena.com/duanzy/motto/sample/protocol"
)

func Wait(app motto.Application, c motto.Context) {
	context := common.Ctx(c)    // The custom context we defined in this application
	mottoCtx := context.Motto() // The Motto base context

	req := mottoCtx.Message.(*pb.ReqWait)

	time.Sleep(time.Duration(req.GetSeconds()) * time.Second)

	mottoCtx.ReplyKind = uint32(pb.MSG_KIND_RESP_WAIT)
}
