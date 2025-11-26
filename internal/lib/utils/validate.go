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

func init() {
	registerFieldValidation("code_slug", func(value string) bool {
		pattern := `^[a-z]+(-[a-z]+)*:[a-z]+(-[a-z]+)*$`
		return regexp.MustCompile(pattern).MatchString(value)
	})

	registerFieldValidation("valid_id", func(value string) bool {
		return ValidID(value)
	})

	registerFieldValidation("radix_color", ValidRadixColor)
}
