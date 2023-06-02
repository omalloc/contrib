package zap_test

import (
	"fmt"

	"github.com/omalloc/contrib/kratos/zap"
)

func ExampleWithLevel() {
	// Output: debug

	log := zap.New(zap.WithLevel(zap.DebugLevel))

	if ok := log.Core().Enabled(zap.DebugLevel); !ok {
		fmt.Println("no-debug")
		return
	}
	fmt.Println("debug")
}
