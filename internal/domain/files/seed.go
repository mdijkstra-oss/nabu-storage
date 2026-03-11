package files

import (
	"embed"
	"log/slog"
	"path/filepath"
)

//go:embed templates/*.md
var templates embed.FS

type requiredFile struct {
	path     string
	template string
}

var requiredFiles = func() []requiredFile {
	entries, err := templates.ReadDir("templates")
	if err != nil {
		panic("failed to read embedded templates: " + err.Error())
	}

	result := make([]requiredFile, 0, len(entries))
	for _, e := range entries {
		content, err := templates.ReadFile(filepath.Join("templates", e.Name()))
		if err != nil {
			panic("failed to read embedded template " + e.Name() + ": " + err.Error())
		}
		result = append(result, requiredFile{path: e.Name(), template: string(content)})
	}
	return result
}()

func SeedRequiredFiles(baseDir, projectID string) {
	for _, rf := range requiredFiles {
		if Exists(baseDir, projectID, rf.path) {
			continue
		}
		slog.Info("seeding required file", "projectID", projectID, "path", rf.path)
		if err := Write(baseDir, projectID, rf.path, rf.template); err != nil {
			slog.Error("failed to seed file", "projectID", projectID, "path", rf.path, "error", err)
		}
	}
}
