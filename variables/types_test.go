package variables

import "testing"

type interfaceToBool struct {
	Value  interface{}
	Result bool
}

var testIsMap = []*interfaceToBool{
	{"teste", false},
	{453478, false},
	{-984512, false},
	{true, false},
	{map[string]bool{}, true},
	{map[int]bool{}, true},
	{map[interface{}]bool{}, true},
	{[]string{}, false},
}

func TestIsMap(t *testing.T) {
	for _, test := range testIsMap {
		if IsMap(test.Value) != test.Result {
			t.Errorf("Incorrect value on IsMap for %v; want: %v; got: %v", test.Value, test.Result, !test.Result)
		}
	}
}

var testIsSlice = []*interfaceToBool{
	{"teste", false},
	{453478, false},
	{-984512, false},
	{true, false},
	{map[string]bool{}, false},
	{map[int]bool{}, false},
	{map[interface{}]bool{}, false},
	{[]string{}, true},
	{[]int{}, true},
	{[]bool{}, true},
	{[]interface{}{}, true},
}

func TestIsSlice(t *testing.T) {
	for _, test := range testIsSlice {
		if IsSlice(test.Value) != test.Result {
			t.Errorf("Incorrect value on IsSlice for %v; want: %v; got: %v", test.Value, test.Result, !test.Result)
		}
	}
}
