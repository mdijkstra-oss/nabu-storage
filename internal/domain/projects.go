package domain

import (
	"os"
	"path/filepath"
	"sort"
	"time"

	"nabu-storage/internal/lib/utils"
)

type Project struct {
	ID        string    `json:"id"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ListProjects reads the persistence directory rather than an index, so a project
// directory created by hand is as visible as one this server wrote.
func ListProjects(baseDir string) ([]Project, error) {
	entries, err := os.ReadDir(baseDir)
	if os.IsNotExist(err) {
		return []Project{}, nil
	}
	if err != nil {
		return nil, err
	}

	projects := make([]Project, 0, len(entries))
	for _, entry := range entries {
		id, ok := projectID(entry)
		if !ok {
			continue
		}
		projects = append(projects, Project{
			ID:        id,
			UpdatedAt: lastWrite(filepath.Join(baseDir, entry.Name())),
		})
	}

	sort.Slice(projects, func(a, b int) bool {
		if projects[a].UpdatedAt.Equal(projects[b].UpdatedAt) {
			return projects[a].ID < projects[b].ID
		}
		return projects[a].UpdatedAt.After(projects[b].UpdatedAt)
	})
	return projects, nil
}

// A directory whose name is not a UUID is not addressable over either endpoint, so
// listing it would offer a project nothing can open.
func projectID(entry os.DirEntry) (string, bool) {
	if !entry.IsDir() {
		return "", false
	}
	return utils.CanonicalID(entry.Name())
}

// The directory's own timestamp moves only when a file is added, removed or renamed;
// rewriting one in place leaves it untouched. The newest file is what "last worked on"
// means to someone who has been typing into a document.
func lastWrite(dir string) time.Time {
	newest := modTime(dir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return newest
	}
	for _, entry := range entries {
		if written := modTime(filepath.Join(dir, entry.Name())); written.After(newest) {
			newest = written
		}
	}
	return newest
}

func modTime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}
