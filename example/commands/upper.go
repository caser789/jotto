package commands

import (
	"flag"
	"fmt"
	"net"
	"time"

	"git.garena.com/duanzy/motto/hotline"
	"git.garena.com/duanzy/motto/motto"
	pb "git.garena.com/duanzy/motto/sample/protocol"
	"github.com/golang/protobuf/proto"
)

type Quote struct {
	quote string
	name  string
}

// About calls the about API provided by the sample server
type Upper struct {
	motto.BaseCommand
	text string
}

func NewUpper() *Upper {
	return &Upper{}
}

func (i *Upper) Name() string {
	return "upper"
}

func (i *Upper) Description() string {
	return "Upper calls the upper API provided by the sample server"
}

func (i *Upper) Boot(flagSet *flag.FlagSet) (err error) {
	flagSet.StringVar(&i.text, "text", "test", "The text you want to convert to uppercase")

	return
}

func (i *Upper) Run(app motto.Application, args []string) (err error) {

	conn, _ := net.Dial("tcp", "127.0.0.1:8080")

	line := hotline.NewHotline(conn, time.Second*10)

	message := &pb.ReqText{
		RequestId: proto.String(fmt.Sprintf("%d", time.Now().UnixNano())),
		Text:      proto.String(i.text),
	}

	data, _ := proto.Marshal(message)

	line.Write(102, data)

	kind, resp, _ := line.Read()

	reply := &pb.RespText{}

	proto.Unmarshal(resp, reply)

	fmt.Println("reply: ", kind, reply)

	return
}
