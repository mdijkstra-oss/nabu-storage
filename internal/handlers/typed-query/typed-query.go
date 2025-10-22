package typedquery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hermes-relay/internal/lib/utils"
	"hermes-relay/internal/projection"
	"net/http"
	"reflect"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type QueryFunc[T, Q, R any] = func(ctx context.Context, store *projection.Store[T], q Q) (R, error)

func Route[T, Q, R any](
	store *projection.Store[T],
	queryFn QueryFunc[T, Q, R],
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var query Q

		if err := bindPathParams(r, &query); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		if err := validate.Struct(query); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		result, err := queryFn(r.Context(), store, query)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		utils.MustNotError(json.NewEncoder(w).Encode(result))
	}
}

func bindPathParams(r *http.Request, dst any) error {
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

		pathValue := chi.URLParam(r, pathTag)
		if pathValue == "" {
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
