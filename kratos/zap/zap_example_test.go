package zap_test

import (
	"github.com/omalloc/contrib/kratos/zap"
)

func ExampleWithLevel() {
	// Output: {"level":"debug"}

	log := zap.New(zap.WithLevel(zap.DebugLevel))
	log.Debug("no-debug")

	log = zap.New(zap.WithLevel(zap.InfoLevel))
	log.Debug("info")
}

func ExampleVerbose() {
	// Output:
	// {"level":"info"}
	// {"level":"debug"}

	log := zap.New(zap.Verbose(true))
	log.Info("info")
	log.Debug("debug")
}
