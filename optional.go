package jsutil

import (
	"syscall/js"
	"time"
)

// MaybeString returns string value of given JavaScript value or returns nil if the value is undefined.
func MaybeString(v js.Value) string {
	if v.IsUndefined() {
		return ""
	}
	return v.String()
}

// MaybeInt returns int value of given JavaScript value or returns nil if the value is undefined.
func MaybeInt(v js.Value) int {
	if v.IsUndefined() {
		return 0
	}
	return v.Int()
}

// MaybeDate returns time.Time value of given JavaScript Date value or returns nil if the value is undefined.
func MaybeDate(v js.Value) (time.Time, error) {
	if v.IsUndefined() {
		return time.Time{}, nil
	}
	return DateToTime(v)
}
