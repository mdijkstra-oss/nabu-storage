package utils

import (
	"strings"

	"github.com/google/uuid"
)

func CanonicalID(id string) (string, bool) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return "", false
	}
	return parsed.String(), true
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
	switch r {
	case '-', '_', '.', ' ', '(', ')', '\'', ',':
		return true
	}
	return false
}
