package file

type CreateFileData struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

type CodingMutation string

const (
	SetCoding    CodingMutation = "SetCoding"
	AppendCoding CodingMutation = "AppendCoding"
	//RemoveCoding CodingMutation = "RemoveCoding"
)

// Todo: Remove by uuid

type CodingAction struct {
	CodeSlug string         `json:"code_slug" validate:"required"`
	Action   CodingMutation `json:"action" validate:"oneof=SetCoding AppendCoding RemoveCoding"`
	Texts    []string       `json:"texts" validate:"min=1,dive,required"`
	ChunkID  string         `json:"chunk_id" validate:"required"`
}

type CodeFileData struct {
	Actions []CodingAction `json:"actions" validate:"required,min=1,dive,required"`
}
