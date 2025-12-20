package utils

import (
	"github.com/go-playground/validator/v10"
	"regexp"
)

var Validate = validator.New()

func registerFieldValidation[T any](tag string, validation func(T) bool) {
	MustNotError(Validate.RegisterValidation(tag, func(fl validator.FieldLevel) bool {
		value, ok := fl.Field().Interface().(T)
		if !ok {
			return false
		}
		return validation(value)
	}))
}

var radixColors = map[string]bool{
	"gray": true, "mauve": true, "slate": true, "sage": true, "olive": true, "sand": true,
	"tomato": true, "red": true, "ruby": true, "crimson": true, "pink": true, "plum": true,
	"purple": true, "violet": true, "iris": true, "indigo": true, "blue": true, "cyan": true,
	"teal": true, "jade": true, "green": true, "grass": true, "bronze": true, "gold": true,
	"brown": true, "orange": true, "amber": true, "yellow": true, "lime": true, "mint": true,
	"sky": true,
}

func ValidRadixColor(color string) bool {
	return radixColors[color]
}

func ValidCodeSlug(slug string) bool {
	pattern := `^[a-z]+(-[a-z]+)*:[a-z]+(-[a-z]+)*$`
	return regexp.MustCompile(pattern).MatchString(slug)
}

func init() {
	registerFieldValidation("code_slug", ValidCodeSlug)

	registerFieldValidation("valid_id", func(value string) bool {
		return ValidID(value)
	})

	registerFieldValidation("valid_id_or_slug", func(value string) bool {
		return ValidID(value) || ValidCodeSlug(value)
	})

	registerFieldValidation("radix_color", ValidRadixColor)

	registerFieldValidation("project_id", ValidProjectID)
	registerFieldValidation("document_id", ValidDocumentID)
	registerFieldValidation("annotation_id", ValidAnnotationID)
	registerFieldValidation("code_id", ValidCodeID)
}
