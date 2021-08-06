package jotto

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gogo/protobuf/proto"
)

type Application interface {
	On(Event, Listener)
	Fire(Event, interface{})
	Boot() error
	Run() error
	Execute(*Processor, *Context)

	Protocol() string
	Address() string
	Routes() map[Route]*Processor
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
	routes   map[Route]*Processor
}

func NewApplication(protocol string, address string, routes map[Route]*Processor) Application {
	app := &BaseApplication{
		protocol: protocol,
		address:  address,
		eventBus: NewEventBus(),
		routes:   routes,
	}

	return app
}

func (app *BaseApplication) Protocol() string {
	return app.protocol
}

func (app *BaseApplication) Address() string {
	return app.address
}

func (app *BaseApplication) Routes() map[Route]*Processor {
	return app.routes
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

	return
}

func (app *BaseApplication) Run() (err error) {
	fmt.Printf("Running %s server at: %s\n", app.protocol, app.address)

	runner := NewRunner(app.protocol)
	runner.Attach(app)

	if runner == nil {
		return errors.New("Unrecognised protocol")
	}

	return runner.Run()
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
