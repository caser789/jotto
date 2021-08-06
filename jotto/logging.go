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
	Trace(format string, v ...interface{})
	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Data(format string, v ...interface{})
	Warning(format string, v ...interface{})
	Error(format string, v ...interface{})
	Fatal(format string, v ...interface{})
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

func (l *StdoutLogger) Trace(format string, v ...interface{}) {
	l.log("TRACE", format, v...)
}
func (l *StdoutLogger) Debug(format string, v ...interface{}) {
	l.log("DEBUG", format, v...)
}
func (l *StdoutLogger) Info(format string, v ...interface{}) {
	l.log("INFO", format, v...)
}
func (l *StdoutLogger) Data(format string, v ...interface{}) {
	l.log("DATA", format, v...)
}
func (l *StdoutLogger) Warning(format string, v ...interface{}) {
	l.log("WARN", format, v...)
}
func (l *StdoutLogger) Error(format string, v ...interface{}) {
	l.log("ERROR", format, v...)
}
func (l *StdoutLogger) Fatal(format string, v ...interface{}) {
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
