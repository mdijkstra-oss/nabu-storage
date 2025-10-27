package file

type CreateFileData struct {
	BaseFile
	Content string `json:"content"`
}

type CodingMutation string

const (
	SetCoding    CodingMutation = "SetCoding"
	AppendCoding CodingMutation = "AppendCoding"
	RemoveCoding CodingMutation = "RemoveCoding"
)

type CodingAction struct {
	CodeSlug string                   `json:"code_slug" validate:"required,code_slug"`
	Action   CodingMutation           `json:"action" validate:"oneof=SetCoding AppendCoding RemoveCoding"`
	Sections []CodedSectionAttributes `json:"texts" validate:"min=1,dive,required"`
	ChunkID  string                   `json:"chunk_id" validate:"required"`
}

type CodeFileData struct {
	Actions []CodingAction `json:"actions" validate:"required,min=1,dive,required"`
}

type MergeCodesData struct {
	Source string `json:"source" validate:"required,code_slug" normalize:"trim,lowercase"`
	Target string `json:"target" validate:"required,code_slug" normalize:"trim,lowercase"`
}
