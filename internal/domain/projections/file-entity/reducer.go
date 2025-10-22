package fileview

import (
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/lib/utils"
	"slices"
)

var Reducer = cqrs.CombineReducers(
	cqrs.For(file.CreatedFile, CreatedFileReducer),
	cqrs.For(file.CodedFile, CodedFileReducer),
	cqrs.For(file.ClearedCoding, ClearedCodingReducer),
)

func CreatedFileReducer(_ *File, message *cqrs.Message, payload *file.CreatedFilePayload) *File {
	return &File{
		ID:      payload.ID,
		Content: payload.Content,
		Attributes: file.Attributes{
			Codes: payload.Codes,
		},
	}
}

func CodedFileReducer(current *File, message *cqrs.Message, payload *file.CodedFilePayload) *File {
	if current.Codes == nil {
		current.Codes = make(map[string][]string)
	}

	for _, action := range payload.Actions {
		switch action.Action {
		case file.SetCoding:
			current.Codes[action.CodeID] = action.Texts

		case file.AppendCoding:
			current.Codes[action.CodeID] = append(current.Codes[action.CodeID], action.Texts...)

		case file.RemoveCoding:
			current.Codes[action.CodeID] = removeTexts(current.Codes[action.CodeID], action.Texts)
			if len(current.Codes[action.CodeID]) == 0 {
				delete(current.Codes, action.CodeID)
			}
		}
	}

	return current
}

func ClearedCodingReducer(current *File, message *cqrs.Message, payload any) *File {
	current.Codes = make(map[string][]string)
	return current
}

func removeTexts(existing, toRemove []string) []string {
	return utils.Filter(existing, func(text string) bool {
		return !slices.Contains(toRemove, text)
	})
}
