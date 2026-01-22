package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type Level = zerolog.Level

const (
	LevelDebug = zerolog.DebugLevel
	LevelInfo  = zerolog.InfoLevel
	LevelWarn  = zerolog.WarnLevel
	LevelError = zerolog.ErrorLevel
)

type Fields map[string]interface{}

var std zerolog.Logger
var colorEnabled bool

// Logger is a small wrapper around zerolog.Logger to preserve the previous
// package API where callers expect a *logger.Logger with methods like Info.
type Logger struct {
	Z zerolog.Logger
}

// NewConsole creates a zerolog ConsoleWriter-backed logger. When color is true
// the console writer will emit ANSI colors. Time format matches "3:04PM".
func NewConsole(out io.Writer, level Level, color bool) *Logger {
	cw := zerolog.ConsoleWriter{Out: out, TimeFormat: "3:04PM", NoColor: !color}
	colorEnabled = color

	// helper to wrap ANSI color codes when `color` is true
	colorWrap := func(s string, code string) string {
		if !color {
			return s
		}
		return "\x1b[" + code + "m" + s + "\x1b[0m"
	}

	// Format short level codes and optionally color them.
	cw.FormatLevel = func(i interface{}) string {
		s := zerolog.NoLevel
		switch v := i.(type) {
		case string:
			if lvl, err := zerolog.ParseLevel(v); err == nil {
				s = lvl
			} else {
				s = zerolog.NoLevel
			}
		case zerolog.Level:
			s = v
		}
		switch s {
		case zerolog.DebugLevel:
			return colorWrap("DBG", "36") // cyan
		case zerolog.InfoLevel:
			return colorWrap("INF", "32") // green
		case zerolog.WarnLevel:
			return colorWrap("WRN", "33") // yellow
		case zerolog.ErrorLevel:
			return colorWrap("ERR", "31") // red
		default:
			return ""
		}
	}

	// Timestamp printed as "3:04PM" and dimmed when color enabled.
	cw.FormatTimestamp = func(i interface{}) string {
		switch v := i.(type) {
		case time.Time:
			return colorWrap(v.Format("3:04PM"), "2")
		case string:
			return colorWrap(v, "2")
		default:
			return ""
		}
	}

	l := zerolog.New(cw).With().Timestamp().Logger().Level(level)
	return &Logger{Z: l}
}

// Colorize wraps the provided string with ANSI color codes when colors are enabled.
// The `code` is the numeric SGR color code (e.g. "31" for red).
func Colorize(s string, code string) string {
	if !colorEnabled {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

// SetStd replaces the package logger used by wrapper functions.
func SetStd(l *Logger) {
	if l == nil {
		std = zerolog.Nop()
		return
	}
	std = l.Z
}

func ensureStd() {
	if std.GetLevel() == zerolog.NoLevel {
		// default console logger with color enabled
		std = NewConsole(os.Stdout, zerolog.InfoLevel, true).Z
	}
}

func Info(msg string, f Fields) {
	ensureStd()
	e := std.Info()
	if f != nil {
		e = e.Fields(f)
	}
	e.Msg(msg)
}

func Debug(msg string, f Fields) {
	ensureStd()
	e := std.Debug()
	if f != nil {
		e = e.Fields(f)
	}
	e.Msg(msg)
}

func Warn(msg string, f Fields) {
	ensureStd()
	e := std.Warn()
	if f != nil {
		e = e.Fields(f)
	}
	e.Msg(msg)
}

func Error(msg string, f Fields) {
	ensureStd()
	e := std.Error()
	if f != nil {
		e = e.Fields(f)
	}
	e.Msg(msg)
}

// NewSimple returns a ConsoleWriter-backed logger wrapped in *Logger.
func NewSimple(out io.Writer, level Level, color bool) *Logger {
	return NewConsole(out, level, color)
}

// NewNop returns a no-op Logger instance.
func NewNop() *Logger {
	return &Logger{Z: zerolog.Nop()}
}

// Methods to allow using *Logger where previous code expected it.
func (l *Logger) Info(msg string, f Fields) {
	e := l.Z.Info()
	if f != nil {
		e = e.Fields(f)
	}
	e.Msg(msg)
}

func (l *Logger) Debug(msg string, f Fields) {
	e := l.Z.Debug()
	if f != nil {
		e = e.Fields(f)
	}
	e.Msg(msg)
}

func (l *Logger) Warn(msg string, f Fields) {
	e := l.Z.Warn()
	if f != nil {
		e = e.Fields(f)
	}
	e.Msg(msg)
}

func (l *Logger) Error(msg string, f Fields) {
	e := l.Z.Error()
	if f != nil {
		e = e.Fields(f)
	}
	e.Msg(msg)
}
