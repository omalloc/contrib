package orm

import (
	"context"
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	glog "gorm.io/gorm/logger"
)

type gormLogger struct {
	debug                 bool
	dbLog                 *log.Helper
	SlowThreshold         time.Duration
	SourceField           string
	SkipCallerLookup      bool
	SkipErrRecordNotFound bool
}
type GormLoggerOption func(*gormLogger)

func NewLogger(opts ...GormLoggerOption) *gormLogger {
	// default options
	r := &gormLogger{
		debug:                 false,
		dbLog:                 log.NewHelper(log.GetLogger()),
		SlowThreshold:         500 * time.Millisecond, // 500毫秒查询 + 500毫秒业务响应 = 1s 用户最佳体验 loading 之内，超过则属于慢查询
		SkipCallerLookup:      false,
		SkipErrRecordNotFound: true,
	}
	// apply custom options
	for _, opt := range opts {
		opt(r)
	}

	return r
}

func WithDebug() GormLoggerOption {
	return func(logger *gormLogger) {
		logger.debug = true
	}
}

// WithLogHelper set gorm-logger and log filters..
func WithLogHelper(klog log.Logger) GormLoggerOption {
	return func(logger *gormLogger) {
		logger.dbLog = log.NewHelper(klog)
	}
}

func WIthSlowThreshold(threshold time.Duration) GormLoggerOption {
	return func(logger *gormLogger) {
		logger.SlowThreshold = threshold
	}
}

func WithSkipCallerLookup(skip bool) GormLoggerOption {
	return func(logger *gormLogger) {
		logger.SkipCallerLookup = skip
	}
}

func WithSkipErrRecordNotFound(skip bool) GormLoggerOption {
	return func(logger *gormLogger) {
		logger.SkipErrRecordNotFound = skip
	}
}

func (gl *gormLogger) LogMode(level glog.LogLevel) glog.Interface {
	return gl
}

func (gl *gormLogger) Info(ctx context.Context, s string, args ...interface{}) {
	gl.dbLog.Info(s, args)
}

func (gl *gormLogger) Warn(ctx context.Context, s string, args ...interface{}) {
	gl.dbLog.Warn(s, args)
}

func (gl *gormLogger) Error(ctx context.Context, s string, args ...interface{}) {
	gl.dbLog.Error(s, args)
}

func (gl *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if !gl.debug {
		return
	}

	elapsed := time.Since(begin)
	timeUsed := float64(elapsed.Nanoseconds()) / 1e6

	fields := make([]interface{}, 0)
	fields = append(fields, "timeUsed", timeUsed)

	sql, rows := fc()

	switch {
	//  check err
	case err != nil && (gl.SkipErrRecordNotFound && !(errors.Is(err, gorm.ErrRecordNotFound))):
		fields = append(fields, "err", err)
		fields = append(fields, "sql", sql)
		fields = append(fields, "rows", rows)
		gl.dbLog.Errorw(fields...)
	// check slow query
	case elapsed > gl.SlowThreshold && gl.SlowThreshold != 0:
		fields = append(fields, "sql", sql)
		fields = append(fields, "rows", rows)
		fields = append(fields, "slowElapsed", gl.SlowThreshold)
		gl.dbLog.Warnw(fields...)
	// normal
	default:
		fields = append(fields, "sql", sql)
		fields = append(fields, "rows", rows)
		gl.dbLog.Debugw(fields...)
	}
}
