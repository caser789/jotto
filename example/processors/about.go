package processors

import (
	"context"
	"fmt"

	"git.garena.com/duanzy/motto/motto"
	"git.garena.com/duanzy/motto/sample/common"
	pb "git.garena.com/duanzy/motto/sample/protocol"
	"github.com/gogo/protobuf/proto"
)

func About(ctx context.Context, app motto.Application, request, response interface{}) (int32, context.Context) {
	orm := common.GetContextOrm(ctx)

	// Example: access ORM from the custom context.
	quote := &pb.Quote{
		Id: proto.Int64(1),
	}
	orm.Get(quote)
	fmt.Println(quote)

	reply := response.(*pb.RespAbout)

	reply.About = proto.String(
		fmt.Sprintf(
			"(%s) This is an example application written in Motto. Quote: %s By: %s",
			common.Cfg(app).Country,
			quote.GetQuote(),
			quote.GetAuthor(),
		),
	)
	return int32(pb.MSG_KIND_RESP_ABOUT), ctx
}
