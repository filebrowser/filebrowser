package frontmatter

import "testing"

type hasRuneTest struct {
	File   []byte
	Return bool
}

var testHasRune = []hasRuneTest{
	hasRuneTest{
		File: []byte(`---
Lorem ipsum dolor sit amet, consectetur adipiscing elit. 
Sed auctor libero eget ante fermentum commodo. 
---`),
		Return: true,
	},
	hasRuneTest{
		File: []byte(`+++
Lorem ipsum dolor sit amet, consectetur adipiscing elit. 
Sed auctor libero eget ante fermentum commodo. 
+++`),
		Return: true,
	},
	hasRuneTest{
		File: []byte(`{
	"json": "Lorem ipsum dolor sit amet"
}`),
		Return: true,
	},
	hasRuneTest{
		File:   []byte(`+`),
		Return: false,
	},
	hasRuneTest{
		File:   []byte(`++`),
		Return: false,
	},
	hasRuneTest{
		File:   []byte(`-`),
		Return: false,
	},
	hasRuneTest{
		File:   []byte(`--`),
		Return: false,
	},
	hasRuneTest{
		File:   []byte(`Lorem ipsum`),
		Return: false,
	},
}

func TestHasRune(t *testing.T) {
	for _, test := range testHasRune {
		if HasRune(test.File) != test.Return {
			t.Error("Incorrect value on HasRune")
		}
	}
}

type appendRuneTest struct {
	Before []byte
	After  []byte
	Mark   rune
}

var testAppendRuneTest = []appendRuneTest{}

func TestAppendRune(t *testing.T) {
	for i, test := range testAppendRuneTest {
		if !compareByte(AppendRune(test.Before, test.Mark), test.After) {
			t.Errorf("Incorrect value on AppendRune of Test %d", i)
		}
	}
}

func compareByte(a, b []byte) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

var testRuneToStringFormat = map[rune]string{
	'-': "yaml",
	'+': "toml",
	'{': "json",
	'}': "json",
	'1': "",
	'a': "",
}

func TestRuneToStringFormat(t *testing.T) {
	for mark, format := range testRuneToStringFormat {
		val, _ := RuneToStringFormat(mark)
		if val != format {
			t.Errorf("Incorrect value on RuneToStringFormat of %v; want: %s; got: %s", mark, format, val)
		}
	}
}

var testStringFormatToRune = map[string]rune{
	"yaml":  '-',
	"toml":  '+',
	"json":  '{',
	"lorem": '0',
}

func TestStringFormatToRune(t *testing.T) {
	for format, mark := range testStringFormatToRune {
		val, _ := StringFormatToRune(format)
		if val != mark {
			t.Errorf("Incorrect value on StringFormatToRune of %s; want: %v; got: %v", format, mark, val)
		}
	}
}
