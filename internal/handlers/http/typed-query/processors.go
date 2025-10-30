package typedquery

import (
	"encoding/json"
	"errors"
	"fmt"
	httphandlers "hermes-relay/internal/handlers/http"
	"hermes-relay/internal/lib/utils"
	"hermes-relay/internal/projection"
	"reflect"
	"strconv"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func Query[Q, R any](
	exec projection.QueryExecutor[Q, R],
) ProcessorFunc {
	return func(request httphandlers.Request) httphandlers.Response {
		var query Q

		if err := bindPathParams(request.Path, &query); err != nil {
			body, _ := json.Marshal(utils.ValidationError{Message: err.Error()})
			return httphandlers.Response{
				StatusCode: 400,
				Body:       body,
			}
		}

		if err := validate.Struct(query); err != nil {
			body, _ := json.Marshal(utils.ToValidationError(err))
			return httphandlers.Response{
				StatusCode: 400,
				Body:       body,
			}
		}

		result := exec(query)

		// Runtime check for nil pointers due to generic R (which can be *R or R for slices)
		v := reflect.ValueOf(result)
		if v.Kind() == reflect.Ptr && v.IsNil() {
			body, _ := json.Marshal(utils.NotFoundError{Message: "No results found"})
			return httphandlers.Response{
				StatusCode: 404,
				Body:       body,
			}
		}

		body, err := json.Marshal(result)
		if err != nil {
			errBody, _ := json.Marshal(utils.InternalError{Message: err.Error()})
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
