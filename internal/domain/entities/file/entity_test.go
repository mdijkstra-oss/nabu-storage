package file

import (
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
)

func TestFileType_IsLocked(t *testing.T) {
	th.RunMapTests(t, map[FileType]bool{
		FileTypeCorpus:   true,
		FileTypeCodebook: false,
		FileTypeMemo:     false,
		FileTypeLLMMemo:  false,
	}, FileType.IsLocked)
}

func TestFileType_IsSingleton(t *testing.T) {
	th.RunMapTests(t, map[FileType]bool{
		FileTypeCorpus:   false,
		FileTypeCodebook: true,
		FileTypeMemo:     false,
		FileTypeLLMMemo:  true,
	}, FileType.IsSingleton)
}
