package context

import (
	"os"
	"strings"

	"github.com/pkg/errors"

	"jrubin.io/blamedns/config"
	"jrubin.io/slog"
	"jrubin.io/slog/handlers/json"
	"jrubin.io/slog/handlers/text"
)

var defaultLogLevel = slog.WarnLevel

type LogContext struct {
	Logger *slog.Logger
	Level  slog.Level
	File   *os.File
}

func ParseFileFlag(name string) (*os.File, error) {
	lname := strings.ToLower(name)

	if lname == "stdout" || name == os.Stdout.Name() {
		return os.Stdout, nil
	}

	if lname == "stderr" || name == os.Stderr.Name() {
		return os.Stderr, nil
	}

	f, err := os.Create(name)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating log file: %s", name)
	}

	return f, nil
}

func NewLogContext(cfg *config.LogConfig) (*LogContext, error) {
	ctx := &LogContext{
		Logger: slog.New(),
		Level:  slog.ParseLevel(cfg.Level, defaultLogLevel),
	}

	var err error
	if ctx.File, err = ParseFileFlag(cfg.File); err != nil {
		return nil, err
	}

	if cfg.JSON || (ctx.File != os.Stdout && ctx.File != os.Stderr) {
		ctx.Logger.RegisterHandler(ctx.Level, json.New(ctx.File))
	} else {
		ctx.Logger.RegisterHandler(ctx.Level, text.New(ctx.File))
	}

	ctx.Logger.WithField("level", cfg.Level).Debug("log level set")
	return ctx, nil
}

func (ctx LogContext) Shutdown() {
	if f := ctx.File; f != nil {
		_ = f.Close()
	}
}
