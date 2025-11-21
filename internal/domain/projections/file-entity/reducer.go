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
			projection.For(file.AddedCodeSections, AddedCodeSectionsReducer),
			projection.For(file.UpdatedCodeSections, UpdatedCodeSectionsReducer),
			projection.For(file.RemovedCodeSections, RemovedCodeSectionsReducer),
			projection.For(file.ClearedCoding, ClearedCodingReducer),
			projection.For(code.DeletedCode, DeletedCodeReducer),
			projection.For(code.UpdatedCode, UpdatedCodeReducer),
			projection.For(code.MergedCodes, MergedCodesReducer),
		),
		projection.DeletedProjectReducer[file.File],
	),
)

func CreatedFileReducer(_ *File, message *commands.AnyMessage, payload *file.CreatedFilePayload) *File {
	fileData := payload.FileData
	if fileData.Time.IsZero() {
		fileData.Time = time.Now() // Todo: get from somewhere? eg upload system I suppose
	}
	return &File{
		ID:       message.AggregateID,
		Healthy:  true,
		FileData: fileData,
		Chunks:   payload.Chunks,
	}
}

func UpdatedFileReducer(current *File, _ *commands.AnyMessage, payload *file.UpdatedFilePayload) *File {
	updated := utils.ApplyPartialUpdate(*current, payload)
	return &updated
}

func toCodedSection(section file.AddedSection) file.CodedSection {
	return file.CodedSection(section)
}

func AddedCodeSectionsReducer(current *File, _ *commands.AnyMessage, payload *file.AddedCodeSectionsPayload) *File {
	current.Chunks = utils.Map(current.Chunks, func(chunk file.Chunk) file.Chunk {
		if chunk.ID != payload.ChunkID {
			return chunk
		}

		validSections := utils.Reduce(payload.Sections, []file.CodedSection{}, func(sections []file.CodedSection, section file.AddedSection) []file.CodedSection {
			if find.CountWords(section.Text) < 3 {
				slog.Warn("Text too short, need at least 3 words", "chunk", payload.ChunkID, "section", section)
				return sections
			}

			foundText, found := find.Find(section.Text, chunk.Content)
			if !found {
				slog.Warn("Text not found", "chunk", payload.ChunkID, "section", section)
				return sections
			}

			codedSection := toCodedSection(section)
			codedSection.Text = foundText
			return append(sections, codedSection)
		})

		chunk.Codes = append(chunk.Codes, validSections...)
		return chunk
	})
	return current
}

func UpdatedCodeSectionsReducer(current *File, _ *commands.AnyMessage, payload *file.UpdateCodeSectionsPayload) *File {
	current.Chunks = utils.Map(current.Chunks, func(chunk file.Chunk) file.Chunk {
		if chunk.ID != payload.ChunkID {
			return chunk
		}

		chunk.Codes = utils.Map(chunk.Codes, func(section file.CodedSection) file.CodedSection {
			for _, op := range payload.Sections {
				if section.ID == op.ID {
					if op.Text != "" {
						section.Text = op.Text
					}
					if op.Reason != "" {
						section.Reason = op.Reason
					}
					return section
				}
			}
			return section
		})
		return chunk
	})
	return current
}

func RemovedCodeSectionsReducer(current *File, _ *commands.AnyMessage, payload *file.RemoveCodeSectionsPayload) *File {
	current.Chunks = utils.Map(current.Chunks, func(chunk file.Chunk) file.Chunk {
		if chunk.ID != payload.ChunkID {
			return chunk
		}

		chunk.Codes = utils.Filter(chunk.Codes, func(section file.CodedSection) bool {
			for _, id := range payload.SectionIDs {
				if section.ID == id {
					return false
				}
			}
			return true
		})
		return chunk
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

