package domain

type Action string

const (
	WriteFile  Action = "WriteFile"
	DeleteFile Action = "DeleteFile"
	RenameFile Action = "RenameFile"
	SyncMeta   Action = "SyncMeta"
)

type Command struct {
	Action    Action `json:"action"`
	Path      string `json:"path,omitempty"`
	NewPath   string `json:"newPath,omitempty"`
	Content   string `json:"content,omitempty"`
	FileCount int    `json:"fileCount,omitempty"`
}
