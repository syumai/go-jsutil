package jsutil

import (
	"syscall/js"
)

var (
	objectClass         = js.Global().Get("Object")
	promiseClass        = js.Global().Get("Promise")
	arrayClass          = js.Global().Get("Array")
	uint8ArrayClass     = js.Global().Get("Uint8Array")
	errorClass          = js.Global().Get("Error")
	readableStreamClass = js.Global().Get("ReadableStream")
	dateClass           = js.Global().Get("Date")
	null                = js.ValueOf(nil)
)

func NewObject() js.Value {
	return objectClass.New()
}

func NewUint8Array(size int) js.Value {
	return uint8ArrayClass.New(size)
}

func NewPromise(fn js.Func) js.Value {
	return promiseClass.New(fn)
}
