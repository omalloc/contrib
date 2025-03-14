package zap

import (
	kratoszap "github.com/go-kratos/kratos/contrib/log/zap/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Level = zapcore.Level

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel Level = iota - 1
	// InfoLevel is the default logging priority.
	InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel
	// DPanicLevel logs are particularly important errors. In development the
	// logger panics after writing the message.
	DPanicLevel
	// PanicLevel logs a message, then panics.
	PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel
)

type Config = zap.Config

type Option func(config *Config)

// WithEncoding setter log encoding. only support `json` and `console`
func WithEncoding(encoding string) Option {
	return func(cfg *Config) {
		cfg.Encoding = encoding
	}
}

// WithLevel setter log level. alias zap.config.OutputPaths
func WithLevel(lvl Level) Option {
	return func(cfg *Config) {
		cfg.Level = zap.NewAtomicLevelAt(lvl)
	}
}

// WithLevelString setter log level (string-alias)
func WithLevelString(lvl string) Option {
	return func(cfg *Config) {
		l, err := zapcore.ParseLevel(lvl)
		if err == nil {
			cfg.Level = zap.NewAtomicLevelAt(l)
		}
	}
}

// WithOutput setter log output paths. valid `stdout` `stderr`
func WithOutput(paths []string) Option {
	return func(cfg *Config) {
		cfg.OutputPaths = paths
	}
}

// Verbose alias WithLevel setter log level to DebugLevel
func Verbose(flag bool) Option {
	if flag {
		return WithLevel(DebugLevel)
	}
	return WithLevel(InfoLevel)
}

// New returns a zap logger.
func New(opts ...Option) *zap.Logger {
	cfg := zap.NewProductionConfig()

	// 防止 kratos log 的 key 重复; took https://github.com/go-kratos/kratos/issues/1722
	// cfg.EncoderConfig.MessageKey = "" // Updated@2025-03-14 msg key lost; took https://github.com/go-kratos/kratos/pull/3171
	cfg.EncoderConfig.TimeKey = ""
	cfg.EncoderConfig.CallerKey = ""
	// 默认输出到 stdout
	cfg.OutputPaths = []string{"stdout"}

	// 默认等级为 debug
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)

	for _, o := range opts {
		o(&cfg)
	}

	logger, _ := cfg.Build()
	return logger
}

// NewLogger returns a wrapped zap logger of kratos.
func NewLogger(opts ...Option) *kratoszap.Logger {
	return kratoszap.NewLogger(New(opts...))
}
