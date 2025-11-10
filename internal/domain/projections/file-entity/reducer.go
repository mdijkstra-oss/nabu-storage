package fileview

import (
	"fmt"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/lib/text-search/chunker"
	"hermes-relay/internal/lib/text-search/find"
	"hermes-relay/internal/lib/utils"
	"log/slog"
	"time"
)

var Reducer = projection.CombineReducers(
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
)

func CreatedFileReducer(_ *File, message *commands.AnyMessage, payload *file.CreatedFilePayload) *File {
	// Todo: What's faster, what's better?
	blocks := chunker.ChunkBlocks(payload.Content, chunker.FullPage, chunker.FullPage+chunker.HalfPage)

	chunks := utils.MapWithIndex(blocks, func(i int, block string) file.Chunk {
		return file.Chunk{
			ID:      fmt.Sprintf("%d", i+1),
			Content: block,
			Codes:   []file.CodedSection{},
		}
	})

	return &File{
		BaseFile: file.BaseFile{
			ID:          message.AggregateID,
			ProjectID:   payload.ProjectID,
			Name:        payload.Name,
			Description: payload.Description,
			Attributes: file.Attributes{
				Title:   "",
				Summary: "",
				Time:    time.Now(), // Todo: get from somewhere? eg upload system I suppose
				Type:    payload.Type,
				Locked:  payload.Locked,
			},
		},
		Chunks: chunks,
	}
}

func UpdatedFileReducer(current *File, _ *commands.AnyMessage, payload *file.UpdatedFilePayload) *File {
	current.Name = payload.Name
	current.Description = payload.Description
	return current
}

func CodedFileReducer(current *File, message *commands.AnyMessage, payload *file.CodeFileData) *File {
	for _, action := range payload.Actions {
		chunkID := utils.FindIndex(current.Chunks, func(c file.Chunk) bool {
			return c.ID == action.ChunkID
		})

		if chunkID == -1 {
			slog.Warn("Chunk not found", "id", action.ChunkID)
			continue
		}

		chunk := &current.Chunks[chunkID]

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
			for _, section := range action.Sections {
				start, end, found := find.FindRange(section.Text, chunk.Content)
				if !found {
					slog.Warn("Text not found", "chunk", action.ChunkID, "section", section)
					continue
				}

				chunk.Codes = append(chunk.Codes, file.CodedSection{
					StartIndex: start,
					EndIndex:   end,
					CodeSlug:   action.CodeSlug,
					CodeID:     action.CodeID,
					CodedSectionAttributes: file.CodedSectionAttributes{
						Text:     section.Text,
						AIReason: section.AIReason,
						Comment:  section.Comment,
					},
				})
			}
		}
	}
	return current
}

func ClearedCodingReducer(current *File, message *commands.AnyMessage, payload any) *File {
	for i := range current.Chunks {
		current.Chunks[i].Codes = []file.CodedSection{}
	}
	return current
}

func DeletedCodeReducer(current *File, message *commands.AnyMessage, _ code.DeletedCodePayload) *File {
	codeID := message.AggregateID

	current.Chunks = utils.Map(current.Chunks, func(chunk file.Chunk) file.Chunk {
		chunk.Codes = utils.Filter(chunk.Codes, func(cs file.CodedSection) bool {
			return cs.CodeID != codeID
		})
		return chunk
	})

	return current
}

func UpdatedCodeReducer(current *File, message *commands.AnyMessage, payload code.UpdateCodePayload) *File {
	codeID := message.AggregateID

	current.Chunks = utils.Map(current.Chunks, func(chunk file.Chunk) file.Chunk {
		chunk.Codes = utils.Map(chunk.Codes, func(cs file.CodedSection) file.CodedSection {
			if cs.CodeID == codeID {
				cs.CodeSlug = payload.Slug
			}
			return cs
		})
		return chunk
	})

	return current
}

func MergedCodesReducer(current *File, _ *commands.AnyMessage, payload code.MergedCodesPayload) *File {
	current.Chunks = utils.Map(current.Chunks, func(chunk file.Chunk) file.Chunk {
		chunk.Codes = utils.Map(chunk.Codes, func(cs file.CodedSection) file.CodedSection {
			if cs.CodeID == payload.SourceID {
				cs.CodeID = payload.TargetID
			}
			return cs
		})
		return chunk
	})

	return current
}
