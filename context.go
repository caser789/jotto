package motto

import (
	"net/http"

	"github.com/gogo/protobuf/proto"
)

type Context struct {
	// Proto messages
	MessageKind uint32
	ReplyKind   uint32
	Message     proto.Message
	Reply       proto.Message

	// HTTP messages
	Request         *http.Request
	ResponseWritter http.ResponseWriter
}
