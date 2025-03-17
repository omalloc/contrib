package orm

import (
	"context"

	"gorm.io/gorm"
)

// txContextKey gorm database transaction context key
type txContextKey struct{}

type DataSourceManager interface {
	GetDataSource() *gorm.DB
}

type Transaction interface {
	Transaction(context.Context, func(ctx context.Context) error) error
	WithContext(context.Context) *gorm.DB
}

type transactionManager struct {
	dsm DataSourceManager
}

func NewTransactionManager(dsm DataSourceManager) Transaction {
	return &transactionManager{
		dsm: dsm,
	}
}

func (tm *transactionManager) WithContext(ctx context.Context) *gorm.DB {
	if ctx == nil {
		return tm.dsm.GetDataSource()
	}

	if tx, ok := ctx.Value(txContextKey{}).(*gorm.DB); ok {
		return tx
	}

	return tm.dsm.GetDataSource().WithContext(ctx)
}

func (tm *transactionManager) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if ctx == nil {
		return tm.with(context.Background(), fn)
	}

	if _, ok := ctx.Value(txContextKey{}).(*gorm.DB); ok {
		return fn(ctx)
	}

	return tm.with(ctx, fn)
}

func (tm *transactionManager) with(ctx context.Context, fn func(ctx context.Context) error) error {
	return tm.dsm.GetDataSource().Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txContextKey{}, tx))
	})
}
