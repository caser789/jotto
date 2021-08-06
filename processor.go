package motto

import "github.com/gogo/protobuf/proto"

type ProcessorHandler func(*Context) (uint32, proto.Message)

type Processor struct {
	Request  proto.Message
	Response proto.Message
	Handler  ProcessorHandler
}
