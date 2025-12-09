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
				projection.For(file.ModifiedCodeSections, ModifiedCodeSectionsReducer),
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

func applySectionAddition(codes []file.CodedSection, op file.SectionOp, actor commands.Actor) []file.CodedSection {
	newSection := file.CodedSection{
		ID:         op.ID,
		CodeID:     op.CodeID,
		Text:       op.Text,
		Reason:     op.Reason,
		Confidence: *op.Confidence,
		LastActor:  actor,
	}
	return append(codes, newSection)
}

func applySectionUpdate(codes []file.CodedSection, op file.SectionOp, actor commands.Actor) []file.CodedSection {
	return utils.Map(codes, func(section file.CodedSection) file.CodedSection {
		if section.ID == op.ID {
			if op.CodeID != "" {
				section.CodeID = op.CodeID
			}
			if op.Text != "" {
				section.Text = op.Text
			}
			if op.Reason != "" {
				section.Reason = op.Reason
			}
			if op.Confidence != nil {
				section.Confidence = *op.Confidence
			}
			section.LastActor = actor
		}
		return section
	})
}

func applySectionDeletion(codes []file.CodedSection, op file.SectionOp) []file.CodedSection {
	return utils.Filter(codes, func(section file.CodedSection) bool {
		return section.ID != op.ID
	})
}

func applySectionOperation(codes []file.CodedSection, op file.SectionOp, actor commands.Actor) []file.CodedSection {
	switch op.Op {
	case "add":
		return applySectionAddition(codes, op, actor)
	case "update":
		return applySectionUpdate(codes, op, actor)
	case "delete":
		return applySectionDeletion(codes, op)
	default:
		panic("unknown operation type: " + op.Op)
	}
}

func ModifiedCodeSectionsReducer(current *File, message *commands.AnyMessage, payload *file.ModifiedCodeSectionsPayload) *File {
	codes := current.Codes
	for _, op := range payload.Operations {
		codes = applySectionOperation(codes, op, message.Actor)
	}
	current.Codes = codes
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

