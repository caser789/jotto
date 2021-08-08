package processors

import (
	"context"
	"time"

	"git.garena.com/duanzy/motto/motto"
	pb "git.garena.com/duanzy/motto/sample/protocol"
)

func Wait(ctx context.Context, app motto.Application, request, response interface{}) (int32, context.Context) {
	req := request.(*pb.ReqWait)

	time.Sleep(time.Duration(req.GetSeconds()) * time.Second)

	return int32(pb.MSG_KIND_RESP_WAIT), ctx
}
