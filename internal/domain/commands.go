package domain

type Action string

const (
	CreateFile Action = "CreateFile"
	UpdateFile Action = "UpdateFile"
	DeleteFile Action = "DeleteFile"
	RenameFile Action = "RenameFile"
	Commit     Action = "Commit"
)

type Command struct {
	Action Action `json:"action"`
	Path   string `json:"path,omitempty"`
	Diff   string `json:"diff,omitempty"`
}
