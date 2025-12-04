package fileview

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/lib/utils"
	"time"
)

var Reducer = projection.WithVersionIncrement(
	projection.WithHealthCheck(
		projection.CombineReducers(
			projection.For(file.CreatedFile, CreatedFileReducer),
			projection.IfExists(
				projection.For(file.UpdatedFile, UpdatedFileReducer),
				projection.For(file.PinnedFile, projection.PinnedEntity[File]),
				projection.For(file.UnpinnedFile, projection.UnpinnedEntity[File]),
				projection.For(file.ReplacedFileContent, ReplacedFileContentReducer),
				projection.For(file.DeletedFile, projection.DeletedEntity[File]),
				projection.For(file.AddedCodeSections, AddedCodeSectionsReducer),
				projection.For(file.UpdatedCodeSections, UpdatedCodeSectionsReducer),
				projection.For(file.RemovedCodeSections, RemovedCodeSectionsReducer),
				projection.For(file.ClearedCoding, ClearedCodingReducer),
				projection.For(file.RemovedCodeFromFile, RemovedCodeFromFileReducer),
				projection.For(code.DeletedCode, DeletedCodeReducer),
				projection.For(code.MergedCodes, MergedCodesReducer),
				projection.For(code.ClearedCodeApplications, ClearedCodeApplicationsReducer),
				projection.For(code.RecodedAll, RecodedAllReducer),
			),
			projection.DeletedProjectReducer[file.File],
		),
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

func ReplacedFileContentReducer(current *File, _ *commands.AnyMessage, payload *file.ReplacedFileContentPayload) *File {
	current.Chunks = []file.Chunk{{
		ID:      "1",
		Content: payload.Content,
		Codes:   []file.CodedSection{},
	}}
	return current
}

func withLastActor(section file.CodedSection, actor commands.Actor) file.CodedSection {
	section.LastActor = actor
	return section
}

func toCodedSection(section file.AddedSection, actor commands.Actor) file.CodedSection {
	return withLastActor(file.CodedSection{
		ID:         section.ID,
		CodeID:     section.CodeID,
		Text:       section.Text,
		Reason:     section.Reason,
		Confidence: section.Confidence,
	}, actor)
}

func AddedCodeSectionsReducer(current *File, message *commands.AnyMessage, payload *file.AddedCodeSectionsPayload) *File {
	current.Chunks = utils.Map(current.Chunks, func(chunk file.Chunk) file.Chunk {
		if chunk.ID != payload.ChunkID {
			return chunk
		}

		sections := utils.Map(payload.Sections, func(s file.AddedSection) file.CodedSection {
			return toCodedSection(s, message.Actor)
		})
		chunk.Codes = append(chunk.Codes, sections...)
		return chunk
	})
	return current
}

func UpdatedCodeSectionsReducer(current *File, message *commands.AnyMessage, payload *file.UpdateCodeSectionsPayload) *File {
	current.Chunks = utils.Map(current.Chunks, func(chunk file.Chunk) file.Chunk {
		chunk.Codes = utils.Map(chunk.Codes, func(section file.CodedSection) file.CodedSection {
			for _, op := range payload.Sections {
				if section.ID == op.ID {
					updated := utils.ApplyPartialUpdate(section, op)
					return withLastActor(updated, message.Actor)
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
	current.Chunks = utils.Map(current.Chunks, func(chunk file.Chunk) file.Chunk {
		chunk.Codes = []file.CodedSection{}
		return chunk
	})
	return current
}

func mapChunkCodes(chunks []file.Chunk, transform func([]file.CodedSection) []file.CodedSection) []file.Chunk {
	return utils.Map(chunks, func(chunk file.Chunk) file.Chunk {
		chunk.Codes = transform(chunk.Codes)
		return chunk
	})
}

func DeletedCodeReducer(current *File, message *commands.AnyMessage, _ code.DeletedCodePayload) *File {
	current.Chunks = mapChunkCodes(current.Chunks, filterByCodeID(message.AggregateID))
	return current
}

func MergedCodesReducer(current *File, _ *commands.AnyMessage, payload code.MergedCodesPayload) *File {
	current.Chunks = mapChunkCodes(current.Chunks, remapCodeID(payload.SourceID, payload.TargetID))
	return current
}

func filterByCodeID(codeID string) func([]file.CodedSection) []file.CodedSection {
	return func(codes []file.CodedSection) []file.CodedSection {
		return utils.Filter(codes, func(cs file.CodedSection) bool {
			return cs.CodeID != codeID
		})
	}
}

func remapCodeID(fromID, toID string) func([]file.CodedSection) []file.CodedSection {
	return func(codes []file.CodedSection) []file.CodedSection {
		return utils.Map(codes, func(cs file.CodedSection) file.CodedSection {
			if cs.CodeID == fromID {
				cs.CodeID = toID
			}
			return cs
		})
	}
}

func RemovedCodeFromFileReducer(current *File, _ *commands.AnyMessage, payload file.RemovedCodeFromFilePayload) *File {
	current.Chunks = mapChunkCodes(current.Chunks, filterByCodeID(payload.CodeID))
	return current
}

func ClearedCodeApplicationsReducer(current *File, message *commands.AnyMessage, _ code.ClearedCodeApplicationsPayload) *File {
	current.Chunks = mapChunkCodes(current.Chunks, filterByCodeID(message.AggregateID))
	return current
}

func RecodedAllReducer(current *File, message *commands.AnyMessage, payload code.RecodedAllPayload) *File {
	current.Chunks = mapChunkCodes(current.Chunks, remapCodeID(message.AggregateID, payload.TargetCodeID))
	return current
}

