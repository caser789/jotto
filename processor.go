package motto

import "github.com/gogo/protobuf/proto"

type ProcessorHandler func(*Context) proto.Message

type Processor struct {
	Request  proto.Message
	Response proto.Message
	Handler  func(*Context) proto.Message
}
