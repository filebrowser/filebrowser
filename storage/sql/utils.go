package sql

import (
	"log"
	"path/filepath"
	"runtime"
	"strings"
)

func getRuntimeFunctionName(frame uint) string {
	pc := make([]uintptr, 1)
	count := runtime.Callers(int(frame)+2, pc)
	if count == 0 {
		return ""
	}
	f := runtime.FuncForPC(pc[0])
	return f.Name()
}

func checkError(err error, message string) bool {
	if err != nil {
		if len(message) > 0 {
			funcname := filepath.Base(getRuntimeFunctionName(1))
			log.Printf("ERROR [%s]: %s\n", funcname, err.Error())
			log.Printf("ERROR [%s]: %s\n", funcname, message)
		}
		return true
	}
	return false
}

func logFunction() {
	funcname := getRuntimeFunctionName(1)
	log.Printf("%s is running\n", funcname)
}

func reverse(list []string) []string {
	var output []string
	for i := len(list) - 1; i >= 0; i-- {
		output = append(output, list[i])
	}
	return output
}

func logBacktrace() {
	funcs := make([]string, 0)
	for _, i := range []int{1, 2, 3} {
		p := filepath.Base(getRuntimeFunctionName(uint(i)))
		if len(p) > 0 {
			funcs = append(funcs, p)
		} else {
			break
		}
	}
	funcs = reverse(funcs)
	log.Printf("%s\n", strings.Join(funcs, " -> "))
}

func LogBacktrace() {
	funcs := make([]string, 0)
	for _, i := range []int{1, 2, 3} {
		p := filepath.Base(getRuntimeFunctionName(uint(i)))
		if len(p) > 0 {
			funcs = append(funcs, p)
		} else {
			break
		}
	}
	funcs = reverse(funcs)
	log.Printf("%s\n", strings.Join(funcs, " -> "))
}
