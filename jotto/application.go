package jotto

import (
	"context"
	"fmt"
	"net"
	"time"
)

// Application is an abstraction of a runnable application in Motto.
type Application interface {
	On(Event, Listener)
	Fire(Event, ...interface{})
	Boot() error
	Run() error
	Reload() error
	Shutdown(timeout time.Duration) error
	Execute(ctx context.Context, processor Processor, request, response interface{}) (int32, context.Context)

	Protocol() string
	Address() string
	Routes() map[Route]Processor
	Jobs() map[int]QueueProcessor

	Get(string) (interface{}, bool)
	Set(string, interface{})
	Settings() Configuration

	SetContextFactory(ContextFactory)
	MakeContext(context.Context, Processor) context.Context

	SetLoggerFactory(LoggerFactory)
	MakeLogger(LoggerContext) Logger

	Cache(name string) CacheDriver
	Queue(name string) *Queue

	GetListener() (net.Listener, error)
	SetListener(net.Listener)
}

const (
	// HTTP - the HTTP protocol
	HTTP = "HTTP"

	// TCP - the TCP protocol
	TCP = "TCP"

	// SPEX - the SPEX protocol
	SPEX = "SPEX"
)

var (
	// BootEvent is fired when the application boots (i.e. When the Boot method is called)
	BootEvent = NewEvent("motto:boot")

	// ReloadEvent is fired when the application configuration is reloaded
	ReloadEvent = NewEvent("motto:reload")

	// TerminateEvent is fired when the application will start to terminate
	TerminateEvent = NewEvent("motto:terminate")

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

	cache map[string]CacheDriver
	queue map[string]*Queue
	jobs  map[int]QueueProcessor

	listener net.Listener
	runner   Runner
}

// NewApplication creates a new application.
func NewApplication(settings Configuration, routes map[Route]Processor, jobs map[int]QueueProcessor, runner Runner) Application {
	if settings == nil {
		settings = NewDefaultSettings()
	}
	app := &BaseApplication{
		eventBus:       NewEventBus(),
		routes:         routes,
		registry:       make(map[string]interface{}),
		settings:       settings,
		contextFactory: func(c context.Context, p Processor) context.Context { return c },
		loggerFactory:  func(a Application, c LoggerContext) Logger { return NewStdoutLogger(c) },
		cache:          make(map[string]CacheDriver),
		queue:          make(map[string]*Queue),
		jobs:           jobs,
	}

	if runner != nil {
		app.runner = runner
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

// Jobs returns the queue job settings of the application
func (app *BaseApplication) Jobs() map[int]QueueProcessor {
	return app.jobs
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
func (app *BaseApplication) MakeContext(ctx context.Context, processor Processor) context.Context {
	return app.contextFactory(ctx, processor)
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
	err = app.settings.Load()

	if err != nil {
		return
	}

	app.protocol = app.settings.Motto().Protocol
	app.address = app.settings.Motto().Address

	app.initializeServices()

	// Fire boot event
	app.Fire(BootEvent, app)

	return
}

// Run starts running the appliation and serving incoming requests
func (app *BaseApplication) Run() (err error) {
	if app.runner == nil {
		app.runner = NewRunner(app.protocol)
	}
	if app.runner == nil {
		return fmt.Errorf("Unrecognised protocol: %s", app.protocol)
	}

	app.runner.Attach(app)

	return app.runner.Run()
}

// Shutdown shuts down the application
func (app *BaseApplication) Shutdown(timeout time.Duration) (err error) {
	app.Fire(TerminateEvent, app)
	return app.runner.Shutdown(timeout)
}

// Execute executes a processor
func (app *BaseApplication) Execute(ctx context.Context, processor Processor, request, response interface{}) (int32, context.Context) {
	return app.ExecuteProcessor(ctx, processor, processor.Middlewares(), request, response)
}

// ExecuteProcessor executes a processor
func (app *BaseApplication) ExecuteProcessor(ctx context.Context, processor Processor, mids []Middleware, request, response interface{}) (int32, context.Context) {
	if len(mids) == 0 {
		return processor.Handler()(ctx, app, request, response)
	}

	return mids[0](ctx, app, request, response, func(c context.Context) (int32, context.Context) {
		return app.ExecuteProcessor(c, processor, mids[1:], request, response)
	})
}

func (app *BaseApplication) Reload() (err error) {
	err = app.settings.Load()
	if err != nil {
		return
	}

	app.initializeServices()
	app.Fire(ReloadEvent, app)

	return
}

func (app *BaseApplication) Cache(name string) CacheDriver {
	if c, ok := app.cache[name]; ok {
		return c
	}
	return NewNullDriver(name)
}

func (app *BaseApplication) Queue(name string) *Queue {
	if q, ok := app.queue[name]; ok {
		return q
	}
	return nil
}

func (app *BaseApplication) GetListener() (listener net.Listener, err error) {
	if app.listener != nil {
		return app.listener, nil
	} else {
		app.listener, err = net.Listen("tcp", app.address)
	}

	return app.listener, err
}

func (app *BaseApplication) SetListener(listener net.Listener) {
	app.listener = listener
}

// Initialize external services such as cache, queue
func (app *BaseApplication) initializeServices() {
	for _, c := range app.settings.Motto().Cache {
		switch c.Driver {
		case "redis":
			app.cache[c.Name] = NewRedisDriver(c.Name, c.Redis)
		case "memcached":
		default:
			// pass
		}
	}

	for _, q := range app.settings.Motto().Queue {
		switch q.Driver {
		case "redis":
			driver := NewRedisDriver(q.Name, q.Redis)
			for _, name := range q.Queues {
				key := q.Name + ":" + name
				app.queue[key] = NewQueue(name, driver)
			}
		default:
			// pass
		}
	}
}
