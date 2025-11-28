package find

import "strings"

var textReplacements = []struct {
	old string
	new string
}{
	{"\u00A0", " "},
	{"\u2019", ""},
	{"\u2018", ""},
	{"'", ""},
	{"\u201C", ""},
	{"\u201D", ""},
	{"\"", ""},
	{"_", " "},
	{"*", " "},
	{"`", " "},
}

func NormalizeText(s string) string {
	for _, r := range textReplacements {
		s = strings.ReplaceAll(s, r.old, r.new)
	}
	return s
}
