package utils

import "github.com/go-playground/validator/v10"

var Validate = validator.New()

func RegisterFieldValidation[T any](tag string, validation func(T) bool) {
	MustNotError(Validate.RegisterValidation(tag, func(fl validator.FieldLevel) bool {
		value, ok := fl.Field().Interface().(T)
		if !ok {
			return false
		}
		return validation(value)
	}))
}
