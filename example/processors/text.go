package processors

import (
	"strings"

	pb "github.com/caser789/jotto/example/protocol"
	"github.com/caser789/jotto/jotto"
	"github.com/gogo/protobuf/proto"
)

func Text(app motto.Application, ctx *motto.Context) {
	message := ctx.Message.(*pb.ReqText)
	reply := ctx.Reply.(*pb.RespText)

	text := message.GetText()

	words := strings.Split(text, " ")
	upper := []string{}

	for _, word := range words {
		upper = append(upper, strings.ToUpper(word))
	}

	text = strings.Join(upper, " ")

	reply.Text = proto.String(text)

	ctx.ReplyKind = uint32(pb.MSG_KIND_RESP_TEXT)
}
