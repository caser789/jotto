package processors

import (
	pb "github.com/caser789/jotto/example/protocol"
	"github.com/caser789/jotto/jotto"
	"github.com/gogo/protobuf/proto"
)

func About(ctx *jotto.Context) {
	reply := ctx.Reply.(*pb.RespAbout)

	reply.About = proto.String("This is an example application written in Jotto, the microservice framework.")
	ctx.ReplyKind = uint32(pb.MSG_KIND_RESP_ABOUT)
}
