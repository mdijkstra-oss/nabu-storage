package file

type CodingMutation string

const (
	SetCoding    CodingMutation = "SetCoding"
	AppendCoding CodingMutation = "AppendCoding"
	RemoveCoding CodingMutation = "RemoveCoding"
)

type CodingAction struct {
	CodeID string         `json:"code_id" validate:"required"`
	Action CodingMutation `json:"action" validate:"oneof=SetCoding AppendCoding RemoveCoding"`
	Texts  []string       `json:"texts" validate:"min=1,dive,required"`
}

type CodeFileData struct {
	Actions []CodingAction `json:"actions" validate:"required,min=1,dive,required"`
}
