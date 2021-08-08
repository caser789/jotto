package jotto

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"

	"git.garena.com/duanzy/motto/hotline"
	"github.com/golang/protobuf/proto"
)

// Runner defines the logic of how an application should be run.
type Runner interface {
	Attach(app Application) error
	Run() error
	Shutdown(timeout time.Duration) error
}

// NewRunner creates a Runner according to the given `protocol`.
func NewRunner(protocol string) (runner Runner) {
	switch protocol {
	case HTTP:
		runner = &HttpRunner{
			router: mux.NewRouter(),
		}
	case TCP:
		runner = &TcpRunner{
			routes: make(map[uint32]Processor),
			alive:  true,
			wg:     &sync.WaitGroup{},
		}
	case SPEX:
		runner = &SpexRunner{}
	}

	return
}

// HttpHandler is a HTTP handler used in the net.http package.
type HttpHandler func(http.ResponseWriter, *http.Request)

// HttpRunner is the built-in HTTP runner of Motto
type HttpRunner struct {
	app    Application
	router *mux.Router
	server *http.Server
}

// Run runs the application in HTTP mode
func (r *HttpRunner) Run() (err error) {
	fmt.Printf("Running %s server at %s\n", r.app.Protocol(), r.app.Address())

	writeTimeout := r.app.Settings().Motto().WriteTimeout
	if writeTimeout == 0 {
		writeTimeout = 10
	}

	readTimeout := r.app.Settings().Motto().ReadTimeout
	if readTimeout == 0 {
		readTimeout = 10
	}

	idleTimeout := r.app.Settings().Motto().IdleTimeout
	if idleTimeout == 0 {
		idleTimeout = 30
	}

	r.server = &http.Server{
		Addr: r.app.Address(),
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * time.Duration(writeTimeout),
		ReadTimeout:  time.Second * time.Duration(readTimeout),
		IdleTimeout:  time.Second * time.Duration(idleTimeout),
		Handler:      r.router, // Pass our instance of gorilla/mux in.
	}

	listener, err := r.app.GetListener()

	if err != nil {
		return err
	}

	return r.server.Serve(listener)
}

// Shutdown shuts down the HTTP server
func (r *HttpRunner) Shutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Wait for existing requests to finish (with a 5-second timeout)
	r.server.Shutdown(ctx)
	return fmt.Errorf("Shutting down")
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

		var (
			err     error
			body    []byte
			ctx     = context.Background()
			message = proto.Clone(processor.Message())
			reply   = proto.Clone(processor.Reply())
		)

		ctx = context.WithValue(ctx, CtxHTTPRequest, request)
		ctx = context.WithValue(ctx, CtxHTTPResponse, writer)
		ctx = context.WithValue(ctx, CtxLogger, logger)
		ctx = context.WithValue(ctx, CtxTime, uint32(time.Now().Unix()))
		ctx = r.app.MakeContext(ctx, processor)

		defer func() {
			if er := recover(); er != nil {
				logger.Errorf("motto|http_runner|recover_from_panic|panic=%v,stack=%s", er, debug.Stack())
				app.Panic(ctx, er, message, reply)
				r.respond(ctx, writer, message, reply)
			}
		}()

		if body, err = ioutil.ReadAll(request.Body); err != nil {
			logger.Errorf("motto|http_runner|failed_to_read_request_body|err=%v", err)
			panic(err)
		}

		ctx = context.WithValue(ctx, CtxHTTPRequestBody, body)

		if len(body) > 0 {
			if err = json.Unmarshal(body, &message); err != nil {
				logger.Errorf("motto|http_runner|failed_to_unmarshal_incoming_message|body=%s,err=%v", body, err)
				panic(err)
			}
		}

		_, ctx = app.Execute(ctx, processor, message, reply)

		r.respond(ctx, writer, message, reply)
	}
}

