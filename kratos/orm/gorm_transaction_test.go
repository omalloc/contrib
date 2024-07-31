package orm_test

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	"github.com/omalloc/contrib/kratos/orm"
)

var cfg = `
server:
  http:
    addr: 0.0.0.0:8000
    timeout: 1s
  grpc:
    addr: 0.0.0.0:9000
    timeout: 1s
data:
  database:
    driver: sqlite
    source: file:test.db?cache=shared&mode=memory
  redis:
    addr: 127.0.0.1:6379
    read_timeout: 0.2s
    write_timeout: 0.2s
`

type Data struct {
	db *gorm.DB
}

func newData() *Data {
	db, err := orm.New(
		orm.WithDriver(sqlite.Open("file:test.db?cache=shared&mode=memory&charset=utf8mb4&parseTime=true&loc=Local")),
		orm.WithTracingOpts(orm.WithDatabaseName("test")),
		orm.WithLogger(
			orm.WithDebug(),
			orm.WIthSlowThreshold(time.Second*2),
			orm.WithSkipCallerLookup(true),
			orm.WithSkipErrRecordNotFound(true),
			orm.WithLogHelper(log.NewFilter(log.GetLogger(), log.FilterLevel(log.LevelDebug))),
		),
	)
	if err != nil {
		panic(err)
	}

	return &Data{
		db: db,
	}
}

func TestTransaction(t *testing.T) {
	data := newData()
	data.db.AutoMigrate(&User{})

	txm := orm.NewTransactionManager(data.db)

	txm.WithContext(context.TODO(), func(ctx context.Context, tx *gorm.DB) error {
		return tx.Model(&User{}).Create(&User{Name: "test1"}).Error
	})
}
