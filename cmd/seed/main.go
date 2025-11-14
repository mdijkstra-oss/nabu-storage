package main

import (
	"flag"
	"hermes-relay/internal/bootstrap"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/persistence"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/lib/utils"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

func main() {
	var reset bool
	var sourceDir string

	flag.BoolVar(&reset, "reset", false, "Reset all data and re-import all files")
	flag.StringVar(&sourceDir, "source", "", "Source directory containing .md files to import")
	flag.Parse()

	bootstrap.SetupLogger(slog.LevelDebug)

	if sourceDir == "" {
		homeDir, _ := os.UserHomeDir()
		sourceDir = filepath.Join(homeDir, "Documents/hermes-source-files")
	}

	publisher := setupPublisher(reset)

	if err := seedFiles(publisher, sourceDir, reset); err != nil {
		slog.Error("failed to seed files", "error", err)
		os.Exit(1)
	}

	slog.Info("seeding complete")
}

func setupPublisher(reset bool) *dispatch.InMemoryPublisher {
	publisher := dispatch.NewInMemoryPublisher()
	registry := bootstrap.SetupRegistry(publisher)

	disk := persistence.New()

	if reset {
		slog.Info("reset flag set, clearing existing data")
		utils.MustNotError(clearPersistenceData())
	} else {
		utils.MustNotError(disk.ReplayAllEvents(publisher))
	}

	bootstrap.SetupCommandHandlers(publisher, registry)
	publisher.Subscribe(dispatch.LimitOnType(commands.DomainEvent, dispatch.ReadOnlyRoutes(disk.Apply())))

	return publisher
}

func clearPersistenceData() error {
	home, _ := os.UserHomeDir()
	basePath := filepath.Join(home, "Documents", "hermes-persistence")
	return os.RemoveAll(basePath)
}

func getFileDescription(content string) string {
	maxLen := 500
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen]
}

func seedFiles(publisher *dispatch.InMemoryPublisher, sourceDir string, reset bool) error {
	projectID := getOrCreateProjectID(publisher.Publish)

	existingFiles := make(map[string]bool)
	if !reset {
		existingFiles = getExistingFileNames()
	}

	files, err := filepath.Glob(filepath.Join(sourceDir, "*.md"))
	if err != nil {
		return err
	}

	imported := 0
	for _, path := range files {
		filename := filepath.Base(path)
		if !reset && existingFiles[filename] {
			continue
		}

		content, _ := os.ReadFile(path)
		msg := &commands.AnyMessage{
			Action:        file.CreateFile,
			Type:          commands.Command,
			AggregateType: file.EntityName,
			Payload: file.CreateFilePayload{
				ProjectID:   projectID,
				Name:        filename,
				Description: getFileDescription(string(content)),
				Content:     string(content),
			},
			Timestamp:   time.Now(),
			AggregateID: utils.NewID(),
		}

		utils.Must(publisher.Publish(msg))
		imported++
	}

	slog.Info("import complete", "imported", imported, "total", len(files))
	return nil
}

func getExistingFileNames() map[string]bool {
	disk := persistence.New()
	events, _ := disk.LoadAllEvents()
	fileNames := make(map[string]bool)

	for _, event := range events {
		if event.Action == file.CreatedFile {
			if payload, ok := event.Payload.(map[string]interface{}); ok {
				if name, ok := payload["Name"].(string); ok {
					fileNames[name] = true
				}
			}
		}
	}

	return fileNames
}

func getOrCreateProjectID(publish dispatch.PublishFunc) string {
	disk := persistence.New()
	events, _ := disk.LoadAllEvents()

	for _, event := range events {
		if event.AggregateType == project.EntityName && event.Action == project.CreatedProject {
			return event.AggregateID
		}
	}

	msg := commands.ToAny(commands.NewCommand[project.CreateProjectPayload, any](
		project.CreateProject,
		project.CreateProjectPayload{Name: "COVID Project"},
		project.EntityName,
		utils.NewID(),
		nil,
	))

	result, _ := publish(msg)
	return result.AggregateID
}
