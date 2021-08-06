package processors

import (
	"fmt"

	"github.com/caser789/jotto/example/common"
	pb "github.com/caser789/jotto/example/protocol"
	"github.com/caser789/jotto/jotto"
	"github.com/gogo/protobuf/proto"
)

func About(app motto.Application, context motto.Context) {
	ctx := context.Motto()
	reply := ctx.Reply.(*pb.RespAbout)

	reply.About = proto.String(fmt.Sprintf("(%s) This is an example application written in Motto, the microservice framework.", common.Cfg(app).Country))
	ctx.ReplyKind = uint32(pb.MSG_KIND_RESP_ABOUT)
}
