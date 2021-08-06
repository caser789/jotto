package motto

import (
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/caser789/jotto/hotline"
	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/mux"
)

type Application interface {
	On(Event, Listener)
	Fire(Event, interface{})
	Boot() error
	Run() error
	Execute(*Processor, *Context) (uint32, proto.Message)
}

const (
	HTTP = "HTTP"
	TCP  = "TCP"
)

var (
	BootEvent  = NewEvent("motto:boot")
	PanicEvent = NewEvent("motto:panic")
)

type BaseApplication struct {
	protocol string
	address  string
	eventBus *EventBus
	routes   map[uint32]*Processor
	router   *mux.Router
}

func NewApplication(protocol string, address string, routes map[Route]*Processor) Application {
	app := &BaseApplication{
		protocol: protocol,
		address:  address,
		eventBus: NewEventBus(),
		routes:   make(map[uint32]*Processor),
		router:   mux.NewRouter(),
	}

	for route, processor := range routes {
		// Setup HTTP router
		app.router.HandleFunc(route.URI(), app.MakeHandler(processor)).Methods(route.Method())

		// Setup TCP router
		app.routes[route.ID()] = processor
	}

	return app
}

func (app *BaseApplication) On(event Event, listener Listener) {
	app.eventBus.On(event, listener)
}

func (app *BaseApplication) Fire(event Event, payload interface{}) {
	app.eventBus.Fire(event, payload)
}

type HttpHandler func(http.ResponseWriter, *http.Request)

func (app *BaseApplication) MakeHandler(processor *Processor) HttpHandler {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := &Context{
			Request:  proto.Clone(processor.Request),
			Response: proto.Clone(processor.Response),
		}

		body, _ := ioutil.ReadAll(r.Body)

		json.Unmarshal(body, &ctx.Request)

		_, reply := app.Execute(processor, ctx)

		resp, _ := json.Marshal(reply)

		w.Write(resp)
	}
}

func (app *BaseApplication) Boot() (err error) {

	http.Handle("/", app.router)

	flag.StringVar(&app.protocol, "protocol", HTTP, "HTTP or TCP")

	app.Fire(BootEvent, app)

	return
}

func (app *BaseApplication) Run() (err error) {
	switch app.protocol {
	case TCP:
		// TODO: start the TCP server
		listener, err := net.Listen("tcp", app.address)

		if err != nil {
			os.Exit(1)
		}

		defer listener.Close()

		for {
			connection, err := listener.Accept()

			if err != nil {
				// TODO: log the error
				continue
			}

			go app.tcpWorker(connection)
		}
	case HTTP:
		return http.ListenAndServe(app.address, nil)
	}

	return errors.New("Unrecognised protocol")
}

func (app *BaseApplication) Execute(processor *Processor, ctx *Context) (kind uint32, reply proto.Message) {
	// TODO: add middleware
	return processor.Handler(ctx)
}

func (app *BaseApplication) tcpWorker(connection net.Conn) {
	timeout := time.Second * 10
	line := hotline.NewHotline(connection, timeout)

	for {
		var kind uint32

		kind, input, err := line.Read()

		if err != nil {
			// TODO: handle error
			return
		}

		processor, exists := app.routes[kind]

		if !exists {
			// TODO: error response
			continue
		}

		ctx := &Context{
			Request:  proto.Clone(processor.Request),
			Response: proto.Clone(processor.Response),
		}

		proto.Unmarshal(input, ctx.Request)

		kind, reply := app.Execute(processor, ctx)

		output, err := proto.Marshal(reply)

		if err != nil {
			// TODO: error response
		}

		line.Write(kind, output)
	}
}
