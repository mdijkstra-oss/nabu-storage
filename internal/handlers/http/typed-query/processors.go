package typedquery

import (
	"encoding/json"
	"errors"
	"hermes-relay/internal/cqrs/projection"
	httphandlers "hermes-relay/internal/handlers/http"
	"hermes-relay/internal/lib/utils"
	"reflect"

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

		if err := bindQueryParams(request.Query, &query); err != nil {
			body, _ := json.Marshal(utils.ValidationError{Message: err.Error()})
			return httphandlers.Response{
				StatusCode: 400,
				Body:       body,
			}
		}

		if err := utils.ApplyDefaults(&query); err != nil {
			body, _ := json.Marshal(utils.InternalError{Message: err.Error()})
			return httphandlers.Response{
				StatusCode: 500,
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

		if utils.IsNilPtr(result) {
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
	return bindParams(pathParams, dst, "path")
}

func bindQueryParams(queryParams map[string]string, dst any) error {
	return bindParams(queryParams, dst, "query")
}

func bindParams(params map[string]string, dst any, tagName string) error {
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
		tag := field.Tag.Get(tagName)

		if tag == "" {
			continue
		}

		value, ok := params[tag]
		if !ok || value == "" {
			continue
		}

		fieldValue := v.Field(i)
		if !fieldValue.CanSet() {
			continue
		}

		if err := utils.SetFieldFromString(fieldValue, value, tag); err != nil {
			return err
		}
	}

	return nil
}
