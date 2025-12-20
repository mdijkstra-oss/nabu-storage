package utils

import (
	"fmt"

	"github.com/google/uuid"
	"strings"
)

var entityPrefixes = map[string]string{
	"Project":    "project",
	"Document":   "document",
	"Annotation": "annotation",
	"Code":       "code",
}

var blockNamespace = uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")

func NewID() string {
	return uuid.New().String()
}

func DeterministicBlockID(documentID string, index int) string {
	name := fmt.Sprintf("%s:block:%d", documentID, index)
	return "block_" + uuid.NewSHA1(blockNamespace, []byte(name)).String()
}

func NewPrefixedID(prefix string) string {
	return prefix + "_" + uuid.New().String()
}

func NewProjectID() string {
	return NewPrefixedID("project")
}

func NewDocumentID() string {
	return NewPrefixedID("document")
}

func NewAnnotationID() string {
	return NewPrefixedID("annotation")
}

func NewCodeID() string {
	return NewPrefixedID("code")
}

func ValidID(id string) bool {
	if id == "" {
		return false
	}
	_, err := uuid.Parse(id)
	return err == nil
}

func ValidPrefixedID(prefix, id string) bool {
	if id == "" {
		return false
	}
	expectedPrefix := prefix + "_"
	if !strings.HasPrefix(id, expectedPrefix) {
		return false
	}
	uuidPart := strings.TrimPrefix(id, expectedPrefix)
	_, err := uuid.Parse(uuidPart)
	return err == nil
}

func ValidateID(prefix, id string) bool {
	if id == "" {
		return false
	}
	if ValidID(id) {
		return true
	}
	return ValidPrefixedID(prefix, id)
}

func NormalizeID(prefix, id string) string {
	if id == "" {
		return ""
	}
	if ValidID(id) {
		return prefix + "_" + id
	}
	return id
}

func NormalizeAggregateID(aggregateType, id string) string {
	prefix, ok := entityPrefixes[aggregateType]
	if !ok {
		return id
	}
	return NormalizeID(prefix, id)
}

func ValidProjectID(id string) bool {
	return ValidateID("project", id)
}

func ValidDocumentID(id string) bool {
	return ValidateID("document", id)
}

func ValidAnnotationID(id string) bool {
	return ValidateID("annotation", id)
}

func ValidCodeID(id string) bool {
	return ValidateID("code", id)
}
