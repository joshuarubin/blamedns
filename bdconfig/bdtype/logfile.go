package bdtype

import (
	"encoding"
	"io"
	"os"

	"jrubin.io/cliconfig"
)

const defaultLogFileName = "stderr"

type LogFile struct {
	io.Writer
	Name string
}

var (
	_ cliconfig.CustomType     = LogFile{}
	_ encoding.TextMarshaler   = LogFile{}
	_ encoding.TextUnmarshaler = &LogFile{}
)

func DefaultLogFile() LogFile {
	file, _ := parseLogFileName(defaultLogFileName)
	return *file
}

func (l LogFile) Default(name string) interface{} {
	return defaultLogFileName
}

func (l LogFile) Equal(val interface{}) bool {
	if sval, ok := val.(string); ok {
		return l.Name == sval
	}
	return false
}

func (l LogFile) UnmarshalCLIConfig(text string) (interface{}, error) {
	var ret LogFile
	if err := ret.UnmarshalText([]byte(text)); err != nil {
		return nil, err
	}
	return ret, nil
}

func (l LogFile) MarshalText() ([]byte, error) {
	return []byte(l.Name), nil
}

func (l *LogFile) UnmarshalText(text []byte) error {
	tmp, err := parseLogFileName(string(text))
	if err != nil {
		return err
	}

	*l = *tmp

	return nil
}

func (l LogFile) File() *os.File {
	if l.Writer == os.Stderr || l.Writer == os.Stdout {
		return nil
	}

	if f, ok := l.Writer.(*os.File); ok {
		return f
	}

	return nil
}

func parseLogFileName(file string) (*LogFile, error) {
	switch file {
	case "stderr", "STDERR":
		return &LogFile{
			Writer: os.Stderr,
			Name:   "stderr",
		}, nil
	case "stdout", "STDOUT":
		return &LogFile{
			Writer: os.Stdout,
			Name:   "stdout",
		}, nil
	default:
		f, err := os.OpenFile(file, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
		// file is only closed on os.Exit
		if err != nil {
			return nil, err
		}

		return &LogFile{
			Writer: f,
			Name:   file,
		}, nil
	}
}
