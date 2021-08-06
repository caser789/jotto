package processors

import (
	"fmt"

	"github.com/caser789/jotto/example/common"
	pb "github.com/caser789/jotto/example/protocol"
	"github.com/caser789/jotto/jotto"
	"github.com/gogo/protobuf/proto"
)

func About(app motto.Application, c motto.Context) {
	context := common.Ctx(c)    // The custom context we defined in this application
	mottoCtx := context.Motto() // The Motto base context

	// Example: access ORM from the custom context.
	quote := &pb.Quote{
		Id: proto.Int64(1),
	}
	context.Orm.Get(quote)
	fmt.Println(quote)

	reply := mottoCtx.Reply.(*pb.RespAbout)

	reply.About = proto.String(
		fmt.Sprintf(
			"(%s) This is an example application written in Motto. Quote: %s By: %s",
			common.Cfg(app).Country,
			quote.GetQuote(),
			quote.GetAuthor(),
		),
	)
	mottoCtx.ReplyKind = uint32(pb.MSG_KIND_RESP_ABOUT)
}
