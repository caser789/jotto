package jotto

import (
	"fmt"
)

type Application interface {
	On(Event, Listener)
	Fire(Event, interface{})
	Boot() error
	Run(Runner) error
	Execute(*Processor, *Context)

	Protocol() string
	Address() string
	Routes() map[Route]*Processor

	Get(string) (interface{}, bool)
	Set(string, interface{})
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
	registry map[string]interface{}
}

func NewApplication(protocol string, address string, routes map[Route]*Processor) Application {
	app := &BaseApplication{
		protocol: protocol,
		address:  address,
		eventBus: NewEventBus(),
		routes:   routes,
		registry: make(map[string]interface{}),
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

func (app *BaseApplication) Boot() (err error) {
	app.Fire(BootEvent, app)

	return
}

func (app *BaseApplication) Run(runner Runner) (err error) {
	if runner == nil {
		runner := NewRunner(app.protocol)
	}
	if runner == nil {
		return fmt.Errorf("Unrecognised protocol: %s", app.protocol)
	}

	runner.Attach(app)
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

func (app *BaseApplication) Get(key string) (value interface{}, ok bool) {
	value, ok = app.registry[key]
	return
}

func (app *BaseApplication) Set(key string, value interface{}) {
	app.registry[key] = value
}
