package common

import (
	"git.garena.com/common/gocommon"
	"git.garena.com/duanzy/motto/motto"
)

func NewCommonLogger(app motto.Application, context motto.LoggerContext) motto.Logger {
	logger := &CommonLogger{
		context: context,
	}

	return logger
}

type CommonLogger struct {
	context motto.LoggerContext
}

func (l *CommonLogger) Tracef(format string, v ...interface{}) {
	l.log(gocommon.Logf, "TRACE", format, v...)
}
func (l *CommonLogger) Debugf(format string, v ...interface{}) {
	l.log(gocommon.Logf, "DEBUG", format, v...)
}
func (l *CommonLogger) Infof(format string, v ...interface{}) {
	l.log(gocommon.Logf, "INFO", format, v...)
}
func (l *CommonLogger) Dataf(format string, v ...interface{}) {
	l.log(gocommon.Logf, "DATA", format, v...)
}
func (l *CommonLogger) Warnf(format string, v ...interface{}) {
	l.log(gocommon.Logf, "WARN", format, v...)
}
func (l *CommonLogger) Errorf(format string, v ...interface{}) {
	l.log(gocommon.Logf, "ERROR", format, v...)
}
func (l *CommonLogger) Fatalf(format string, v ...interface{}) {
	l.log(gocommon.Logf, "FATAL", format, v...)
}

func (l *CommonLogger) log(logf func(string, ...interface{}), level, format string, v ...interface{}) {
	values := make([]interface{}, len(v)+2)

	values[0] = level
	values[1] = l.context.String()

	idx := 2
	for _, val := range v {
		values[idx] = val
		idx++
	}

	logf("[%s]<%s> "+format, values...)

}
