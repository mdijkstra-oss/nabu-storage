package fileview

import (
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/file"
)

func GetByID(files []file.File, id string) *file.File {
	return projection.GetByID(files, id)
}

func Exists(files []file.File, id string) bool {
	return projection.EntityExists(files, id)
}

func FindSection(f file.File, sectionID string) *file.CodedSection {
	for i := range f.Codes {
		if f.Codes[i].ID == sectionID {
			return &f.Codes[i]
		}
	}
	return nil
}
