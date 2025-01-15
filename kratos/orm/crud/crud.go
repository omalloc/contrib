package crud

import (
	"context"

	"gorm.io/gorm"

	"github.com/omalloc/contrib/protobuf"
)

type CRUD[T any] interface {
	Create(ctx context.Context, t *T) error
	Update(ctx context.Context, id int64, t *T) error
	Delete(ctx context.Context, id int64) error
	SelectList(ctx context.Context, pagination *protobuf.Pagination) ([]*T, error)
	SelectOne(ctx context.Context, id int64) (*T, error)
}

type crud[T any] struct {
	db *gorm.DB
}

func (r *crud[T]) Create(ctx context.Context, t *T) error {
	return r.db.WithContext(ctx).
		Model(new(T)).
		Create(t).Error
}

func (r *crud[T]) Update(ctx context.Context, id int64, t *T) error {
	return r.db.WithContext(ctx).
		Model(new(T)).
		Where("id = ?", id).
		Save(t).Error
}

func (r *crud[T]) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(new(T)).Error
}

func (r *crud[T]) SelectList(ctx context.Context, pagination *protobuf.Pagination) ([]*T, error) {
	var (
		list []*T
		err  error
	)
	err = r.db.WithContext(ctx).Model(new(T)).
		Count(pagination.Count()).
		Offset(pagination.Offset()).
		Limit(pagination.Limit()).
		Find(&list).Error

	return list, err
}

func (r *crud[T]) SelectOne(ctx context.Context, id int64) (*T, error) {
	var (
		t   T
		err error
	)
	err = r.db.WithContext(ctx).
		Model(&t).
		Where("id = ?", id).
		Take(&t).Error

	return &t, err
}

func New[T any](db *gorm.DB) CRUD[T] {
	return &crud[T]{db: db}
}
