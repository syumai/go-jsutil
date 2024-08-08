package jsutil

import "syscall/js"

// StrRecordToMap converts JavaScript side's Record<string, string> into map[string]string.
func StrRecordToMap(v js.Value) map[string]string {
	if v.IsUndefined() || v.IsNull() {
		return map[string]string{}
	}
	entries := objectClass.Call("entries", v)
	entriesLen := entries.Get("length").Int()
	result := make(map[string]string, entriesLen)
	for i := 0; i < entriesLen; i++ {
		entry := entries.Index(i)
		key := entry.Index(0).String()
		value := entry.Index(1).String()
		result[key] = value
	}
	return result
}
