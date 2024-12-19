package runtime

import (
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
)

const MaxMemory = 512 << 20 // 512MB

func AutoGOMAXPROCS() (int, int) {
	// set the max proc = preCore / 2
	maxThreads := 1000
	if envStr, ok := os.LookupEnv("APP_MAXTHREADS"); ok {
		if v, err := strconv.Atoi(envStr); err == nil {
			maxThreads = v
		}
	}

	debug.SetMaxThreads(maxThreads)
	debug.SetMemoryLimit(MaxMemory)
	return runtime.GOMAXPROCS(runtime.NumCPU()), maxThreads
}
