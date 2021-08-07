package commands

import (
	"flag"
	"fmt"
	"net"
	"time"

	"git.garena.com/duanzy/motto/hotline"
	"git.garena.com/duanzy/motto/motto"
	pb "git.garena.com/duanzy/motto/sample/protocol"
	"github.com/gogo/protobuf/proto"
)

// Wait calls the wait API provided by the sample server
type Wait struct {
	motto.BaseCommand
	seconds int
}

func NewWait() *Wait {
	return &Wait{}
}

func (i *Wait) Name() string {
	return "wait"
}

func (i *Wait) Description() string {
	return "Wait calls the wait API provided by the sample server"
}

func (i *Wait) Boot(flagSet *flag.FlagSet) (err error) {
	flagSet.IntVar(&i.seconds, "text", 5, "Seconds to wait")

	return
}

func (i *Wait) Run(app motto.Application, args []string) (err error) {

	conn, _ := net.Dial("tcp", "127.0.0.1:8080")

	line := hotline.NewHotline(conn, time.Second*10)

	message := &pb.ReqWait{
		RequestId: proto.String(fmt.Sprintf("%d", time.Now().UnixNano())),
		Seconds:   proto.Int32(int32(i.seconds)),
	}

	data, _ := proto.Marshal(message)

	line.Write(104, data)

	kind, resp, _ := line.Read()

	reply := &pb.RespWait{}

	proto.Unmarshal(resp, reply)

	fmt.Println("reply: ", kind, reply)

	return
}
