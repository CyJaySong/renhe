package rlog

import (
	"fmt"
	"runtime"
	"strings"
)

// captureStack 捕获当前调用栈（跳过 skip 层），返回格式化的堆栈字符串。
// 格式：
//
//	Stack:
//	1. main.Test
//	   /Users/me/project/main.go:42
//	2. main.main
//	   /Users/me/project/main.go:15
func captureStack(skip int) string {
	const maxDepth = 32
	pcs := make([]uintptr, maxDepth)
	n := runtime.Callers(skip, pcs)
	if n == 0 {
		return ""
	}

	frames := runtime.CallersFrames(pcs[:n])
	var b strings.Builder
	b.WriteString("Stack:")

	idx := 1
	for {
		frame, more := frames.Next()
		if frame.Function == "" {
			if !more {
				break
			}
			continue
		}
		// 过滤 runtime 内部帧
		if strings.HasPrefix(frame.Function, "runtime.") {
			if !more {
				break
			}
			continue
		}
		b.WriteString(fmt.Sprintf("\n%d. %s\n   %s:%d", idx, frame.Function, frame.File, frame.Line))
		idx++
		if !more {
			break
		}
	}

	if idx == 1 {
		return ""
	}
	return b.String()
}
