package storage

import "go.uber.org/zap"

type Logger struct {
	*zap.SugaredLogger
}

func (l Logger) Warning(v ...any) {
	l.Warn(v...)
}

func (l Logger) Warningf(format string, v ...any) {
	l.Warnf(format, v...)
}
