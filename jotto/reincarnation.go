package motto

import (
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Reincarnation is an interface for zero downtime respawning.
type Reincarnation interface {
	Serve()
	Reload()
	Reincarnate()
}

// NewSoul creates a new soul that can serve multiple applications.
func NewSoul(apps []Application) *Soul {
	return &Soul{
		apps: apps,
	}
}

// Soul represents a soul (which contains multiple applications) that can reincarnate itself.
// It implements the `Reincarnation` interface.
type Soul struct {
	apps      []Application
	listeners []net.Listener
}

// Serve starts serving the applications
func (s *Soul) Serve() error {
	for _, app := range s.apps {
		app.Boot()

		/*
			s.listeners = append(s.listeners, app.GetNetListener())
		*/
		go app.Run()
	}

	s.listen()

	return nil
}

// Reload triggers the reload event of all applications.
func (s *Soul) Reload() {
	for _, app := range s.apps {
		app.Reload()
	}
}

// Reincarnate terminates the current process and respawn another one in its place.
func (s *Soul) Reincarnate() {
	// modify listener file discriptor
	// fork
}

func (s *Soul) exit() {
	// Now it's time to rest in peace
	wg := &sync.WaitGroup{}
	wg.Add(len(s.apps))

	for _, app := range s.apps {
		go func(app Application) {
			app.Shutdown(time.Duration(5) * time.Second)
			wg.Done()
		}(app)
	}

	wg.Wait()
}

func (s *Soul) listen() {
	signalChannel := make(chan os.Signal, 1)

	signal.Notify(signalChannel, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)

	done := false

	for !done {
		sig := <-signalChannel

		switch sig {
		case syscall.SIGUSR1:
			s.Reload() // Reload configuration
		case syscall.SIGUSR2:
			s.Reincarnate() // Update binary
		default:
			s.exit()
			done = true
		}
	}

}
