package slog

import (
	"bufio"
	"io"
	"runtime"
)

// Writer returns an io.Pipe where each line written to that writer will be
// printed using the handlers for the given Level. It is the caller's
// responsibility to close it.
func (logger *Logger) Writer(level Level) *io.PipeWriter {
	reader, writer := io.Pipe()

	if level < PanicLevel {
		level = PanicLevel
	}

	if level > DebugLevel {
		level = DebugLevel
	}

	var printFunc func(msg string)
	switch level {
	case DebugLevel:
		printFunc = logger.Debug
	case InfoLevel:
		printFunc = logger.Info
	case WarnLevel:
		printFunc = logger.Warn
	case ErrorLevel:
		printFunc = logger.Error
	case FatalLevel:
		printFunc = logger.Fatal
	case PanicLevel:
		printFunc = logger.Panic
	}

	go logger.writerScanner(reader, printFunc)
	runtime.SetFinalizer(writer, writerFinalizer)

	return writer
}

func (logger *Logger) writerScanner(reader io.ReadCloser, printFunc func(msg string)) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		printFunc(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		logger.WithError(err).Error("Error while reading from Writer")
	}

	_ = reader.Close()
}

func writerFinalizer(writer *io.PipeWriter) {
	_ = writer.Close()
}
