package commands

import (
	"context"
	"fmt"
	"reflect"
)

// Todo: validation etc
// Not sure how, pass in existing into func too? Try patch, see if still valid?
func NewUpdateRoute(postfix string, target interface{}, pathsToFilter []string) CommandRouter {
	return func(ctx context.Context, message *Message, publisher PublishFunc) (*Message, error) {
		if (message.Action == "UpdateFile" || message.Action == "PatchFile") && message.Payload != nil {
			return CommandToDomainEvent(message), nil
		}

		return nil, fmt.Errorf("not implemented")
	}
}

// Todo: move somewhere pass in
//
//	type User struct {
//		ID    string `json:"id"`
//		Email string `json:"email" validate:"required,email"`
//		Name  string `json:"name" validate:"required"`
//	}
func AllowedPaths(v interface{}) []string {
	var paths []string
	walkFields(reflect.TypeOf(v), "", &paths)
	return paths
}

func walkFields(t reflect.Type, prefix string, paths *[]string) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Tag.Get("updatable") == "false" {
			continue // skip password etc
		}

		jsonName := field.Tag.Get("json")
		path := prefix + "/" + jsonName
		*paths = append(*paths, path)

		if field.Type.Kind() == reflect.Struct {
			walkFields(field.Type, path, paths)
		}
	}
}
