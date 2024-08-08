package jsutil

import (
	"syscall/js"
	"time"
)

// DateToTime converts JavaScript side's Data object into time.Time.
func DateToTime(v js.Value) (time.Time, error) {
	milli := v.Call("getTime").Float()
	return time.UnixMilli(int64(milli)), nil
}

// TimeToDate converts Go side's time.Time into Date object.
func TimeToDate(t time.Time) js.Value {
	return dateClass.New(t.UnixMilli())
}
