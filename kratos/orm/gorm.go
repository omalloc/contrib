package orm

import (
	"errors"

	"gorm.io/gorm"
	glog "gorm.io/gorm/logger"
)

var (
	ErrDriverNotFound = errors.New("gorm: driver not found")
)

type DataConf interface {
	GetDriver() string
	GetSource() string
}

// Config a gorm custom config
type Config struct {
	opts     *gorm.Config
	driver   gorm.Dialector
	log      glog.Interface
	tracer   *GormOpenTelemetryPlugin
	tracing  bool
	hasDebug bool
}

type Option func(*Config)

// WithDriver set gorm-driver.
func WithDriver(driver gorm.Dialector) Option {
	return func(c *Config) {
		c.driver = driver
	}
}

// WithTracing set gorm-tracing. used for opentelemetry.
func WithTracing() Option {
	return func(c *Config) {
		c.tracing = true
	}
}

// WithTracingOpts set gorm-tracing some opts. used for opentelemetry.
func WithTracingOpts(opts ...TraceOption) Option {
	return func(c *Config) {
		c.tracing = true
		c.tracer = NewTracer(opts...)
	}
}

// WithLogger set gorm-logger and has debug logger writer..
func WithLogger(opts ...GormLoggerOption) Option {
	return func(c *Config) {
		c.log = NewLogger(opts...)
	}
}

func New(opts ...Option) (*gorm.DB, error) {
	c := &Config{
		hasDebug: false,
		tracing:  false,
		opts:     nil,
		log:      glog.Default,
	}
	for _, o := range opts {
		o(c)
	}

	c.opts = &gorm.Config{
		Logger:               c.log, // 默认使用 gorm.Default
		CreateBatchSize:      1000,  // 默认 1000
		AllowGlobalUpdate:    false, // 默认不允许全表更新
		DisableAutomaticPing: false, // 默认不禁用自动ping (数据库连接保活)
	}

	if c.driver == nil {
		return nil, ErrDriverNotFound
	}

	db, err := gorm.Open(c.driver, c.opts)
	if err != nil {
		return nil, err
	}

	// set opentelemetry tracing.
	if c.tracing {
		if c.tracer == nil {
			c.tracer = NewTracer()
		}
		_ = db.Use(c.tracer)
	}

	return db, nil
}
