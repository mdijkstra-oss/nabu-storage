package utils

import (
	"strings"

	"github.com/google/uuid"
)

func ValidID(id string) bool {
	if id == "" {
		return false
	}
	_, err := uuid.Parse(id)
	return err == nil
}

func ValidFilePath(path string) bool {
	if path == "" {
		return false
	}
	if strings.HasPrefix(path, ".") {
		return false
	}
	if strings.Contains(path, "..") {
		return false
	}
	for _, r := range path {
		if !isSafeFilenameChar(r) {
			return false
		}
	}
	return true
}

func isSafeFilenameChar(r rune) bool {
	if r >= 'a' && r <= 'z' {
		return true
	}
	if r >= 'A' && r <= 'Z' {
		return true
	}
	if r >= '0' && r <= '9' {
		return true
	}
	return r == '-' || r == '_' || r == '.'
}
