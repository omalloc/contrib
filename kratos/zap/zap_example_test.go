package zap_test

import (
	"github.com/omalloc/contrib/v2/kratos/zap"
)

func ExampleWithLevel() {
	// Output:
	// {"level":"info"}

	log := zap.New(zap.WithLevel(zap.InfoLevel))
	log.Debug("no-debug")
	log.Info("info")
}

func ExampleVerbose() {
	// Output:
	// {"level":"info"}
	// {"level":"debug"}

	log := zap.New(zap.Verbose(true))
	log.Info("info")
	log.Debug("debug")
}
