package runtime

import (
	"fmt"
	"runtime"
	"strings"
)

func PrintStackTrace(skip int) string {
	// Capture the stack trace
	pc := make([]uintptr, 10)
	n := runtime.Callers(skip, pc)
	frames := runtime.CallersFrames(pc[:n])

	// Iterate over the frames and print them
	sb := strings.Builder{}
	for {
		frame, more := frames.Next()
		if strings.HasPrefix(frame.Function, "runtime.panic") || strings.HasPrefix(frame.Function, "runtime.gopanic") {
			sb.WriteString("panic: ")
			continue
		}
		sb.WriteString(fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}

	return sb.String()
}
