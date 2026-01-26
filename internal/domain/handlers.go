package domain

import (
	"log/slog"

	"hermes-relay/internal/domain/files"
	"hermes-relay/internal/lib/utils"
)

func requireValidPath(cmd *Command) error {
	if !utils.ValidFilePath(cmd.Path) {
		return utils.FieldError("path", "invalid path")
	}
	return nil
}

func HandleCreateFile(cmd *Command, projectID, baseDir string) error {
	if err := requireValidPath(cmd); err != nil {
		return err
	}
	if files.Exists(baseDir, projectID, cmd.Path) {
		return utils.FieldError("path", "file already exists")
	}
	return files.Create(baseDir, projectID, cmd.Path, cmd.Diff)
}

func HandleUpdateFile(cmd *Command, projectID, baseDir string) error {
	if err := requireValidPath(cmd); err != nil {
		return err
	}
	return files.Update(baseDir, projectID, cmd.Path, cmd.Diff)
}

func HandleDeleteFile(cmd *Command, projectID, baseDir string) error {
	if err := requireValidPath(cmd); err != nil {
		return err
	}
	if !files.Exists(baseDir, projectID, cmd.Path) {
		return &utils.NotFoundError{Message: "file not found: " + cmd.Path}
	}
	return files.Delete(baseDir, projectID, cmd.Path)
}

func HandleRenameFile(cmd *Command, projectID, baseDir string) error {
	if err := requireValidPath(cmd); err != nil {
		return err
	}
	newPath := cmd.Diff
	if !utils.ValidFilePath(newPath) {
		return utils.FieldError("diff", "invalid new path")
	}
	if !files.Exists(baseDir, projectID, cmd.Path) {
		return &utils.NotFoundError{Message: "file not found: " + cmd.Path}
	}
	if files.Exists(baseDir, projectID, newPath) {
		return utils.FieldError("diff", "file already exists")
	}
	return files.Rename(baseDir, projectID, cmd.Path, newPath)
}

func HandleCommit(cmd *Command, projectID, baseDir string) error {
	slog.Info("commit called", "projectId", projectID)
	return nil
}

var Handlers = map[Action]func(*Command, string, string) error{
	CreateFile: HandleCreateFile,
	UpdateFile: HandleUpdateFile,
	DeleteFile: HandleDeleteFile,
	RenameFile: HandleRenameFile,
	Commit:     HandleCommit,
}

func Execute(cmd *Command, projectID, baseDir string) error {
	handler, ok := Handlers[cmd.Action]
	if !ok {
		panic("unknown action: " + string(cmd.Action))
	}
	return handler(cmd, projectID, baseDir)
}
