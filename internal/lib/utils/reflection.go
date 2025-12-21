package utils

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/jinzhu/copier"
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

var defaultExcludeKeys = []string{"ID"}

func ApplyPartialUpdate[T any](current T, updates any, excludeKeys ...string) T {
	result := current

	updatesValue := reflect.ValueOf(updates)
	if updatesValue.Kind() == reflect.Ptr {
		updatesValue = updatesValue.Elem()
	}

	resultPtr := reflect.ValueOf(&result)
	resultValue := resultPtr.Elem()

	DeepCopyFields(resultValue)

	excluded := toSet(append(defaultExcludeKeys, excludeKeys...))

	updatesType := updatesValue.Type()
	for i := 0; i < updatesValue.NumField(); i++ {
		updateField := updatesValue.Field(i)
		fieldName := updatesType.Field(i).Name

		if excluded[fieldName] {
			continue
		}

		if !updateField.IsZero() {
			resultField := resultValue.FieldByName(fieldName)
			if resultField.IsValid() && resultField.CanSet() {
				resultField.Set(updateField)
			}
		}
	}

	return result
}

func toSet(keys []string) map[string]bool {
	set := make(map[string]bool, len(keys))
	for _, k := range keys {
		set[k] = true
	}
	return set
}

type StructFieldVisitor func(field reflect.StructField, value reflect.Value) error

func WalkReflectValue(v reflect.Value, visitor func(reflect.Value) error) error {
	switch v.Kind() {
	case reflect.Ptr:
		if !v.IsNil() {
			return WalkReflectValue(v.Elem(), visitor)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if field.CanSet() {
				if err := visitor(field); err != nil {
					return err
				}
				if err := WalkReflectValue(field, visitor); err != nil {
					return err
				}
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if err := visitor(v.Index(i)); err != nil {
				return err
			}
			if err := WalkReflectValue(v.Index(i), visitor); err != nil {
				return err
			}
		}
	}
	return nil
}

func WalkStructFields(v reflect.Value, visitor StructFieldVisitor) error {
	switch v.Kind() {
	case reflect.Ptr:
		if !v.IsNil() {
			return WalkStructFields(v.Elem(), visitor)
		}
	case reflect.Struct:
		typ := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if field.CanSet() {
				if err := visitor(typ.Field(i), field); err != nil {
					return err
				}
				if err := WalkStructFields(field, visitor); err != nil {
					return err
				}
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if err := WalkStructFields(v.Index(i), visitor); err != nil {
				return err
			}
		}
	}
	return nil
}

func CopyMatchingFields(src, dst any) error {
	srcValue := reflect.ValueOf(src)
	if srcValue.Kind() == reflect.Ptr {
		srcValue = srcValue.Elem()
	}

	return IterateFields(dst, func(field reflect.StructField, dstField reflect.Value) (bool, error) {
		srcField := srcValue.FieldByName(field.Name)
		if srcField.IsValid() && dstField.CanSet() {
			dstField.Set(srcField)
		}
		return true, nil
	})
}

func DeepCopyFields(v reflect.Value) {
	_ = WalkReflectValue(v, func(field reflect.Value) error {
		switch field.Kind() {
		case reflect.Slice:
			if !field.IsNil() && field.CanSet() {
				newSlice := reflect.MakeSlice(field.Type(), field.Len(), field.Len())
				reflect.Copy(newSlice, field)
				field.Set(newSlice)
			}
		case reflect.Map:
			if !field.IsNil() && field.CanSet() {
				newMap := reflect.MakeMap(field.Type())
				iter := field.MapRange()
				for iter.Next() {
					newMap.SetMapIndex(iter.Key(), iter.Value())
				}
				field.Set(newMap)
			}
		}
		return nil
	})
}

func DeepCopy[T any](src T) T {
	var dst T
	_ = copier.CopyWithOption(&dst, &src, copier.Option{DeepCopy: true})
	return dst
}

func DeepCopyMap[K comparable, V any](m map[K]V) map[K]V {
	newMap := make(map[K]V, len(m))
	for k, v := range m {
		newMap[k] = DeepCopy(v)
	}
	return newMap
}
