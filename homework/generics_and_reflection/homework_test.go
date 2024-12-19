package main

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v homework_test.go

type Person struct {
	Name    string `properties:"name"`
	Address string `properties:"address,omitempty"`
	Age     int    `properties:"age"`
	Married bool   `properties:"married"`
}

func Serialize(person Person) string {
	out, _ := Marshal(person)
	return string(out)
}

const (
	propsTag = "properties"

	propsTokenIgnore = "-"
	propsTokenOmit   = "omitempty"

	propsTokenSep = ","
	propsNameSep  = "."
	propsElemsSep = ","
	propsKVSep    = '='
	propsRowSep   = '\n'
)

var (
	ErrNilPtr         = errors.New("nil pointer")
	ErrNotAStruct     = errors.New("not a struct")
	ErrInvalidToken   = errors.New("invalid token")
	ErrNotImplemented = errors.New("not implemented")
)

func Marshal(in any) (out []byte, err error) {
	return marshal("", in)
}

func marshal(prefix string, in any) (out []byte, err error) {
	v, ok := deref(reflect.ValueOf(in))
	if !ok {
		return nil, ErrNilPtr
	}
	if v.Kind() != reflect.Struct {
		return nil, ErrNotAStruct
	}

	var name string
	for i := 0; ; i++ { // while do
		fldval := v.Field(i)
		fldtyp := reflect.TypeOf(in).Field(i)

		if tag, ok := fldtyp.Tag.Lookup(propsTag); !ok {
			name = strings.ToLower(fldtyp.Name)
		} else {
			tokens := strings.Split(tag, propsTokenSep)
			if len(tokens) == 0 || len(tokens) > 2 {
				return nil, ErrInvalidToken
			}

			if len(tokens) == 2 {
				switch tokens[1] {
				case propsTokenIgnore:
					continue
				case propsTokenOmit:
					if fldval.IsZero() {
						continue
					}
				default:
					return nil, ErrInvalidToken
				}
			}

			if name = tokens[0]; prefix != "" {
				name = fmt.Sprint(prefix, propsNameSep, name)
			}
		}

		if fldval, ok = deref(fldval); !ok {
			continue
		}

		var row []byte
		switch fldval.Kind() {
		case reflect.Struct:
			if row, err = marshal(name, fldval.Interface()); err != nil {
				return nil, err
			}
		case reflect.Array, reflect.Slice:
			elms := make([]string, fldval.Len())
			for j := 0; j < fldval.Len(); j++ {
				elm := fldval.Index(j)
				if elm, ok = deref(elm); !ok {
					continue
				}
				if elm.Kind() == reflect.Struct ||
					elm.Kind() == reflect.Array ||
					elm.Kind() == reflect.Slice ||
					elm.Kind() == reflect.Map ||
					elm.Kind() == reflect.Interface {
					panic(ErrNotImplemented)
				}
				elms[j] = v2s(elm)
			}
			row = mkrow(name, strings.Join(elms, propsElemsSep))
		case reflect.Map, reflect.Interface:
			panic(ErrNotImplemented)
		default: // simple types: string, int, bool, etc...
			row = mkrow(name, v2s(fldval))
		}

		out = append(out, row...)
		if i == v.NumField()-1 {
			break
		}
		out = append(out, propsRowSep)
	}
	return out, nil
}

func deref(v reflect.Value) (reflect.Value, bool) {
	if v.Kind() != reflect.Ptr {
		return v, true
	}

	for ; v.Kind() == reflect.Ptr; v = v.Elem() {
		if v.IsNil() {
			return v, false
		}
	}
	return v, true
}

func v2s(v reflect.Value) string {
	if v.Kind() == reflect.String {
		return v.String()
	}
	return fmt.Sprintf("%v", v.Interface())
}

func mkrow(name, value string) []byte {
	row := make([]byte, 0, len(name)+1+len(value))
	row = append(row, name...)
	row = append(row, propsKVSep)
	row = append(row, value...)
	return row
}

func TestSerialization(t *testing.T) {
	tests := map[string]struct {
		person Person
		result string
	}{
		"test case with empty fields": {
			result: "name=\nage=0\nmarried=false",
		},
		"test case with fields": {
			person: Person{
				Name:    "John Doe",
				Age:     30,
				Married: true,
			},
			result: "name=John Doe\nage=30\nmarried=true",
		},
		"test case with omitempty field": {
			person: Person{
				Name:    "John Doe",
				Age:     30,
				Married: true,
				Address: "Paris",
			},
			result: "name=John Doe\naddress=Paris\nage=30\nmarried=true",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := Serialize(test.person)
			assert.Equal(t, test.result, result)
		})
	}
}

