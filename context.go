package motto

import "github.com/gogo/protobuf/proto"

type Context struct {
	Request  proto.Message
	Response proto.Message
}
