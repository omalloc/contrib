package runtime_test

import (
	"strings"
	"testing"

	"github.com/omalloc/contrib/x/runtime"
)

func TestPrintStackTrace(t *testing.T) {
	// 测试基本堆栈跟踪输出
	trace := runtime.PrintStackTrace(1)

	// 验证输出包含当前测试函数名
	if !strings.Contains(trace, "TestPrintStackTrace") {
		t.Errorf("堆栈跟踪中应该包含测试函数名 'TestPrintStackTrace'，但实际输出为:\n%s", trace)
	}

	// 验证输出包含文件名和行号
	if !strings.Contains(trace, "recovered_test.go") {
		t.Errorf("堆栈跟踪中应该包含测试文件名 'recovered_test.go'，但实际输出为:\n%s", trace)
	}

	// 验证输出格式
	lines := strings.Split(strings.TrimSpace(trace), "\n")
	if len(lines) < 2 {
		t.Errorf("堆栈跟踪应至少包含两行，但实际输出为:\n%s", trace)
	}

	// 测试不同的 skip 值
	trace2 := runtime.PrintStackTrace(2)
	if trace == trace2 {
		t.Error("不同的 skip 值应产生不同的堆栈跟踪")
	}
}

// 用于测试 panic 情况的辅助函数
func causePanic() string {
	panic("测试 panic")
}

func TestPrintStackTraceWithPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			trace := runtime.PrintStackTrace(1)

			// 验证输出包含 panic 前缀
			if !strings.Contains(trace, "panic:") {
				t.Errorf("panic 堆栈跟踪应包含 'panic:' 前缀，但实际输出为:\n%s", trace)
			}

			// 验证输出包含触发 panic 的函数名
			if !strings.Contains(trace, "causePanic") {
				t.Errorf("堆栈跟踪应包含 'causePanic' 函数，但实际输出为:\n%s", trace)
			}
		}
	}()

	causePanic()
}
