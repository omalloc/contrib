package orm_test

import (
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/omalloc/contrib/kratos/orm"
)

type User struct {
	gorm.Model
	Name string `json:"name" gorm:"column:name;"`
}

func ExampleNew() {

	db, err := orm.New(
		orm.WithDriver(sqlite.Open("file:test.db?cache=shared&mode=memory")),
		orm.WithTracing(),
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
	_ = db.AutoMigrate(&User{})
	db.Migrator()

	if err := db.Create(&User{Name: "test1"}).Error; err != nil {
		println(err.Error())
	}

	if err := db.Create(&User{Name: "test1"}).Error; err != nil {
		println(err.Error())
	}

	var users []User
	if err := db.Model(&User{}).Find(&users).Error; err != nil {
		println(err.Error())
	}

	for _, user := range users {
		fmt.Printf("%d--%s\n", user.ID, user.Name)
	}

	// Output:
	// 1--test1
	// 2--test1
}
