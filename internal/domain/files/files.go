package files

import (
	"os"
	"path/filepath"

	"nabu-storage/internal/lib/utils"
)

func ProjectDir(baseDir, projectID string) string {
	return filepath.Join(baseDir, projectID)
}

func ProjectExists(baseDir, projectID string) bool {
	info, err := os.Stat(ProjectDir(baseDir, projectID))
	return err == nil && info.IsDir()
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

// Write replaces a file by renaming a complete one over it. Rename is atomic
// within a directory, so anything reading the project directory while the
// server runs sees either the old file or the new one, never a truncated one,
// and a failed write leaves the previous content in place. The temp file is
// dot-prefixed, which List already skips.
func Write(baseDir, projectID, path, content string) error {
	if err := EnsureProjectDir(baseDir, projectID); err != nil {
		return err
	}

	target := FilePath(baseDir, projectID, path)
	tmp, err := os.CreateTemp(filepath.Dir(target), ".write-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()

	if _, err := tmp.WriteString(content); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	// CreateTemp opens 0600; these files are meant to be readable outside the app.
	if err := os.Chmod(tmp.Name(), 0644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), target)
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
