package main

import (
	"fmt"
	"net"
	"time"

	pb "github.com/caser789/jotto/example/protocol"
	"github.com/caser789/jotto/hotline"
	"github.com/gogo/protobuf/proto"
)

func main() {
	conn, _ := net.Dial("tcp", "127.0.0.1:8080")

	line := hotline.NewHotline(conn, time.Second*10)

	message := &pb.ReqAbout{
		RequestId: proto.String(fmt.Sprintf("%v", time.Now())),
	}

	data, _ := proto.Marshal(message)

	line.Write(100, data)

	kind, resp, _ := line.Read()

	reply := &pb.RespAbout{}

	proto.Unmarshal(resp, reply)

	fmt.Println("reply: ", kind, reply)
}