func (r *HttpRunner) respond(ctx context.Context, writer http.ResponseWriter, req, reply interface{}) {
	var (
		err    error
		resp   []byte
		logger = GetLogger(ctx)
	)
	if v, ok := ctx.Value(CtxHTTPResponseBody).([]byte); ok {
		// Response body generated, directly use it
		resp = v
	} else {
		// Response body not generated, marshal from proto message
		if resp, err = json.Marshal(reply); err != nil {
			logger.Errorf("motto|http_runner|failed_to_marshal_outgoing_message|reply=%v,err=%v", reply, err)
			panic(err)
		}
	}

	writer.Header().Set("Content-Type", "application/json")

	// Attach headers emitted by application
	if headers, ok := ctx.Value(CtxHTTPResponseHeaders).(map[string]string); ok {
		for k, v := range headers {
			writer.Header().Set(k, v)
		}
	}

	if status, ok := ctx.Value(CtxHTTPStatus).(int); ok {
		writer.WriteHeader(status)
	}

	writer.Write(resp)
}

// TcpRunner is the built-in TCP runner of Motto
type TcpRunner struct {
	app    Application
	routes map[uint32]Processor
	alive  bool
	wg     *sync.WaitGroup
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
	logger.Debugf("Running %s server at %s", r.app.Protocol(), r.app.Address())

	listener, err := r.app.GetListener()

	if err != nil {
		logger.Fatalf("Failed to listen on (%s//%s). (error=%v)", r.app.Protocol(), r.app.Address(), err)
		return
	}

	defer listener.Close()

	for r.alive {
		connection, err := listener.Accept()

		if err != nil {
			logger.Errorf("Failed to accept incomming connection. (error=%v)", err)
			continue
		}

		r.wg.Add(1)
		go r.worker(connection)
	}

	return
}

// Shutdown shuts down the TCP server
func (r *TcpRunner) Shutdown(timeout time.Duration) error {
	r.alive = false // Signal listener to stop accepting new connections; workers to exit.

	c := make(chan struct{})
	go func() {
		defer close(c)
		r.wg.Wait()
	}()

	// Wait for existing requests to finish for up to `timeout`.
	select {
	case <-c:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("Shutdown wait timeout")
	}
}

func (r *TcpRunner) worker(connection net.Conn) {
	defer r.wg.Done()
	defer connection.Close()

	// TODO: move into configuration
	timeout := time.Second * 10
	line := hotline.NewHotline(connection, timeout)
	defer line.Close()

	for r.alive {
		var kind uint32

		logger := r.app.MakeLogger(map[string]interface{}{
			"trace_id": GenerateTraceID(),
		})

		logger.Tracef("Logger created")

		kind, input, err := line.Read()

		if err != nil {
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				logger.Errorf("Hotline %s timed out, error: %v", line, err)
			} else if err == io.EOF {
				// Ignore
			} else {
				logger.Errorf("Failed to read from hotline %s, error: %v", line, err)
			}
			return
		}

		processor, exists := r.routes[kind]

		ctx := context.Background()
		ctx = context.WithValue(ctx, CtxLogger, logger)

		if !exists {
			// The given message identifier (kind) does not exist in the routing
			// table. We will fire an event to let the application handle this
			// case. The application is supposed to initialize the ctx.Reply field
			// with a proper proto.Message and fill in the ctx.ReplyKind.
			r.app.Fire(RouteNotFoundEvent, ctx)
			continue
		}

		ctx = r.app.MakeContext(ctx, processor)

		message := proto.Clone(processor.Message())
		reply := proto.Clone(processor.Reply())

		proto.Unmarshal(input, message)

		code, ctx := r.app.Execute(ctx, processor, message, reply)

		output, err := proto.Marshal(reply)

		if err != nil {
			// In case of a marshal error, we will panic and let the application
			// deal with the aftermath.
			r.app.Fire(PanicEvent, ctx)
		}

		err = line.Write(uint32(code), output)

		if err != nil {
			logger.Errorf("Failed to write to hotline %s, error: %v", line, err)
		}
	}
}

