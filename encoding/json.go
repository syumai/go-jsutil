package encoding

import (
	"encoding/json"
	"syscall/js"
)

var jsjson = js.Global().Get("JSON")

// MarshalJSONValue encodes a Go value into a JavaScript value.
// The encoding is done via Go's encoding/json.Marshal and JavaScript's JSON.parse.
func MarshalJSONValue(v any) (js.Value, error) {
	value, err := json.Marshal(v)
	if err != nil {
		return js.Value{}, err
	}
	parsed := jsjson.Call("parse", string(value))
	return parsed, nil
}

// UnmarshalJSONValue decodes a JavaScript value into a Go value.
// The decoding is done via JavaScript's JSON.stringify and Go's encoding/json.Unmarshal.
func UnmarshalJSONValue(src js.Value, v any) error {
	stringified := jsjson.Call("stringify", src)
	str := stringified.String()
	return json.Unmarshal([]byte(str), v)
}
