package variables

import "testing"

type testFieldInStructData struct {
	f1 string
	f2 bool
	f3 int
	f4 func()
}

type testFieldInStruct struct {
	data   interface{}
	field  string
	result bool
}

var testFieldInStructCases = []testFieldInStruct{
	{testFieldInStructData{}, "f1", true},
	{testFieldInStructData{}, "f2", true},
	{testFieldInStructData{}, "f3", true},
	{testFieldInStructData{}, "f4", true},
	{testFieldInStructData{}, "f5", false},
	{[]string{}, "", false},
	{map[string]int{"oi": 4}, "", false},
	{"asa", "", false},
	{"int", "", false},
}

func TestFieldInStruct(t *testing.T) {
	for _, pair := range testFieldInStructCases {
		v := FieldInStruct(pair.data, pair.field)
		if v != pair.result {
			t.Error(
				"For", pair.data,
				"expected", pair.result,
				"got", v,
			)
		}
	}
}
