package domain

type Action string

const (
	CreateFile Action = "CreateFile"
	UpdateFile Action = "UpdateFile"
	WriteFile  Action = "WriteFile"
	DeleteFile Action = "DeleteFile"
	RenameFile Action = "RenameFile"
	Commit     Action = "Commit"
)

type Command struct {
	Action  Action `json:"action"`
	Path    string `json:"path,omitempty"`
	NewPath string `json:"newPath,omitempty"`
	Diff    string `json:"diff,omitempty"`
	Content string `json:"content,omitempty"`
}
