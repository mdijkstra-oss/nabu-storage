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

type FieldHandler func(field reflect.StructField, fieldValue reflect.Value) (shouldContinue bool, err error)

func IterateFields(dst any, handler FieldHandler) error {
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
		field := t.Field(i)
		fieldValue := v.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			if err := IterateFields(fieldValue.Addr().Interface(), handler); err != nil {
				return err
			}
			continue
		}

		shouldContinue, err := handler(field, fieldValue)
		if err != nil {
			return err
		}
		if !shouldContinue {
			break
		}
	}

	return nil
}

func SetFieldFromString(field reflect.Value, value string, fieldName string) error {
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return SetFieldFromString(field.Elem(), value, fieldName)
	}

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
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value for %s: %w", fieldName, err)
		}
		field.SetFloat(floatVal)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}
	return nil
}

func BindParams(params map[string]string, dst any, tagName string) error {
	return IterateFields(dst, func(field reflect.StructField, fieldValue reflect.Value) (bool, error) {
		tag := field.Tag.Get(tagName)
		if tag == "" {
			return true, nil
		}

		value, ok := params[tag]
		if !ok || value == "" {
			return true, nil
		}

		err := SetFieldFromString(fieldValue, value, tag)
		return true, err
	})
}

func ApplyDefaults(dst any) error {
	return IterateFields(dst, func(field reflect.StructField, fieldValue reflect.Value) (bool, error) {
		if !fieldValue.IsZero() {
			return true, nil
		}

		defaultVal := field.Tag.Get("default")
		if defaultVal == "" {
			return true, nil
		}

		err := SetFieldFromString(fieldValue, defaultVal, field.Name)
		return true, err
	})
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
