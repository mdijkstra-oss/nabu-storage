package typedquery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	httphandlers "hermes-relay/internal/handlers/http"
	"hermes-relay/internal/projection"
	"reflect"
	"strconv"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type QueryFunc[T, Q, R any] = func(ctx context.Context, store *projection.Store[T], q Q) (R, error)

func Query[T, Q, R any](
	store *projection.Store[T],
	queryFn QueryFunc[T, Q, R],
) ProcessorFunc {
	return QueryWithMap[T, Q, R, R](store, queryFn, func(r R) R { return r })
}

func QueryWithMap[T, Q, R, Y any](
	store *projection.Store[T],
	queryFn QueryFunc[T, Q, R],
	mapFn func(R) Y,
) ProcessorFunc {
	return func(ctx context.Context, request httphandlers.Request) httphandlers.Response {
		var query Q

		if err := bindPathParams(request.Path, &query); err != nil {
			body, _ := json.Marshal(map[string]string{"message": err.Error()})
			return httphandlers.Response{
				StatusCode: 400,
				Body:       body,
			}
		}

		if err := validate.Struct(query); err != nil {
			body, _ := json.Marshal(map[string]string{"message": err.Error()})
			return httphandlers.Response{
				StatusCode: 400,
				Body:       body,
			}
		}

		result, err := queryFn(ctx, store, query)
		if err != nil {
			body, _ := json.Marshal(map[string]string{"message": err.Error()})
			return httphandlers.Response{
				StatusCode: 404,
				Body:       body,
			}
		}

		mapped := mapFn(result)

		body, err := json.Marshal(mapped)
		if err != nil {
			errBody, _ := json.Marshal(map[string]string{"message": err.Error()})
			return httphandlers.Response{
				StatusCode: 500,
				Body:       errBody,
			}
		}

		return httphandlers.Response{
			StatusCode: 200,
			Body:       body,
		}
	}
}

// Todo: refactor this?
func bindPathParams(pathParams map[string]string, dst any) error {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr {
		return errors.New("dst must be a pointer")
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return errors.New("dst must be pointer to struct")
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		pathTag := field.Tag.Get("path")

		if pathTag == "" {
			continue
		}

		pathValue, ok := pathParams[pathTag]
		if !ok || pathValue == "" {
			continue
		}

		fieldValue := v.Field(i)
		if !fieldValue.CanSet() {
			continue
		}

		switch fieldValue.Kind() {
		case reflect.String:
			fieldValue.SetString(pathValue)
		case reflect.Int, reflect.Int64, reflect.Int32:
			intVal, err := strconv.ParseInt(pathValue, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid int value for %s: %w", pathTag, err)
			}
			fieldValue.SetInt(intVal)
		case reflect.Bool:
			boolVal, err := strconv.ParseBool(pathValue)
			if err != nil {
				return fmt.Errorf("invalid bool value for %s: %w", pathTag, err)
			}
			fieldValue.SetBool(boolVal)
		default:
			return fmt.Errorf("unsupported field type: %s", fieldValue.Kind())
		}
	}

	return nil
}
