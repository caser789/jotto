package motto

import "github.com/gogo/protobuf/proto"

type Context struct {
	request  proto.Message
	response protoMessage
}