// CliRunner is the built-in runner for running the application on command line
type CliRunner struct {
	app  Application
	bus  *CommandBus
	args []string
}

func NewCliRunner(bus *CommandBus, args []string) (runner *CliRunner) {
	return &CliRunner{
		bus:  bus,
		args: args,
	}
}

func (r *CliRunner) Attach(app Application) (err error) {
	r.app = app
	return
}

func (r *CliRunner) Run() (err error) {
	if len(r.args) < 1 {
		fmt.Println("Not enough arguments")
		r.help()
		return
	}

	name := r.args[0]

	// 2. Find command in the bus
	command, err := r.bus.Find(name)

	if err != nil {
		fmt.Printf("Cannot find command: %s (%v)\n", name, err)
		r.help()
		return
	}

	// flag.Bool(command.Name(), true, "command name")
	flagSet := flag.NewFlagSet(command.Name(), flag.ExitOnError)

	// 3. Run command initializations.
	command.Boot(flagSet)

	flagSet.Parse(r.args[1:])

	// 4. Run the command.
	return command.Run(r.app, flag.Args())
}

func (r *CliRunner) Shutdown(timeout time.Duration) error {
	return nil
}

func (r *CliRunner) help() {
	fmt.Printf("Usage: %s -<command-name> ...<flags> ...<args>    To run a command\n", os.Args[0])
	fmt.Printf("       %s -<command-name> -h                      To get usage information of a specific command\n\n", os.Args[0])
	r.bus.Print()
}

func NewQueueWorkerRunner(queue string, size int) *QueueWorkerRunner {
	workers := make(chan bool, size)

	for i := 0; i < size; i++ {
		workers <- true
	}

	return &QueueWorkerRunner{
		queue:   queue,
		alive:   true,
		workers: workers,
	}
}

type QueueWorkerRunner struct {
	app     Application
	queue   string
	alive   bool
	workers chan bool
}

func (r *QueueWorkerRunner) Attach(app Application) error {
	r.app = app

	return nil
}

func (r *QueueWorkerRunner) Run() error {
	defer func() {
		if er := recover(); er != nil {
			fmt.Printf("motto|queue_runner|recover_from_panic|panic=%v,stack=%s\n", er, debug.Stack())
		}
	}()

	Q := r.app.Queue(r.queue)

	go r.watcher()

	for r.alive {
		logger := r.app.MakeLogger(map[string]interface{}{
			"trace_id": GenerateTraceID(),
		})

		ok := r.acquire(time.Second * 2)
		if !ok {
			logger.Tracef("Failed to get hold of an available worker. Giving up.")
			continue
		}

		job, err := Q.Dequeue()

		if err != nil {
			if err != redis.Nil {
				logger.Errorf("Pop job error: %v", err)
			}
			r.release()
			continue
		}

		logger.Dataf("Received job: %+v", job)

		jobs := r.app.Jobs()

		processor, ok := jobs[job.Type]

		if !ok {
			logger.Errorf("Job processor not found for job %s", job.TraceID)
			Q.Fail(job)
			r.release()
			continue
		}

		go r.process(processor, job, r.app, logger, Q)
	}

	return nil
}

// Acquire a worker from pool
func (r *QueueWorkerRunner) acquire(timeout time.Duration) bool {
	select {
	case <-r.workers:
		return true
	case <-time.After(timeout):
		return false
	}
}

// Release a worker to pool
func (r *QueueWorkerRunner) release() bool {
	select {
	case r.workers <- true:
		return true
	default:
		return false
	}
}

