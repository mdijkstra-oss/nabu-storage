package files

import (
	"os"
	"path/filepath"
	"sort"

	"hermes-relay/internal/lib/utils"
)

func ProjectDir(baseDir, projectID string) string {
	return filepath.Join(baseDir, projectID)
}

func FilePath(baseDir, projectID, path string) string {
	return filepath.Join(ProjectDir(baseDir, projectID), path)
}

func EnsureProjectDir(baseDir, projectID string) error {
	return os.MkdirAll(ProjectDir(baseDir, projectID), 0755)
}

func Read(baseDir, projectID, path string) (string, error) {
	data, err := os.ReadFile(FilePath(baseDir, projectID, path))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func Write(baseDir, projectID, path, content string) error {
	if err := EnsureProjectDir(baseDir, projectID); err != nil {
		return err
	}
	return os.WriteFile(FilePath(baseDir, projectID, path), []byte(content), 0644)
}

func Delete(baseDir, projectID, path string) error {
	return os.Remove(FilePath(baseDir, projectID, path))
}

func Rename(baseDir, projectID, oldPath, newPath string) error {
	return os.Rename(
		FilePath(baseDir, projectID, oldPath),
		FilePath(baseDir, projectID, newPath),
	)
}

func Exists(baseDir, projectID, path string) bool {
	_, err := os.Stat(FilePath(baseDir, projectID, path))
	return err == nil
}

func List(baseDir, projectID string) ([]string, error) {
	dir := ProjectDir(baseDir, projectID)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}

	return utils.Filter(
		utils.Map(entries, func(e os.DirEntry) string { return e.Name() }),
		func(name string) bool { return !isHidden(name) },
	), nil
}

func isHidden(name string) bool {
	return len(name) > 0 && name[0] == '.'
}

func SortForInitialSend(names []string) []string {
	sorted := make([]string, len(names))
	copy(sorted, names)
	sort.Strings(sorted)
	return sorted
}
