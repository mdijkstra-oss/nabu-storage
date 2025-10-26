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

func init() {
	registerFieldValidation("code_slug", func(value string) bool {
		pattern := `^[a-z]+(-[a-z]+)*:[a-z]+(-[a-z]+)*$`
		return regexp.MustCompile(pattern).MatchString(value)
	})
}
