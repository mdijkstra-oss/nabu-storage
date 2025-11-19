package fileview

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/lib/text-search/find"
	"hermes-relay/internal/lib/utils"
	"log/slog"
	"time"
)

var Reducer = projection.WithHealthCheck(
	projection.CombineReducers(
		projection.For(file.CreatedFile, CreatedFileReducer),
		projection.IfExists(
			projection.For(file.UpdatedFile, UpdatedFileReducer),
			projection.For(file.DeletedFile, projection.DeletedEntity[File]),
			projection.For(file.CodedFile, CodedFileReducer),
			projection.For(file.ClearedCoding, ClearedCodingReducer),
			projection.For(code.DeletedCode, DeletedCodeReducer),
			projection.For(code.UpdatedCode, UpdatedCodeReducer),
			projection.For(code.MergedCodes, MergedCodesReducer),
		),
		projection.DeletedProjectReducer[file.File],
	),
)

func CreatedFileReducer(_ *File, message *commands.AnyMessage, payload *file.CreatedFilePayload) *File {
	return &File{
		BaseFile: file.BaseFile{
			ID:          message.AggregateID,
			ProjectID:   payload.ProjectID,
			Name:        payload.Name,
			Description: payload.Description,
			Healthy:     true,
			Attributes: file.Attributes{
				Title:   "",
				Summary: "",
				Time:    time.Now(), // Todo: get from somewhere? eg upload system I suppose
				Type:    payload.Type,
				Locked:  payload.Locked,
			},
		},
		Chunks: payload.Chunks,
	}
}

func UpdatedFileReducer(current *File, _ *commands.AnyMessage, payload *file.UpdatedFilePayload) *File {
	updated := utils.ApplyPartialUpdate(*current, payload)
	return &updated
}

func CodedFileReducer(current *File, message *commands.AnyMessage, payload *file.CodeFileData) *File {
	current.Chunks = utils.Reduce(payload.Actions, current.Chunks, func(chunks []file.Chunk, action file.CodingAction) []file.Chunk {
		return utils.Map(chunks, func(chunk file.Chunk) file.Chunk {
			if chunk.ID != action.ChunkID {
				return chunk
			}

			switch action.Action {
			case file.RemoveCoding:
				chunk.Codes = utils.Filter(chunk.Codes, func(section file.CodedSection) bool {
					return section.CodeID != action.CodeID
				})

			case file.SetCoding:
				chunk.Codes = utils.Filter(chunk.Codes, func(section file.CodedSection) bool {
					return section.CodeID != action.CodeID
				})
				fallthrough

			case file.AppendCoding:
				newCodes := utils.Reduce(action.Sections, []file.CodedSection{}, func(codes []file.CodedSection, section file.CodedSectionAttributes) []file.CodedSection {
					foundText, found := find.FindRange(section.Text, chunk.Content)
					if !found {
						slog.Warn("Text not found", "chunk", action.ChunkID, "section", section)
						return codes
					}

					return append(codes, file.CodedSection{
						CodeSlug: action.CodeSlug,
						CodeID:   action.CodeID,
						CodedSectionAttributes: file.CodedSectionAttributes{
							Text:   foundText,
							Reason: section.Reason,
						},
					})
				})
				chunk.Codes = append(chunk.Codes, newCodes...)
			}

			return chunk
		})
	})
	return current
}

func ClearedCodingReducer(current *File, message *commands.AnyMessage, payload any) *File {
	for i := range current.Chunks {
		current.Chunks[i].Codes = []file.CodedSection{}
	}
	return current
}

func mapChunkCodes(chunks []file.Chunk, transform func([]file.CodedSection) []file.CodedSection) []file.Chunk {
	return utils.Map(chunks, func(chunk file.Chunk) file.Chunk {
		chunk.Codes = transform(chunk.Codes)
		return chunk
	})
}

func DeletedCodeReducer(current *File, message *commands.AnyMessage, _ code.DeletedCodePayload) *File {
	codeID := message.AggregateID
	current.Chunks = mapChunkCodes(current.Chunks, func(codes []file.CodedSection) []file.CodedSection {
		return utils.Filter(codes, func(cs file.CodedSection) bool {
			return cs.CodeID != codeID
		})
	})
	return current
}

func UpdatedCodeReducer(current *File, message *commands.AnyMessage, payload code.UpdateCodePayload) *File {
	codeID := message.AggregateID
	current.Chunks = mapChunkCodes(current.Chunks, func(codes []file.CodedSection) []file.CodedSection {
		return utils.Map(codes, func(cs file.CodedSection) file.CodedSection {
			if cs.CodeID == codeID {
				cs.CodeSlug = payload.Slug
			}
			return cs
		})
	})
	return current
}

func MergedCodesReducer(current *File, _ *commands.AnyMessage, payload code.MergedCodesPayload) *File {
	current.Chunks = mapChunkCodes(current.Chunks, func(codes []file.CodedSection) []file.CodedSection {
		return utils.Map(codes, func(cs file.CodedSection) file.CodedSection {
			if cs.CodeID == payload.SourceID {
				cs.CodeID = payload.TargetID
			}
			return cs
		})
	})
	return current
}
