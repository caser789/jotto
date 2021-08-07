package motto

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
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
	fds := s.LoadFileDescriptors()

	for k, app := range s.apps {
		app.Boot()

		if len(fds) > k {
			l, err := net.FileListener(os.NewFile(fds[k], fmt.Sprintf("listen_%d", k)))

			if err != nil {
				log.Fatal("listen failed:", err)
			}

			app.SetListener(l)
		}

		go app.Run()
	}

	fmt.Println("pid", os.Getpid())

	time.Sleep(time.Millisecond * 200)

	for _, app := range s.apps {
		l, err := app.GetListener()
		fmt.Println("listener: ", l, err)
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
func (s *Soul) Reincarnate() (err error) {
	// modify listener file discriptor
	// fork

	var listeners []net.Listener

	for _, app := range s.apps {
		l, err := app.GetListener()

		if err != nil {
			log.Println("reincarnate: failed to get listener", err)
			return err
		}

		listeners = append(listeners, l)
	}

	descriptors := make([]uintptr, len(listeners))
	for index, listener := range listeners {
		fp, err := listener.(*net.TCPListener).File()

		if err != nil {
			log.Println("reincarnate: failed to get file", err)
			return err
		}

		// Prevent file descriptor from being closed on exec
		_, _, err = syscall.Syscall(syscall.SYS_FCNTL, fp.Fd(), syscall.F_SETFD, 0)

		if err != syscall.Errno(0) {
			log.Println("reincarnate: syscall returned non-zero", err)
			return err
		}

		descriptors[index] = fp.Fd()
	}

	s.SaveFileDescriptors(descriptors)

	return s.Respawn(descriptors)
}

func (s *Soul) Respawn(fds []uintptr) (err error) {
	bin := os.Args[0]
	if _, err = os.Stat(bin); err != nil {
		return
	}
	wd, err := os.Getwd()
	if err != nil {
		return
	}
	maxfd := uintptr(0)
	for _, v := range fds {
		if v > maxfd {
			maxfd = v
		}
	}
	files := make([]*os.File, maxfd+1)
	files[syscall.Stdin] = os.Stdin
	files[syscall.Stdout] = os.Stdout
	files[syscall.Stderr] = os.Stderr
	for _, v := range fds {
		files[v] = os.NewFile(v, fmt.Sprintf("tcp:%d", v))
	}
	p, err := os.StartProcess(bin, os.Args, &os.ProcAttr{
		Dir:   wd,
		Env:   os.Environ(),
		Files: files,
		Sys:   &syscall.SysProcAttr{},
	})
	if err != nil {
		return err
	}
	log.Printf("forked pid:%v", p.Pid)

	return
}

func (s *Soul) SaveFileDescriptors(fds []uintptr) {
	str := ""
	for _, fd := range fds {
		str += fmt.Sprintf("%d,", fd)
	}
	os.Setenv("SERVER_FDS", str)
}

func (s *Soul) LoadFileDescriptors() (fds []uintptr) {
	strs := strings.Split(os.Getenv("SERVER_FDS"), ",")
	for _, v := range strs {
		fd, err := strconv.Atoi(v)
		if err == nil {
			fds = append(fds, uintptr(fd))
		}
	}
	return
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
			err := s.Reincarnate() // Update binary
			if err != nil {
				log.Println("reincarnate returned err: ", err)
			}
			s.exit()
		default:
			s.exit()
			done = true
		}
	}

}
