package normalizer

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type normFunc func(string) string

var defaults = make(map[string]normFunc)

func Normalize(v any) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return errors.New("must pass pointer to struct")
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return errors.New("must pass pointer to struct")
	}

	return normalizeValue(elem)
}

func NormalizeValue(s string, funcs ...normFunc) string {
	result := s
	for _, fn := range funcs {
		result = fn(result)
	}
	return result
}

func normalizeValue(val reflect.Value) error {
	switch val.Kind() {
	case reflect.Struct:
		return normalizeStruct(val)
	case reflect.Ptr:
		if !val.IsNil() {
			return normalizeValue(val.Elem())
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			if err := normalizeValue(val.Index(i)); err != nil {
				return err
			}
		}
	}
	return nil
}

func normalizeStruct(val reflect.Value) error {
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if !field.CanSet() {
			continue
		}

		tag := fieldType.Tag.Get("normalize")

		if field.Kind() == reflect.String && tag != "" {
			value := field.String()

			if value == "" {
				continue
			}

			normalizers := strings.Split(tag, ",")
			for _, normName := range normalizers {
				normName = strings.TrimSpace(normName)
				fn, ok := defaults[normName]
				if !ok {
					return fmt.Errorf("unknown normalizer: %s", normName)
				}
				value = fn(value)
			}

			field.SetString(value)
		}

		if err := normalizeValue(field); err != nil {
			return err
		}
	}

	return nil
}
