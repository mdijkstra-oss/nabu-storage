package utils

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

func IsNilPtr[T any](v T) bool {
	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Ptr && rv.IsNil()
}

func SetFieldFromString(field reflect.Value, value string, fieldName string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int64, reflect.Int32:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid int value for %s: %w", fieldName, err)
		}
		field.SetInt(intVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid bool value for %s: %w", fieldName, err)
		}
		field.SetBool(boolVal)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}
	return nil
}

func ApplyDefaults(dst any) error {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.New("dst must be non-nil pointer")
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return errors.New("dst must be pointer to struct")
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() || !field.IsZero() {
			continue
		}

		defaultVal := t.Field(i).Tag.Get("default")
		if defaultVal == "" {
			continue
		}

		if err := SetFieldFromString(field, defaultVal, t.Field(i).Name); err != nil {
			return err
		}
	}
	return nil
}

func ApplyDefaultsFromMap(dst any, sourceMap map[string]any) error {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errors.New("dst must be non-nil pointer")
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return errors.New("dst must be pointer to struct")
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}

		jsonTag := t.Field(i).Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		if _, exists := sourceMap[jsonTag]; exists {
			continue
		}

		defaultVal := t.Field(i).Tag.Get("default")
		if defaultVal == "" {
			continue
		}

		if err := SetFieldFromString(field, defaultVal, t.Field(i).Name); err != nil {
			return err
		}
	}
	return nil
}
