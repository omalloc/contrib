package orm

import (
	"context"

	"gorm.io/gorm"
)

// txContextKey gorm database transaction context key
type txContextKey struct{}

type Transaction interface {
	WithContext(context.Context, func(ctx context.Context, tx *gorm.DB) error) error
}

type transactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(dataSource *gorm.DB) Transaction {
	return &transactionManager{
		db: dataSource,
	}
}

func (tm *transactionManager) WithContext(ctx context.Context, fn func(ctx context.Context, tx *gorm.DB) error) error {
	if tx, ok := ctx.Value(txContextKey{}).(*gorm.DB); ok {
		return fn(ctx, tx)
	}

	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txContextKey{}, tx), tx)
	})
}
