package domain

import (
	"nabu-storage/internal/domain/files"
	"nabu-storage/internal/lib/utils"
)

func requireValidPath(cmd *Command) error {
	if !utils.ValidFilePath(cmd.Path) {
		return utils.FieldError("path", "invalid path")
	}
	return nil
}

func handleWriteFile(cmd *Command, projectID, baseDir string) error {
	if err := requireValidPath(cmd); err != nil {
		return err
	}
	return files.Write(baseDir, projectID, cmd.Path, cmd.Content)
}

func handleDeleteFile(cmd *Command, projectID, baseDir string) error {
	if err := requireValidPath(cmd); err != nil {
		return err
	}
	if !files.Exists(baseDir, projectID, cmd.Path) {
		return &utils.NotFoundError{Message: "file not found: " + cmd.Path}
	}
	return files.Delete(baseDir, projectID, cmd.Path)
}

func handleRenameFile(cmd *Command, projectID, baseDir string) error {
	if err := requireValidPath(cmd); err != nil {
		return err
	}
	if !utils.ValidFilePath(cmd.NewPath) {
		return utils.FieldError("newPath", "invalid new path")
	}
	if !files.Exists(baseDir, projectID, cmd.Path) {
		return &utils.NotFoundError{Message: "file not found: " + cmd.Path}
	}
	if files.Exists(baseDir, projectID, cmd.NewPath) {
		return utils.FieldError("newPath", "file already exists")
	}
	return files.Rename(baseDir, projectID, cmd.Path, cmd.NewPath)
}

var actionHandlers = map[Action]func(*Command, string, string) error{
	WriteFile:  handleWriteFile,
	DeleteFile: handleDeleteFile,
	RenameFile: handleRenameFile,
}

func Execute(cmd *Command, projectID, baseDir string) error {
	handler, ok := actionHandlers[cmd.Action]
	if !ok {
		return utils.FieldError("action", "unknown action: "+string(cmd.Action))
	}
	return handler(cmd, projectID, baseDir)
}
