package fileview

import (
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
)

func FindFileByType(proj project.Project, fileType file.FileType) *file.File {
	for _, f := range proj.Files {
		if f.Type == fileType {
			return &f
		}
	}
	return nil
}
