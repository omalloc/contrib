package crud_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/omalloc/contrib/kratos/orm/crud"
	"github.com/omalloc/contrib/protobuf"
)

// 测试用的模型结构
type TestModel struct {
	ID        int64     `gorm:"primarykey"`
	Name      string    `gorm:"column:name"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func setupTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *gorm.DB) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	dialector := mysql.New(mysql.Config{
		Conn:                      mockDB,
		SkipInitializeWithVersion: true,
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)

	return mockDB, mock, db
}

func TestCRUD_Create(t *testing.T) {
	mockDB, mock, db := setupTestDB(t)
	defer mockDB.Close()

	crud := crud.New[TestModel](db)
	ctx := context.Background()

	testModel := &TestModel{
		Name:      "test",
		CreatedAt: time.Now(),
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `test_models`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := crud.Create(ctx, testModel)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), testModel.ID)
}

func TestCRUD_Update(t *testing.T) {
	mockDB, mock, db := setupTestDB(t)
	defer mockDB.Close()

	crud := crud.New[TestModel](db)
	ctx := context.Background()

	testModel := &TestModel{
		ID:   1,
		Name: "updated",
	}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `test_models`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := crud.Update(ctx, testModel.ID, testModel)
	assert.NoError(t, err)
}

func TestCRUD_Delete(t *testing.T) {
	mockDB, mock, db := setupTestDB(t)
	defer mockDB.Close()

	crud := crud.New[TestModel](db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM `test_models`").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := crud.Delete(ctx, 1)
	assert.NoError(t, err)
}

func TestCRUD_SelectList(t *testing.T) {
	mockDB, mock, db := setupTestDB(t)
	defer mockDB.Close()

	crud := crud.New[TestModel](db)
	ctx := context.Background()
	pagination := &protobuf.Pagination{
		Current:  1,
		PageSize: 10,
	}

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `test_models`").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery("SELECT \\* FROM `test_models`").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at"}).
			AddRow(1, "test1", time.Now()).
			AddRow(2, "test2", time.Now()))

	list, err := crud.SelectList(ctx, pagination)
	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestCRUD_SelectOne(t *testing.T) {
	mockDB, mock, db := setupTestDB(t)
	defer mockDB.Close()

	crud := crud.New[TestModel](db)
	ctx := context.Background()

	expectedTime := time.Now()
	mock.ExpectQuery("SELECT \\* FROM `test_models`").
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "created_at"}).
			AddRow(1, "test", expectedTime))

	model, err := crud.SelectOne(ctx, 1)
	assert.NoError(t, err)
	assert.NotNil(t, model)
	assert.Equal(t, int64(1), model.ID)
	assert.Equal(t, "test", model.Name)
	assert.Equal(t, expectedTime, model.CreatedAt)
}
