package cfgssh

import (
	logg "log"
)

type (
	// DebugLogger wrap *log.Loggers with this to display debug logging
	DebugLogger *logg.Logger

	// InfoLogger wrap *log.Loggers with this to display info logging
	InfoLogger *logg.Logger
)

// log is a struct that provides debug and info logging. This name was intentionally chosen
// so that it conflicts the the std log package. And forces contributors to use this struct
// instead of the std logger.
type log struct{ debug, info *logg.Logger }

// OnErr checks to see if err is nil, if it is nil, then no error message is displayed. If
// it is not nil, then the message is displayed. It is a convenience method for the typical
// if err != nil conditional.
func (l log) OnErr(err error) log {
	if err != nil {
		return l
	}
	return log{} // they will be nil, so they won't log
}

func (l log) Debug(v ...interface{}) {
	if l.debug != nil {
		l.debug.Println(v...)
	}
}

func (l log) Debugf(fmt string, v ...interface{}) {
	if l.debug != nil {
		l.debug.Printf(fmt, v...)
	}
}

func (l log) Info(v ...interface{}) {
	if l.info != nil {
		l.info.Println(v...)
	}
}

func (l log) Infof(fmt string, v ...interface{}) {
	if l.info != nil {
		l.info.Printf(fmt, v...)
	}
}