func (r *QueueWorkerRunner) process(processor QueueProcessor, job *Job, app Application, logger Logger, Q *Queue) (err error) {
	defer func() {
		er := Q.Attempt(job)

		if er != nil {
			logger.Errorf("QueueWorkerRunner|process|attemp_job_failure:%v", er)
		}

		if ex := recover(); ex == nil {
			/*
			 * Job processor returned normally. Check its err and determine what to do.
			 */
			var action string
			var perr error

			switch err {
			case ErrorJobHandled: // ignore handled job
				action = "ignore"
			case ErrorJobMustRetry: // we must retry this job; use exponential backoff to attempt it later
				perr = Q.Defer(job, r.backoff(job.Attempts))
				action = "defer"
			case nil: // auto complete
				perr = Q.Complete(job)
				action = "complete"
			default: // auto retry on error
				if job.Attempts <= 10 { // Backoff exponentially for 10 times
					perr = Q.Defer(job, r.backoff(job.Attempts))
					action = "defer"
				} else { // Failed 10 times in a row, giving up
					perr = Q.Fail(job)
					action = "fail"
				}
			}
			callbackProcessor, er := r.app.GetJobCallBackFunc(job.Type)
			if er == nil {
				callbackProcessor(Q, job, app, action, err)
			}
			logger.Dataf("QueueWorkerRunner|process|action=%s,err=%v,job_id=%s", action, perr, job.TraceID)
		} else {
			/*
			 * Job processor crashed, enter exception handling logic.
			 */
			logger.Errorf("QueueWorkerRunner|process|job_crashed|error=%v,job_id=%s", ex, job.TraceID)
			logger.Errorf("QueueWorkerRunner|process|panic=%v,stack=%s", ex, debug.Stack())

			action := ""
			if job.Attempts <= 10 { // Backoff exponentially for 10 times
				err = Q.Defer(job, r.backoff(job.Attempts))
				action = "defer"
			} else { // Failed 10 times in a row, giving up
				err = Q.Fail(job)
				action = "fail"
			}
			callbackProcessor, er := r.app.GetJobCallBackFunc(job.Type)
			if er == nil {
				callbackProcessor(Q, job, app, action, err)
			}
			logger.Dataf("QueueWorkerRunner|process|action=%s,err=%v,job_id=%s", action, err, job.TraceID)
		}

		ok := r.release()
		if !ok {
			logger.Errorf("QueueWorkerRunner|process|worker_pool_full|cannot_release_worker|terminate_without_replenishing_the_pool")
		}

	}()

	// Execute the job processor
	err = processor(Q, job, app, logger)

	return
}

func (r *QueueWorkerRunner) backoff(attempt int64) time.Duration {
	return time.Second * time.Duration(math.Pow(2, float64(attempt)))
}

func (r *QueueWorkerRunner) watcher() {
	Q := r.app.Queue(r.queue)
	logger := r.app.MakeLogger(nil)

	for r.alive {
		stats, err := Q.Stats()

		logger.Dataf("%+v", stats)

		if err == nil {
			if stats.Waiting > 0 {
				scheduled, err := Q.driver.ScheduleDeferred(Q.name)

				if err == nil {
					logger.Dataf("Queue: scheduled %d jobs.", scheduled)
				} else {
					logger.Errorf("Queue: failed to schedule deferred jobs. (err=%v)", err)
				}
			}
		} else {
			logger.Errorf("Queue: failed to retrieve queue stats. (err=%v)", err)
		}

		time.Sleep(time.Duration(1) * time.Second)
	}
}

// Shutdown - shutdown the runner
func (r *QueueWorkerRunner) Shutdown(timeout time.Duration) error {
	r.alive = false
	return nil
}

// SpexRunner - run the application in Spex
type SpexRunner struct {
	app Application
}

// Attach - attach application to runner
func (r *SpexRunner) Attach(app Application) (err error) {
	r.app = app
	return
}

// Run - run the application
func (r *SpexRunner) Run() (err error) {
	return
}

// Shutdown - shutdown the runner
func (r *SpexRunner) Shutdown(timeout time.Duration) error {
	return nil
}
