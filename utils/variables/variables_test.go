package variables

import "testing"

type testDefinedData struct {
	f1 string
	f2 bool
	f3 int
	f4 func()
}

type testDefined struct {
	data   interface{}
	field  string
	result bool
}

var testDefinedCases = []testDefined{
	{testDefinedData{}, "f1", true},
	{testDefinedData{}, "f2", true},
	{testDefinedData{}, "f3", true},
	{testDefinedData{}, "f4", true},
	{testDefinedData{}, "f5", false},
	{[]string{}, "", false},
	{map[string]int{"oi": 4}, "", false},
	{"asa", "", false},
	{"int", "", false},
}

func TestDefined(t *testing.T) {
	for _, pair := range testDefinedCases {
		v := Defined(pair.data, pair.field)
		if v != pair.result {
			t.Error(
				"For", pair.data,
				"expected", pair.result,
				"got", v,
			)
		}
	}
}
