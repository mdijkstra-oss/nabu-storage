package file

import (
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/utils"
	"slices"
)

var Reducer = cqrs.CombineReducers(
	cqrs.For(CreatedFile, CreatedFileReducer),
	cqrs.For(CodedFile, CodedFileReducer),
	cqrs.For(ClearedCoding, ClearedCodingReducer),
)

func CreatedFileReducer(_ *File, message *cqrs.Message, payload *CreatedFilePayload) *File {
	return &File{
		ID:      payload.ID,
		Content: payload.Content,
		Attributes: Attributes{
			Codes: payload.Codes,
		},
	}
}

func CodedFileReducer(current *File, message *cqrs.Message, payload *CodedFilePayload) *File {
	if current.Codes == nil {
		current.Codes = make(map[string][]string)
	}

	for _, action := range payload.Actions {
		switch action.Action {
		case SetCoding:
			current.Codes[action.CodeID] = action.Texts

		case AppendCoding:
			current.Codes[action.CodeID] = append(current.Codes[action.CodeID], action.Texts...)

		case RemoveCoding:
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
