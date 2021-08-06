package motto

import "github.com/gogo/protobuf/proto"

type ProcessorHandler func(*Context)

type Processor struct {
	Message     proto.Message
	Reply       proto.Message
	Handler     ProcessorHandler
	Middlewares []Middleware
}
