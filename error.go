package jsutil

import (
	"fmt"
	"syscall/js"
)

func Error(msg string) js.Value {
	return errorClass.New(msg)
}

func Errorf(format string, args ...any) js.Value {
	return errorClass.New(fmt.Sprintf(format, args...))
}
