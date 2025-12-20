package normalizer

import (
	"hermes-relay/internal/lib/utils"
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

func ProjectID(s string) string {
	return utils.NormalizeID("project", s)
}

func DocumentID(s string) string {
	return utils.NormalizeID("document", s)
}

func AnnotationID(s string) string {
	return utils.NormalizeID("annotation", s)
}

func CodeID(s string) string {
	return utils.NormalizeID("code", s)
}

func init() {
	defaults["trim"] = Trim
	defaults["collapse"] = Collapse
	defaults["kebab"] = Kebab
	defaults["lowercase"] = Lowercase
	defaults["uppercase"] = Uppercase
	defaults["project_id"] = ProjectID
	defaults["document_id"] = DocumentID
	defaults["annotation_id"] = AnnotationID
	defaults["code_id"] = CodeID
}
