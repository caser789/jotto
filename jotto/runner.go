package jotto

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"

	"github.com/caser789/jotto/hotline"
	"github.com/gogo/protobuf/proto"
)

// Runner defines the logic of how an application should be run.
type Runner interface {
	Attach(app Application) error
	Run() error
}

// NewRunner creates a Runner according to the given `protocol`.
func NewRunner(protocol string) (runner Runner) {
	switch protocol {
	case HTTP:
		runner = &HttpRunner{
			router: mux.NewRouter(),
		}
		return runner
	case TCP:
		runner = &TcpRunner{
			routes: make(map[uint32]Processor),
		}
		return runner
	}

	return nil
}

// HttpHandler is a HTTP handler used in the net.http package.
type HttpHandler func(http.ResponseWriter, *http.Request)

// HttpRunner is the built-in HTTP runner of Motto
type HttpRunner struct {
	app    Application
	router *mux.Router
}

// Run runs the application in HTTP mode
func (r *HttpRunner) Run() error {
	fmt.Printf("Running %s server at %s\n", r.app.Protocol(), r.app.Address())
	http.Handle("/", r.router)
	return http.ListenAndServe(r.app.Address(), nil)
}

// Attach binds the appliation to the runner and initializes the HTTP router.
func (r *HttpRunner) Attach(app Application) (err error) {
	r.app = app

	for route, processor := range app.Routes() {
		// Setup HTTP router
		r.router.HandleFunc(route.URI(), r.handler(processor, app)).Methods(route.Method())
	}

	return
}

func (r *HttpRunner) handler(processor Processor, app Application) HttpHandler {

	return func(writer http.ResponseWriter, request *http.Request) {
		logger := app.MakeLogger(map[string]interface{}{
			"trace_id": GenerateTraceID(),
		})

		ctx := &BaseContext{
			Message:         proto.Clone(processor.Message()),
			Reply:           proto.Clone(processor.Reply()),
			Request:         request,
			ResponseWritter: writer,
			Logger:          logger,
		}

		context := app.MakeContext(processor, ctx)

		body, err := ioutil.ReadAll(request.Body)

		if err != nil {
			logger.Error("Failed to read request body")
			app.Fire(PanicEvent, ctx)
		}

		err = json.Unmarshal(body, &ctx.Message)

		if err != nil {
			logger.Error("Failed to unmarshal incoming message. (body=%s)", body)
			app.Fire(PanicEvent, ctx)
		}

		app.Execute(processor, context)

		resp, err := json.Marshal(ctx.Reply)

		if err != nil {
			logger.Error("Failed to marshal outgoing message. (reply=%v)", ctx.Reply)
			app.Fire(PanicEvent, ctx)
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(resp)
	}
}

// TcpRunner is the built-in TCP runner of Motto
type TcpRunner struct {
	app    Application
	routes map[uint32]Processor
}

// Attach binds the application to the runner and initializes the TCP router
func (r *TcpRunner) Attach(app Application) (err error) {
	r.app = app

	for route, processor := range app.Routes() {
		// Setup TCP router
		r.routes[route.ID()] = processor
	}

	return
}

// Run starts the TCP server and serves incoming requests
func (r *TcpRunner) Run() (err error) {
	logger := r.app.MakeLogger(nil)
	logger.Debug("Running %s server at %s", r.app.Protocol(), r.app.Address())

	listener, err := net.Listen("tcp", r.app.Address())

	if err != nil {
		logger.Fatal("Failed to listen on (%s//%s). (error=%v)", r.app.Protocol(), r.app.Address(), err)
		return
	}

	defer listener.Close()

	for {
		connection, err := listener.Accept()

		if err != nil {
			logger.Error("Failed to accept incomming connection. (error=%v)", err)
			continue
		}

		go r.worker(connection)
	}
}

func (r *TcpRunner) worker(connection net.Conn) {
	defer connection.Close()

	// TODO: move into configuration
	timeout := time.Second * 10
	line := hotline.NewHotline(connection, timeout)
	defer line.Close()

	for {
		var kind uint32

		logger := r.app.MakeLogger(map[string]interface{}{
			"trace_id": GenerateTraceID(),
		})

		logger.Trace("Logger created")

		kind, input, err := line.Read()

		if err != nil {
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				logger.Error("Hotline %s timed out, error: %v", line, err)
			} else if err == io.EOF {
				// Ignore
			} else {
				logger.Error("Failed to read from hotline %s, error: %v", line, err)
			}
			return
		}

		processor, exists := r.routes[kind]

		ctx := &BaseContext{
			MessageKind: kind,
			Logger:      logger,
		}

		if !exists {
			// The given message identifier (kind) does not exist in the routing
			// table. We will fire an event to let the application handle this
			// case. The application is supposed to initialize the ctx.Reply field
			// with a proper proto.Message and fill in the ctx.ReplyKind.
			r.app.Fire(RouteNotFoundEvent, ctx)
		} else {
			ctx.Message = proto.Clone(processor.Message())
			ctx.Reply = proto.Clone(processor.Reply())

			context := r.app.MakeContext(processor, ctx)

			proto.Unmarshal(input, ctx.Message)

			r.app.Execute(processor, context)
		}

		output, err := proto.Marshal(ctx.Reply)

		if err != nil {
			// In case of a marshal error, we will panic and let the application
			// deal with the aftermath.
			r.app.Fire(PanicEvent, ctx)
		}

		err = line.Write(ctx.ReplyKind, output)

		if err != nil {
			logger.Error("Failed to write to hotline %s, error: %v", line, err)
		}
	}
}

// CliRunner is the built-in runner for running the application on command line
type CliRunner struct {
	app Application
	bus *CommandBus
}

func NewCliRunner(bus *CommandBus) (runner *CliRunner) {
	return &CliRunner{
		bus: bus,
	}
}

func (r *CliRunner) Attach(app Application) (err error) {
	r.app = app
	return
}

func (r *CliRunner) Run() (err error) {
	if len(os.Args) < 2 {
		r.help()
		return
	}

	name := os.Args[1]

	// 2. Find command in the bus
	command, err := r.bus.Find(name[1:])

	if err != nil {
		r.help()
		return
	}

	flag.Bool(command.Name(), true, "command name")

	// 3. Run command initializations.
	command.Boot()

	flag.Parse()

	r.app.Boot()

	// 4. Run the command.
	command.Run(r.app, flag.Args())
	return
}

func (r *CliRunner) help() {
	fmt.Printf("Usage: %s -<command-name> ...<flags> ...<args>    To run a command\n", os.Args[0])
	fmt.Printf("       %s -<command-name> -h                      To get usage information of a specific command\n\n", os.Args[0])
	r.bus.Print()
}
