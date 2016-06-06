package slog

// Interface represents the API of both Logger and Entry.
type Interface interface {
	WithFields(fields Fielder) *Entry
	WithField(key string, value interface{}) *Entry
	WithError(err error) *Entry
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)
	Panic(msg string)
	Trace(level Level, msg string) *Entry
}
