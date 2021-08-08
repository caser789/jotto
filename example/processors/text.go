package processors

import (
	"context"
	"strings"
	"time"

	"git.garena.com/duanzy/motto/motto"
	pb "git.garena.com/duanzy/motto/sample/protocol"
	"github.com/gogo/protobuf/proto"
)

func Text(ctx context.Context, app motto.Application, request, response interface{}) (int32, context.Context) {
	logger := motto.GetLogger(ctx)

	logger.Debugf("debug message from processor, time: %v", time.Now())

	message := request.(*pb.ReqText)
	reply := response.(*pb.RespText)

	text := message.GetText()

	words := strings.Split(text, " ")
	upper := []string{}

	for _, word := range words {
		upper = append(upper, strings.ToUpper(word))
	}

	text = strings.Join(upper, " ")

	reply.Text = proto.String(text)

	return int32(pb.MSG_KIND_RESP_TEXT), ctx
}
