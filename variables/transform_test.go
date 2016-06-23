package variables

import "testing"

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
