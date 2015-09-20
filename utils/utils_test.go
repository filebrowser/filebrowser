package utils

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
