package jotto

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// Application is an abstraction of a runnable application in Motto.
type Application interface {
	On(Event, Listener)
	Fire(Event, ...interface{})
	Boot() error
	Run(Runner) error
	Execute(Processor, Context)

	Protocol() string
	Address() string
	Routes() map[Route]Processor

	Get(string) (interface{}, bool)
	Set(string, interface{})
	Settings() Configuration

	SetContextFactory(ContextFactory)
	MakeContext(Processor, *BaseContext) Context

	SetLoggerFactory(LoggerFactory)
	MakeLogger(LoggerContext) Logger
}

const (
	// HTTP the protocol
	HTTP = "HTTP"

	// TCP the protocol
	TCP = "TCP"
)

var (
	// BootEvent is fired when the application boots (i.e. When the Boot method is called)
	BootEvent = NewEvent("motto:boot")

	// ReloadEvent is fired when the application configuration is reloaded
	ReloadEvent = NewEvent("motto:reload")

	// PanicEvent is fired when a non-recoverable error happened
	PanicEvent = NewEvent("motto:panic")

	// RouteNotFoundEvent is fired by the TCP runner when a unknown command ID is received
	RouteNotFoundEvent = NewEvent("motto:routing:notfound")
)

// BaseApplication is the default implementation of `Application` in Motto
type BaseApplication struct {
	protocol       string
	address        string
	eventBus       *EventBus
	routes         map[Route]Processor
	registry       map[string]interface{}
	settings       Configuration
	contextFactory ContextFactory
	loggerFactory  LoggerFactory
}

// NewApplication creates a new application.
func NewApplication(settings Configuration, routes map[Route]Processor) Application {
	app := &BaseApplication{
		eventBus:       NewEventBus(),
		routes:         routes,
		registry:       make(map[string]interface{}),
		settings:       settings,
		contextFactory: func(p Processor, c *BaseContext) Context { return c },
		loggerFactory:  func(a Application, c LoggerContext) Logger { return NewStdoutLogger(c) },
	}

	return app
}

// Protocol returns the protocol of the application
func (app *BaseApplication) Protocol() string {
	return app.protocol
}

// Address returns the address of the application
func (app *BaseApplication) Address() string {
	return app.address
}

// Routes returns the route settings of the application
func (app *BaseApplication) Routes() map[Route]Processor {
	return app.routes
}

// On registers an event listener
func (app *BaseApplication) On(event Event, listener Listener) {
	app.eventBus.On(event, listener)
}

// Fire fires an event with payload
func (app *BaseApplication) Fire(event Event, payload ...interface{}) {
	app.eventBus.Fire(event, payload...)
}

// Get retrieves an entry from the application's registry
func (app *BaseApplication) Get(key string) (value interface{}, ok bool) {
	value, ok = app.registry[key]
	return
}

// Set puts `value` into the application's registry under `key`
func (app *BaseApplication) Set(key string, value interface{}) {
	app.registry[key] = value
}

// Settings returns the settings of the application
func (app *BaseApplication) Settings() Configuration {
	return app.settings
}

// SetContextFactory sets a custom context factory function
func (app *BaseApplication) SetContextFactory(factory ContextFactory) {
	app.contextFactory = factory
}

// MakeContext creates an execution context using the context factory
func (app *BaseApplication) MakeContext(processor Processor, ctx *BaseContext) Context {
	return app.contextFactory(processor, ctx)
}

// SetLoggerFactory sets a custom logger factory function
func (app *BaseApplication) SetLoggerFactory(factory LoggerFactory) {
	app.loggerFactory = factory
}

// MakeLogger creates a logger using the logger factory
func (app *BaseApplication) MakeLogger(c LoggerContext) Logger {
	return app.loggerFactory(app, c)
}

// Boot initializes the application
func (app *BaseApplication) Boot() (err error) {
	// Load configuration
	app.settings.Load()

	app.protocol = app.settings.Motto().Protocol
	app.address = app.settings.Motto().Address

	app.On(ReloadEvent, app.Reload)

	// Fire boot event
	app.Fire(BootEvent, app)

	// Listen for reload signal
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGUSR2)

	go func() {
		for {
			<-s
			app.settings.Load()
			app.Fire(ReloadEvent, app)
		}
	}()

	return
}

// Run starts running the appliation and serving incoming requests
func (app *BaseApplication) Run(runner Runner) (err error) {
	if runner == nil {
		runner = NewRunner(app.protocol)
	}

	if runner == nil {
		return fmt.Errorf("Unrecognised protocol: %s", app.protocol)
	}

	runner.Attach(app)
	return runner.Run()
}

// Execute executes a processor
func (app *BaseApplication) Execute(processor Processor, ctx Context) {
	app.ExecuteProcessor(processor, ctx, processor.Middlewares())
}

// ExecuteProcessor executes a processor
func (app *BaseApplication) ExecuteProcessor(processor Processor, ctx Context, mids []Middleware) (err error) {
	if len(mids) == 0 {
		processor.Handler()(app, ctx)
		return
	}

	return mids[0](app, ctx, func(c Context) error {
		return app.ExecuteProcessor(processor, c, mids[1:])
	})
}

func (app *BaseApplication) Reload(payloads ...interface{}) {
}
