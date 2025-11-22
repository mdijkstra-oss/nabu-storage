package normalizer

import (
	"errors"
	"fmt"
	"hermes-relay/internal/lib/utils"
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

	return utils.WalkStructFields(elem, normalizeField)
}

func NormalizeValue(s string, funcs ...normFunc) string {
	result := s
	for _, fn := range funcs {
		result = fn(result)
	}
	return result
}

func normalizeField(fieldType reflect.StructField, fieldValue reflect.Value) error {
	if fieldValue.Kind() != reflect.String {
		return nil
	}

	tag := fieldType.Tag.Get("normalize")
	if tag == "" {
		return nil
	}

	value := fieldValue.String()
	if value == "" {
		return nil
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

	fieldValue.SetString(value)
	return nil
}
