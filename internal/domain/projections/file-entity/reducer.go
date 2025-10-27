package fileview

import (
	"github.com/google/uuid"
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/lib/text-search/chunker"
	"hermes-relay/internal/lib/text-search/find"
	"hermes-relay/internal/lib/utils"
	"log/slog"
)

var Reducer = cqrs.CombineReducers(
	cqrs.For(file.CreatedFile, CreatedFileReducer),
	cqrs.For(file.CodedFile, CodedFileReducer),
	cqrs.For(file.ClearedCoding, ClearedCodingReducer),
)

func CreatedFileReducer(_ *File, message *cqrs.AnyMessage, payload *file.CreatedFilePayload) *File {
	// Todo: What's faster, what's better?
	blocks := chunker.ChunkBlocks(payload.Content, chunker.FullPage, chunker.FullPage+chunker.HalfPage)

	chunks := utils.Map(blocks, func(block string) file.Chunk {
		return file.Chunk{
			ID:      uuid.New().String(),
			Content: block,
			Codes:   []file.CodedSection{},
		}
	})

	return &File{
		BaseFile: file.BaseFile{
			ID:   message.AggregateID,
			Name: payload.Name,
			Attributes: file.Attributes{
				Title:   payload.Title,
				Summary: payload.Summary,
				Time:    payload.Time,
			},
		},
		Content: payload.Content,
		Chunks:  chunks,
	}
}

func CodedFileReducer(current *File, message *cqrs.AnyMessage, payload *file.CodeFileData) *File {
	for _, action := range payload.Actions {
		chunkIdx := -1
		for i, c := range current.Chunks {
			if c.ID == action.ChunkID {
				chunkIdx = i
				break
			}
		}

		if chunkIdx == -1 {
			slog.Warn("Chunk not found", "id", action.ChunkID)
			continue
		}

		chunk := &current.Chunks[chunkIdx]

		switch action.Action {
		case file.RemoveCoding:
			chunk.Codes = utils.Filter(chunk.Codes, func(code file.CodedSection) bool {
				return code.CodeSlug != action.CodeSlug
			})

		case file.SetCoding:
			chunk.Codes = utils.Filter(chunk.Codes, func(code file.CodedSection) bool {
				return code.CodeSlug != action.CodeSlug
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

func ClearedCodingReducer(current *File, message *cqrs.AnyMessage, payload any) *File {
	for i := range current.Chunks {
		current.Chunks[i].Codes = []file.CodedSection{}
	}
	return current
}
