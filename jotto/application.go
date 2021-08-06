package jotto

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
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
	Execute(*Processor, *Context)
	Protocol() string
}

const (
	HTTP = "HTTP"
	TCP  = "TCP"
)

var (
	BootEvent  = NewEvent("jotto:boot")
	PanicEvent = NewEvent("jotto:panic")
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

func (app *BaseApplication) Protocol() string {
	return app.protocol
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

func (app *BaseApplication) Boot() (err error) {
	flag.StringVar(&app.protocol, "protocol", HTTP, "HTTP or TCP")

	app.Fire(BootEvent, app)

	flag.Parse()
	http.Handle("/", app.router)

	return
}

func (app *BaseApplication) Run() (err error) {
	fmt.Printf("Running %s server at: %s\n", app.protocol, app.address)
	switch app.protocol {
	case TCP:
		// TODO: start the TCP server
		listener, err := net.Listen("tcp", app.address)

		if err != nil {
			panic(err)
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

func (app *BaseApplication) Execute(processor *Processor, ctx *Context) {
	app.ExecuteProcessor(processor, ctx, processor.Middlewares)
}

func (app *BaseApplication) ExecuteProcessor(processor *Processor, ctx *Context, mids []Middleware) (err error) {
	if len(mids) == 0 {
		processor.Handler(app, ctx)
		return
	}

	return mids[0](app, ctx, func(c *Context) error {
		return app.ExecuteProcessor(processor, c, mids[1:])
	})
}

func (app *BaseApplication) tcpWorker(connection net.Conn) {
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

		processor, exists := app.routes[kind]

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

		app.Execute(processor, ctx)

		output, err := proto.Marshal(ctx.Reply)

		if err != nil {
			// TODO: error response
		}

		line.Write(ctx.ReplyKind, output)
	}
}
