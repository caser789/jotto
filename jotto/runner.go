package jotto

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/caser789/jotto/hotline"
	"github.com/gogo/protobuf/proto"
)

type Runner interface {
	Attach(app Application) error
	Run() error
}

func NewRunner(protocol string) Runner {
	switch protocol {
	case HTTP:
		runner := &HttpRunner{
			router: mux.NewRouter(),
		}
		return runner
	case TCP:
		runner := &TcpRunner{
			routes: make(map[uint32]*Processor),
		}
		return runner
	}

	return nil
}

type HttpRunner struct {
	app    Application
	router *mux.Router
}

func (r *HttpRunner) Run() error {
	http.Handle("/", r.router)
	return http.ListenAndServe(r.app.Address(), nil)
}

func (r *HttpRunner) Attach(app Application) (err error) {
	r.app = app

	for route, processor := range app.Routes() {
		// Setup HTTP router
		r.router.HandleFunc(route.URI(), r.handler(processor, app)).Methods(route.Method())
	}

	return
}

func (runner *HttpRunner) handler(processor *Processor, app Application) HttpHandler {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := &Context{
			Message:         proto.Clone(processor.Message),
			Reply:           proto.Clone(processor.Reply),
			Request:         r,
			ResponseWritter: w,
		}

		body, _ := ioutil.ReadAll(r.Body)

		json.Unmarshal(body, &ctx.Message)

		app.Execute(processor, ctx)

		resp, _ := json.Marshal(ctx.Reply)

		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
	}
}

type TcpRunner struct {
	app    Application
	routes map[uint32]*Processor
}

func (r *TcpRunner) Attach(app Application) (err error) {
	r.app = app

	for route, processor := range app.Routes() {
		// Setup TCP router
		r.routes[route.ID()] = processor
	}

	return
}

func (r *TcpRunner) Run() (err error) {
	// TODO: start the TCP server
	listener, err := net.Listen("tcp", r.app.Address())

	if err != nil {
		return
	}

	defer listener.Close()

	for {
		connection, err := listener.Accept()

		if err != nil {
			// TODO: log the error
			continue
		}

		go r.worker(connection)
	}
}

func (r *TcpRunner) worker(connection net.Conn) {
	// TODO: move into configuration
	timeout := time.Second * 10
	line := hotline.NewHotline(connection, timeout)
	defer line.Close()

	for {
		var kind uint32

		kind, input, err := line.Read()

		if err != nil {
			// TODO: handle error
			return
		}

		processor, exists := r.routes[kind]

		if !exists {
			// TODO: error response
			continue
		}

		ctx := &Context{
			MessageKind: kind,
			Message:     proto.Clone(processor.Message),
			Reply:       proto.Clone(processor.Reply),
		}

		proto.Unmarshal(input, ctx.Message)

		r.app.Execute(processor, ctx)

		output, err := proto.Marshal(ctx.Reply)

		if err != nil {
			// TODO: error response
		}

		line.Write(ctx.ReplyKind, output)
	}
}
