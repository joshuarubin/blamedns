package bdconfig

import (
	"io"
	"os"
)

const defaultLogFileName = "stderr"

type logFile struct {
	io.Writer
	Name   string
	IsFile bool
}

func defaultLogFile() *logFile {
	file, _ := parseLogFileName(defaultLogFileName)
	return file
}

func (l logFile) Default(name string) interface{} {
	return defaultLogFileName
}

func (l logFile) Equal(val interface{}) bool {
	if sval, ok := val.(string); ok {
		return l.Name == sval
	}
	return false
}

func (l logFile) UnmarshalCLIConfig(text string) (interface{}, error) {
	var ret logFile
	if err := ret.UnmarshalText([]byte(text)); err != nil {
		return nil, err
	}
	return ret, nil
}

func (l *logFile) UnmarshalText(text []byte) error {
	tmp, err := parseLogFileName(string(text))
	if err != nil {
		return err
	}

	*l = *tmp

	return nil
}

func parseLogFileName(file string) (*logFile, error) {
	switch file {
	case "stderr", "STDERR":
		return &logFile{
			Writer: os.Stderr,
			Name:   "stderr",
			IsFile: false,
		}, nil
	case "stdout", "STDOUT":
		return &logFile{
			Writer: os.Stdout,
			Name:   "stdout",
			IsFile: false,
		}, nil
	default:
		f, err := os.OpenFile(file, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
		// file is only closed on os.Exit
		if err != nil {
			return nil, err
		}

		return &logFile{
			Writer: f,
			Name:   file,
			IsFile: true,
		}, nil
	}
}
