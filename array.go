package jsutil

import "syscall/js"

// ArrayFrom calls Array.from to given argument and returns result Array.
func ArrayFrom(v js.Value) js.Value {
	return arrayClass.Call("from", v)
}
