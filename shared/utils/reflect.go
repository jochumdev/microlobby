package utils

import (
	"reflect"
	"runtime"
	"strings"
)

// ReflectFunctionName Guess Struct.Method from the given Function
func ReflectFunctionName(i interface{}) string {
	switch v := i.(type) {
	case string:
		return v
	default:
		path := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
		// LTrim until the first /
		path = path[strings.LastIndex(path, "/")+1:]

		// Return after the next point
		return path[strings.Index(path, ".")+1:]
	}
}
