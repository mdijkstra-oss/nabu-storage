package domain

type Action string

const (
	WriteFile  Action = "WriteFile"
	DeleteFile Action = "DeleteFile"
	RenameFile Action = "RenameFile"
	SyncMeta   Action = "SyncMeta"
)

type Command struct {
	Action  Action `json:"action"`
	Path    string `json:"path,omitempty"`
	NewPath string `json:"newPath,omitempty"`
	Content string `json:"content,omitempty"`
}

type SyncMetaFrame struct {
	Action    Action `json:"action"`
	FileCount int    `json:"fileCount"`
}

func NewSyncMetaFrame(fileCount int) SyncMetaFrame {
	return SyncMetaFrame{Action: SyncMeta, FileCount: fileCount}
}