func TestMarshal(t *testing.T) {
	tests := []struct {
		name string
		in   interface{}
		want string
	}{
		{
			name: "simple types",
			in: struct {
				Str  string `properties:"str"`
				Num  int    `properties:"num"`
				Bool bool   `properties:"bool"`
			}{
				Str: "string", Num: 1, Bool: true,
			},
			want: "str=string\nnum=1\nbool=true",
		},
		{
			name: "pointers",
			in: struct {
				Str  *string `properties:"str"`
				Num  *int    `properties:"num"`
				Bool *bool   `properties:"bool"`
			}{
				Str:  ptr("string"),
				Num:  ptr(123),
				Bool: ptr(true),
			},
			want: "str=string\nnum=123\nbool=true",
		},
		{
			name: "array",
			in: struct {
				StrArr []string `properties:"strArr"`
				NumArr []int    `properties:"numArr"`
				Bool   []bool   `properties:"boolArr"`
			}{
				StrArr: []string{"string", "int", "bool"},
				NumArr: []int{123, 456, 789, 123, 456, 789},
				Bool:   []bool{true, false},
			},
			want: "strArr=string,int,bool\nnumArr=123,456,789,123,456,789\nboolArr=true,false",
		},
		{
			name: "slice",
			in: struct {
				StrSlice  []string `properties:"strSlice"`
				NumSlice  []int    `properties:"numSlice"`
				BoolSlice []bool   `properties:"boolSlice"`
			}{
				StrSlice:  []string{"string", "int", "bool"},
				NumSlice:  []int{123, 456, 789, 123, 456, 789},
				BoolSlice: []bool{true, false},
			},
			want: "strSlice=string,int,bool\nnumSlice=123,456,789,123,456,789\nboolSlice=true,false",
		},
		{
			name: "array of pointers",
			in: struct {
				StrArrPtr  []*string `properties:"strArrPtr"`
				NumArrPtr  []*int    `properties:"numArrPtr"`
				BoolArrPtr []*bool   `properties:"boolArrPtr"`
			}{
				StrArrPtr:  []*string{ptr("string"), ptr("int"), ptr("bool")},
				NumArrPtr:  []*int{ptr(123), ptr(456), ptr(789)},
				BoolArrPtr: []*bool{ptr(true), ptr(false)},
			},
			want: "strArrPtr=string,int,bool\nnumArrPtr=123,456,789\nboolArrPtr=true,false",
		},
		{
			name: "embedded struct",
			in: struct {
				Str      string `properties:"str"`
				Embedded struct {
					Str  string `properties:"str"`
					Num  int    `properties:"num"`
					Bool bool   `properties:"bool"`
				} `properties:"embedded"`
				Num  int  `properties:"num"`
				Bool bool `properties:"bool"`
			}{
				Str: "string",
				Embedded: struct {
					Str  string `properties:"str"`
					Num  int    `properties:"num"`
					Bool bool   `properties:"bool"`
				}{
					Str:  "embedded string",
					Num:  123,
					Bool: true,
				},
				Num:  456,
				Bool: false,
			},
			want: "str=string\nembedded.str=embedded string\nembedded.num=123\nembedded.bool=true\nnum=456\nbool=false",
		},
		{
			name: "embedded struct with array",
			in: struct {
				Str      string `properties:"str"`
				Embedded struct {
					Str    string `properties:"str"`
					NumArr []int  `properties:"numArr"`
				}
				Num int `properties:"num"`
			}{
				Str: "lvl1",
				Embedded: struct {
					Str    string `properties:"str"`
					NumArr []int  `properties:"numArr"`
				}{
					Str:    "lvl2",
					NumArr: []int{123, 456, 789, 123, 456, 789},
				},
				Num: 1,
			},
			want: "str=lvl1\nembedded.str=lvl2\nembedded.numArr=123,456,789,123,456,789\nnum=1",
		},
		{
			name: "nested embedded struct",
			in: struct {
				Str      string `properties:"str"`
				Embedded struct {
					Str      string `properties:"str"`
					Embedded struct {
						Str      string `properties:"str"`
						Embedded struct {
							Str string `properties:"str"`
						} `properties:"embedded"`
					} `properties:"embedded"`
				} `properties:"embedded"`
			}{
				Str: "lvl1",
				Embedded: struct {
					Str      string `properties:"str"`
					Embedded struct {
						Str      string `properties:"str"`
						Embedded struct {
							Str string `properties:"str"`
						} `properties:"embedded"`
					} `properties:"embedded"`
				}{
					Str: "lvl2",
					Embedded: struct {
						Str      string `properties:"str"`
						Embedded struct {
							Str string `properties:"str"`
						} `properties:"embedded"`
					}{
						Str: "lvl3",
						Embedded: struct {
							Str string `properties:"str"`
						}{
							Str: "lvl4",
						},
					},
				},
			},
			want: "str=lvl1\nembedded.str=lvl2\nembedded.embedded.str=lvl3\nembedded.embedded.embedded.str=lvl4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := Marshal(tt.in)
			assert.Equal(t, tt.want, string(got))
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}
