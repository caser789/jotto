package motto

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type Application interface {
	On(Event, Listener)
	Fire(Event, interface{})
	Boot() error
	Run() error
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
	router   *mux.Router
}

func NewApplication(protocol string, address string, routes map[Route]*Processor) (app Application) {
	app = &BaseApplication{
		protocol: protocol,
		address:  address,
		eventBus: NewEventBus(),
		routes:   routes,
		router:   mux.NewRouter(),
	}

	return
}

// Register the event to listenr
func (app *BaseApplication) On(event Event, listener Listener) {
	app.eventBus.On(event, listener)
}

func (app *BaseApplication) Fire(event Event, payload interface{}) {
	app.eventBus.Fire(event, payload)
}

type HttpHandler func(http.ResponseWriter, *http.Request)

func (app *BaseApplication) MakeHandler(processor *Processor) HttpHandler {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func (app *BaseApplication) Boot() (err error) {
	app.Fire(BootEvent, nil)

	for route, processor := range app.routes {
		app.router.HandleFunc(route.URI(), app.MakeHandler(processor).Methods(route.Method()))
		fmt.Println(route)
	}

	return
}

func (app *BaseApplication) Run() (err error) {
	switch app.protocol {
	case TCP:
		return
	case HTTP:
		http.Handle("/", app.router)
		return
	}

	return
}
