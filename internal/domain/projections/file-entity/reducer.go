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
				projection.For(file.UpdatedFile, projection.UpdatedEntity[File, file.UpdatedFilePayload]),
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
		fileData.Time = time.Now()
	}
	return &File{
		ID:       message.AggregateID,
		Healthy:  true,
		FileData: fileData,
		Content:  payload.Content,
		Codes:    payload.Codes,
	}
}

func ReplacedFileContentReducer(current *File, _ *commands.AnyMessage, payload *file.ReplacedFileContentPayload) *File {
	current.Content = payload.Content
	current.Codes = []file.CodedSection{}
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
	sections := utils.Map(payload.Sections, func(s file.AddedSection) file.CodedSection {
		return toCodedSection(s, message.Actor)
	})
	current.Codes = append(current.Codes, sections...)
	return current
}

func UpdatedCodeSectionsReducer(current *File, message *commands.AnyMessage, payload *file.UpdateCodeSectionsPayload) *File {
	current.Codes = utils.Map(current.Codes, func(section file.CodedSection) file.CodedSection {
		for _, op := range payload.Sections {
			if section.ID == op.ID {
				updated := utils.ApplyPartialUpdate(section, op)
				return withLastActor(updated, message.Actor)
			}
		}
		return section
	})
	return current
}

func RemovedCodeSectionsReducer(current *File, _ *commands.AnyMessage, payload *file.RemoveCodeSectionsPayload) *File {
	current.Codes = utils.Filter(current.Codes, func(section file.CodedSection) bool {
		for _, id := range payload.SectionIDs {
			if section.ID == id {
				return false
			}
		}
		return true
	})
	return current
}

func ClearedCodingReducer(current *File, message *commands.AnyMessage, payload any) *File {
	current.Codes = []file.CodedSection{}
	return current
}

func DeletedCodeReducer(current *File, message *commands.AnyMessage, _ code.DeletedCodePayload) *File {
	current.Codes = filterByCodeID(message.AggregateID)(current.Codes)
	return current
}

func MergedCodesReducer(current *File, _ *commands.AnyMessage, payload code.MergedCodesPayload) *File {
	current.Codes = remapCodeID(payload.SourceID, payload.TargetID)(current.Codes)
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
	current.Codes = filterByCodeID(payload.CodeID)(current.Codes)
	return current
}

func ClearedCodeApplicationsReducer(current *File, message *commands.AnyMessage, _ code.ClearedCodeApplicationsPayload) *File {
	current.Codes = filterByCodeID(message.AggregateID)(current.Codes)
	return current
}

func RecodedAllReducer(current *File, message *commands.AnyMessage, payload code.RecodedAllPayload) *File {
	current.Codes = remapCodeID(message.AggregateID, payload.TargetCodeID)(current.Codes)
	return current
}

