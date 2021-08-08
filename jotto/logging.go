package motto

import (
	"fmt"
	"strings"

	"github.com/rs/xid"
)

// Logger is the interface used for logging in Motto.
// Custom loggers implemented by application must conform
// to this interface.
type Logger interface {
	Tracef(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Dataf(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})
}

// LoggerContext is a container to store contextual information
// for loggers.
type LoggerContext map[string]interface{}

func (lc *LoggerContext) String() string {
	format := strings.Repeat("%s:%+v,", len(*lc))
	values := make([]interface{}, 2*len(*lc))

	idx := 0
	for k, v := range *lc {
		values[idx] = k
		idx++
		values[idx] = v
		idx++
	}

	text := fmt.Sprintf(format, values...)

	return strings.Trim(text, ",")
}

// LoggerFactory creates a Logger given a LoggerContext.
type LoggerFactory func(Application, LoggerContext) Logger

// GenerateTraceID generates a unique trace id for log entries within a request cycle.
func GenerateTraceID() string {
	return xid.New().String()
}

// NewStdoutLogger creates a new logger that logs everything to stdout.
func NewStdoutLogger(c LoggerContext) *StdoutLogger {
	return &StdoutLogger{
		context: c,
	}
}

// StdoutLogger prints logs to stdout
type StdoutLogger struct {
	context LoggerContext
}

// Tracef logs at the `trace` level
func (l *StdoutLogger) Tracef(format string, v ...interface{}) {
	l.log("TRACE", format, v...)
}

// Debugf logs at the `debug` level
func (l *StdoutLogger) Debugf(format string, v ...interface{}) {
	l.log("DEBUG", format, v...)
}

// Infof logs at the `info` level
func (l *StdoutLogger) Infof(format string, v ...interface{}) {
	l.log("INFO", format, v...)
}

// Dataf logs at the `data` level
func (l *StdoutLogger) Dataf(format string, v ...interface{}) {
	l.log("DATA", format, v...)
}

// Warnf logs at the `warn` level
func (l *StdoutLogger) Warnf(format string, v ...interface{}) {
	l.log("WARN", format, v...)
}

// Errorf logs at the `error` level
func (l *StdoutLogger) Errorf(format string, v ...interface{}) {
	l.log("ERROR", format, v...)
}

// Fatalf logs at the `fatal` level
func (l *StdoutLogger) Fatalf(format string, v ...interface{}) {
	l.log("FATAL", format, v...)
}

func (l *StdoutLogger) log(level, format string, v ...interface{}) {

	values := make([]interface{}, len(v)+2)

	values[0] = level
	values[1] = l.context.String()

	idx := 2
	for _, val := range v {
		values[idx] = val
		idx++
	}

	fmt.Printf("[%s][%s]"+format+"\n", values...)
}
