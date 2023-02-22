package logs

import (
	"time"
)

var (
	TimeFieldName        = "time"
	TraceFieldName       = "trace"
	LevelFieldName       = "level"
	MsgFieldName         = "msg"
	ErrorFieldName       = "error"
	CallerFieldName      = "caller"
	TimeFieldFormat      = "2006/01/02 15:04:05.000"
	DurationFieldUnit    = time.Millisecond
	DurationFieldInteger = false
)
