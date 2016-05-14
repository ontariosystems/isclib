package log

import (
	"fmt"
	"runtime"
	"strings"
)

// CallInfo describes the stack location where a log message was logged
type CallInfo struct {
	File         string `json:"file"`
	Line         int    `json:"line"`
	FunctionName string `json:"func"`
}

// CallerInfo will get the stack location for a given call stack depth
func CallerInfo(callDepth int) *CallInfo {
	if callDepth <= 0 {
		return &CallInfo{
			File:         "",
			Line:         0,
			FunctionName: fmt.Sprintf("Invalid stack depth of %d", callDepth),
		}
	}

	// Inspect runtime call stack
	pc := make([]uintptr, callDepth)
	numFoundCallers := runtime.Callers(callDepth, pc)

	if numFoundCallers == 0 {
		return &CallInfo{
			File:         "",
			Line:         0,
			FunctionName: fmt.Sprintf("Stack depth %d too deep", callDepth),
		}
	}

	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])

	// Truncate abs file path
	if slash := strings.LastIndex(file, "/"); slash >= 0 {
		file = file[slash+1:]
	}

	// Truncate package name
	funcName := f.Name()
	if slash := strings.LastIndex(funcName, "."); slash >= 0 {
		//funcName = funcName[slash+1:]
	}

	return &CallInfo{
		File:         file,
		Line:         line,
		FunctionName: funcName,
	}
}
