package document

import "hermes-relay/internal/cqrs/commands"

const (
	CreatedDocument  = "CreatedDocument"
	UpdatedDocument  = "UpdatedDocument"
	PinnedDocument   = "PinnedDocument"
	UnpinnedDocument = "UnpinnedDocument"
	DeletedDocument  = "DeletedDocument"

	InsertedBlocks  = "InsertedBlocks"
	DeletedBlocks   = "DeletedBlocks"
	MovedBlocks     = "MovedBlocks"
	ReplacedContent = "ReplacedContent"
	UpdatedBlock    = "UpdatedBlock"

	AddedDocumentTags   = "AddedDocumentTags"
	RemovedDocumentTags = "RemovedDocumentTags"

	AddedAnnotation        = "AddedAnnotation"
	RemovedAnnotations     = "RemovedAnnotations"
	UpdatedAnnotationProps = "UpdatedAnnotationProps"
)

type CreatedDocumentPayload struct {
	ProjectID   string `json:"project_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdatedDocumentPayload = UpdateDocumentPayload
type PinnedDocumentPayload = commands.EmptyPayload
type UnpinnedDocumentPayload = commands.EmptyPayload
type DeletedDocumentPayload = commands.EmptyPayload

type InsertedBlocksPayload = InsertBlocksPayload
type DeletedBlocksPayload = DeleteBlocksPayload
type MovedBlocksPayload = MoveBlocksPayload
type ReplacedContentPayload = ReplaceContentPayload
type UpdatedBlockPayload = UpdateBlockPayload

type AddedDocumentTagsPayload = AddDocumentTagsPayload
type RemovedDocumentTagsPayload = RemoveDocumentTagsPayload

type AddedAnnotationPayload = AddAnnotationPayload
type RemovedAnnotationsPayload = RemoveAnnotationsPayload
type UpdatedAnnotationPropsPayload = UpdateAnnotationPropsPayload
