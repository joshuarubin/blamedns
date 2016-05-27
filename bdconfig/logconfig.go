package bdconfig

import (
	"io"
	"io/ioutil"

	"github.com/Sirupsen/logrus"

	"jrubin.io/blamedns/bdconfig/bdtype"
)

type LogConfig struct {
	File  bdtype.LogFile  `toml:"file" cli:",set log filename (stderr, stdout or any file name)"`
	Level bdtype.LogLevel `toml:"level" cli:",set log level (debug, info, warning, error)"`
}

var defaultLogConfig = LogConfig{
	File:  bdtype.DefaultLogFile(),
	Level: bdtype.DefaultLogLevel,
}

type logHook struct {
	Logger *logrus.Logger
}

func (h logHook) Levels() []logrus.Level {
	ret := make([]logrus.Level, h.Logger.Level+1)

	for i := logrus.PanicLevel; i <= h.Logger.Level; i++ {
		ret[i] = i
	}

	return ret
}

func (h logHook) Fire(entry *logrus.Entry) error {
	ctxLog := h.Logger.WithFields(entry.Data)

	switch entry.Level {
	case logrus.PanicLevel:
		ctxLog.Panic(entry.Message)
	case logrus.FatalLevel:
		ctxLog.Fatal(entry.Message)
	case logrus.ErrorLevel:
		ctxLog.Error(entry.Message)
	case logrus.WarnLevel:
		ctxLog.Warn(entry.Message)
	case logrus.InfoLevel:
		ctxLog.Info(entry.Message)
	case logrus.DebugLevel:
		ctxLog.Debug(entry.Message)
	}

	return nil
}

func newLogHook(out io.Writer, level logrus.Level) *logHook {
	logger := logrus.New()
	logger.Out = out
	logger.Level = level

	return &logHook{Logger: logger}
}

func (cfg LogConfig) Init(root *Config) {
	logger := root.Logger

	hookLogger := newLogHook(cfg.File, cfg.Level.Level())

	if f := cfg.File.File(); f != nil {
		logger.WithField("name", cfg.File.Name).Info("log location")

		hookLogger.Logger.Formatter = &logrus.TextFormatter{
			DisableColors: true,
		}
	}

	// we want to use logrus JUST for the hooks, so we send all it's data to the
	// bit bucket, but turn on debug level logging so that hooks are always
	// called
	logger.Out = ioutil.Discard
	logger.Level = logrus.DebugLevel
	logger.Hooks.Add(hookLogger)
	logger.WithField("level", hookLogger.Logger.Level).Debug("log level set")
}

func (cfg LogConfig) Shutdown() {
	if f := cfg.File.File(); f != nil {
		_ = f.Close()
	}
}
