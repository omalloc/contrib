package orm_test

import (
	"context"
	"fmt"
	"time"

	// "gorm.io/driver/sqlite" // with cgo sqlite3
	"github.com/glebarez/sqlite" // without cgo sqlite3
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"

	"github.com/omalloc/contrib/kratos/orm"
	"github.com/omalloc/contrib/kratos/orm/crud"
	"github.com/omalloc/contrib/protobuf"
)

type User struct {
	ID   int64  `json:"id" gorm:"column:id;primaryKey;autoIncrement;"`
	Name string `json:"name" gorm:"column:name;"`

	orm.DBModel
}

type userRepo struct {
	crud.CRUD[User]

	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *userRepo {
	return &userRepo{
		CRUD: crud.New[User](db),
		db:   db,
	}
}

func ExampleNew() {
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

	// 创建测试表
	_ = db.Session(&gorm.Session{SkipHooks: true}).AutoMigrate(&User{})

	repo := NewUserRepo(db)

	if err := repo.Create(context.Background(), &User{Name: "test1"}); err != nil {
		println(err.Error())
	}

	if err := repo.Create(context.Background(), &User{Name: "test1"}); err != nil {
		println(err.Error())
	}

	// 自动分页
	pagination := protobuf.PageWrap(nil)
	users, err := repo.SelectList(context.Background(), pagination)
	if err != nil {
		println(err.Error())
	}

	if len(users) != int(pagination.Resp().Total) {
		println("data total size not equal, got %d want %d", len(users), pagination.Resp().Total)
	}

	for _, user := range users {
		fmt.Printf("%d--%s\n", user.ID, user.Name)
	}

	// Output:
	// 1--test1
	// 2--test1
}
