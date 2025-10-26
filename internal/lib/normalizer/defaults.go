package normalizer

import (
	"regexp"
	"strings"
)

func Trim(s string) string {
	return strings.TrimSpace(s)
}

func Collapse(s string) string {
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(s, " ")
}

func Kebab(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	re := regexp.MustCompile(`[^a-z0-9]+`)
	return strings.Trim(re.ReplaceAllString(s, "-"), "-")
}

func Lowercase(s string) string {
	return strings.ToLower(s)
}

func Uppercase(s string) string {
	return strings.ToUpper(s)
}

func init() {
	defaults["trim"] = Trim
	defaults["collapse"] = Collapse
	defaults["kebab"] = Kebab
	defaults["lowercase"] = Lowercase
	defaults["uppercase"] = Uppercase
}
