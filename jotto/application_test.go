package motto_test

import (
	"testing"
	"time"

	"git.garena.com/duanzy/motto/motto"

	"github.com/stretchr/testify/assert"
)

func TestDaemonRunsForeverWithApplication(t *testing.T) {
	cfg := motto.NewDefaultSettings()
	cfg.Motto().Protocol = "HTTP"
	app := motto.NewApplication(cfg, nil, nil, nil)

	cases := []struct {
		name    string
		worker  motto.DaemonWorker
		done    bool
		timeout bool
	}{
		{
			"forever",
			func(app motto.Application, cancel <-chan struct{}, args ...interface{}) {
				for {
					time.Sleep(time.Millisecond * 100)
				}
			}, false, true,
		},
		{
			"return",
			func(app motto.Application, cancel <-chan struct{}, args ...interface{}) {
				return
			}, true, false,
		},
	}

	for _, tcase := range cases {
		app.RegisterDaemon(tcase.name, tcase.worker)
	}

	app.Boot()
	go app.Run()

	var (
		daemon motto.Daemon
		err    error
	)
	for _, tcase := range cases {
		if daemon, err = app.GetDaemon(tcase.name); err != nil {
			t.Fatalf("get daemon `%s` failed", tcase.name)
		}

		var done, timeout bool
		select {
		case <-daemon.Done():
			done = true
		case <-time.After(time.Second):
			timeout = true
		}

		assert.Equal(t, tcase.done, done)
		assert.Equal(t, tcase.timeout, timeout)
	}
}

func TestDaemonCanBeCancelled(t *testing.T) {
	cfg := motto.NewDefaultSettings()
	cfg.Motto().Protocol = "HTTP"
	app := motto.NewApplication(cfg, nil, nil, nil)
	daemon := app.RegisterDaemon("daemon:test", func(app motto.Application, cancel <-chan struct{}, args ...interface{}) {
		for {
			select {
			case <-cancel:
				return
			default:
				// Run forever if not cancelled
				time.Sleep(time.Millisecond * 100)
			}
		}
	}, []interface{}{"hello", "motto"})

	app.Boot()
	go app.Run()

	daemon.Cancel()

	var done, timeout bool
	select {
	case <-daemon.Done():
		done = true
	case <-time.After(time.Second):
		timeout = true
	}

	assert.True(t, done)
	assert.False(t, timeout)
}
