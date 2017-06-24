package variables

import "testing"

type interfaceToBool struct {
	Value  interface{}
	Result bool
}

var testIsMap = []*interfaceToBool{
	&interfaceToBool{"teste", false},
	&interfaceToBool{453478, false},
	&interfaceToBool{-984512, false},
	&interfaceToBool{true, false},
	&interfaceToBool{map[string]bool{}, true},
	&interfaceToBool{map[int]bool{}, true},
	&interfaceToBool{map[interface{}]bool{}, true},
	&interfaceToBool{[]string{}, false},
}

func TestIsMap(t *testing.T) {
	for _, test := range testIsMap {
		if IsMap(test.Value) != test.Result {
			t.Errorf("Incorrect value on IsMap for %v; want: %v; got: %v", test.Value, test.Result, !test.Result)
		}
	}
}

var testIsSlice = []*interfaceToBool{
	&interfaceToBool{"teste", false},
	&interfaceToBool{453478, false},
	&interfaceToBool{-984512, false},
	&interfaceToBool{true, false},
	&interfaceToBool{map[string]bool{}, false},
	&interfaceToBool{map[int]bool{}, false},
	&interfaceToBool{map[interface{}]bool{}, false},
	&interfaceToBool{[]string{}, true},
	&interfaceToBool{[]int{}, true},
	&interfaceToBool{[]bool{}, true},
	&interfaceToBool{[]interface{}{}, true},
}

func TestIsSlice(t *testing.T) {
	for _, test := range testIsSlice {
		if IsSlice(test.Value) != test.Result {
			t.Errorf("Incorrect value on IsSlice for %v; want: %v; got: %v", test.Value, test.Result, !test.Result)
		}
	}
}
