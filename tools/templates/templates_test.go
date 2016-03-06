package templates

import "testing"

type canBeEdited struct {
	file   string
	result bool
}

var canBeEditedPairs = []canBeEdited{
	{"file.markdown", true},
	{"file.md", true},
	{"file.json", true},
	{"file.toml", true},
	{"file.yaml", true},
	{"file.css", true},
	{"file.sass", true},
	{"file.scss", true},
	{"file.js", true},
	{"file.html", true},
	{"file.git", false},
	{"file.log", false},
	{"file.sh", false},
	{"file.png", false},
	{"file.jpg", false},
}

func TestCanBeEdited(t *testing.T) {
	for _, pair := range canBeEditedPairs {
		v := CanBeEdited(pair.file)
		if v != pair.result {
			t.Error(
				"For", pair.file,
				"expected", pair.result,
				"got", v,
			)
		}
	}
}

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

type testSplitCapitalize struct {
	name   string
	result string
}

var testSplitCapitalizeCases = []testSplitCapitalize{
	{"loremIpsum", "Lorem ipsum"},
	{"LoremIpsum", "Lorem ipsum"},
	{"loremipsum", "Loremipsum"},
	{"YouTube", "YouTube"},
	{"GitHub", "GitHub"},
	{"GooglePlus", "Google Plus"},
	{"Facebook", "Facebook"},
}

func TestSplitCapitalize(t *testing.T) {
	for _, pair := range testSplitCapitalizeCases {
		v := SplitCapitalize(pair.name)
		if v != pair.result {
			t.Error(
				"For", pair.name,
				"expected", pair.result,
				"got", v,
			)
		}
	}
}
