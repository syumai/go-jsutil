package encoding

import (
	"reflect"
	"syscall/js"
	"testing"
)

type Example struct {
	Field1 string `json:"field1"`
	Field2 string `json:"field2"`
}

func TestMarshalJSONValue(t *testing.T) {
	tests := map[string]struct {
		v           any
		wantJSValue js.Value
		isPrimitive bool
	}{
		"string": {
			v:           "hello",
			wantJSValue: js.ValueOf("hello"),
			isPrimitive: true,
		},
		"number": {
			v:           42,
			wantJSValue: js.ValueOf(42),
			isPrimitive: true,
		},
		"bool": {
			v:           true,
			wantJSValue: js.ValueOf(true),
			isPrimitive: true,
		},
		"slice": {
			v:           []any{"a", "b", "c"},
			wantJSValue: js.ValueOf([]any{"a", "b", "c"}),
			isPrimitive: false,
		},
		"map": {
			// map keys can't be sorted, so we use single key Object.
			v:           map[string]any{"a": "b"},
			wantJSValue: js.ValueOf(map[string]any{"a": "b"}),
			isPrimitive: false,
		},
		"struct": {
			v:           Example{Field1: "f1", Field2: "f2"},
			wantJSValue: js.ValueOf(map[string]any{"field1": "f1", "field2": "f2"}),
			isPrimitive: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			gotJSValue, err := MarshalJSONValue(tc.v)
			if err != nil {
				t.Fatalf("want err: nil, got: %v", err)
			}
			if tc.isPrimitive {
				if !tc.wantJSValue.Equal(gotJSValue) {
					t.Fatalf("want equal: true, got: false")
				}
				return
			}
			// We can't compare the two js.Value directly, so we compare their string representation.
			wantJSStr := jsjson.Call("stringify", tc.wantJSValue).String()
			gotJSStr := jsjson.Call("stringify", gotJSValue).String()
			if wantJSStr != gotJSStr {
				t.Fatalf("want: %s, got: %s", wantJSStr, gotJSStr)
			}
		})
	}
}

func TestUnmarshalJSONValue(t *testing.T) {
	tests := map[string]struct {
		src  js.Value
		want any
	}{
		"string": {
			src:  js.ValueOf("hello"),
			want: "hello",
		},
		"number": {
			src: js.ValueOf(42),
			// JavaScript's number is converted into float64 implicitly.
			want: float64(42),
		},
		"bool": {
			src:  js.ValueOf(true),
			want: true,
		},
		"slice": {
			src:  js.ValueOf([]any{"a", "b", "c"}),
			want: []any{"a", "b", "c"},
		},
		"map": {
			// map keys can't be sorted, so we use single key Object.
			src:  js.ValueOf(map[string]any{"a": "b"}),
			want: map[string]any{"a": "b"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var got any
			if err := UnmarshalJSONValue(tc.src, &got); err != nil {
				t.Fatalf("want err: nil, got: %v", err)
			}
			if !reflect.DeepEqual(tc.want, got) {
				t.Fatalf("want: %v, got: %v", tc.want, got)
			}
		})
	}

	t.Run("struct", func(t *testing.T) {
		src := js.ValueOf(map[string]any{"field1": "f1", "field2": "f2"})
		var got Example
		if err := UnmarshalJSONValue(src, &got); err != nil {
			t.Fatalf("want err: nil, got: %v", err)
		}
		want := Example{Field1: "f1", Field2: "f2"}
		if want != got {
			t.Fatalf("want: %v, got: %v", want, got)
		}
	})
}
