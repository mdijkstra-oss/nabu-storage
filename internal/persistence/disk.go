package persistence

import (
	"encoding/json"
	"fmt"
	"hermes-relay/internal/cqrs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
)

// getBasePath returns the base path for event persistence
func getBasePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		slog.Warn("failed to get user home directory, using relative path", "error", err)
		return "persistence-data"
	}
	return filepath.Join(home, "Documents", "hermes-persistence")
}

// PersistEvent writes a domain event to disk using append-only writes
// Events are stored in JSONL format (one JSON object per line)
// File structure: ~/Documents/hermes-persistence/{AggregateType}/{AggregateId}.jsonl
func PersistEvent(message *cqrs.AnyMessage) error {
	if message.AggregateType == "" {
		return fmt.Errorf("cannot persist event without aggregate type")
	}
	if message.AggregateID == "" {
		return fmt.Errorf("cannot persist event without aggregate ID")
	}

	filePath := getEventFilePath(message.AggregateType, message.AggregateID)

	if err := appendEventToFile(filePath, message); err != nil {
		return fmt.Errorf("failed to append event: %w", err)
	}

	slog.Debug("persisted event",
		"aggregateType", message.AggregateType,
		"aggregateId", message.AggregateID,
		"action", message.Action)

	return nil
}

// LoadAllEvents reads all persisted events from disk
// Returns events sorted by timestamp
func LoadAllEvents() ([]cqrs.AnyMessage, error) {
	var allEvents []cqrs.AnyMessage
	basePath := getBasePath()

	// Check if base path exists
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		slog.Info("no persistence directory found, starting fresh", "path", basePath)
		return allEvents, nil
	}

	// Walk through all aggregate type directories
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read persistence directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		aggregateType := entry.Name()
		typeEvents, err := loadEventsForType(aggregateType)
		if err != nil {
			return nil, fmt.Errorf("failed to load events for type %s: %w", aggregateType, err)
		}

		allEvents = append(allEvents, typeEvents...)
	}

	// Sort by timestamp
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].Timestamp.Before(allEvents[j].Timestamp)
	})

	slog.Info("loaded events from disk", "count", len(allEvents))
	return allEvents, nil
}

// loadEventsForType loads all events for a specific aggregate type
func loadEventsForType(aggregateType string) ([]cqrs.AnyMessage, error) {
	var events []cqrs.AnyMessage
	basePath := getBasePath()

	typePath := filepath.Join(basePath, aggregateType)
	entries, err := os.ReadDir(typePath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".jsonl" {
			continue
		}

		filePath := filepath.Join(typePath, entry.Name())
		fileEvents, err := readEventsFromFile(filePath)
		if err != nil {
			slog.Warn("failed to read event file",
				"path", filePath,
				"error", err)
			continue
		}

		events = append(events, fileEvents...)
	}

	return events, nil
}

// getEventFilePath constructs the file path for an aggregate's events
func getEventFilePath(aggregateType cqrs.AggregateType, aggregateID string) string {
	basePath := getBasePath()
	return filepath.Join(basePath, string(aggregateType), aggregateID+".jsonl")
}

// readEventsFromFile reads events from a JSONL file (one JSON object per line)
func readEventsFromFile(filePath string) ([]cqrs.AnyMessage, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var events []cqrs.AnyMessage
	decoder := json.NewDecoder(file)

	for decoder.More() {
		var event cqrs.AnyMessage
		if err := decoder.Decode(&event); err != nil {
			return nil, fmt.Errorf("failed to decode event: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}

// appendEventToFile appends a single event to a JSONL file
func appendEventToFile(filePath string, event *cqrs.AnyMessage) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Open file for appending (create if doesn't exist)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Encode event as JSON and write with newline
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(event); err != nil {
		return fmt.Errorf("failed to encode event: %w", err)
	}

	return nil
}

// Apply is a readonly route that persists domain events to disk
func Apply(message *cqrs.AnyMessage) error {
	if err := PersistEvent(message); err != nil {
		slog.Error("failed to persist event",
			"action", message.Action,
			"aggregateId", message.AggregateID,
			"error", err)
		return err
	}
	return nil
}

// GetProjectIDs reads all project IDs from disk
func GetProjectIDs() ([]string, error) {
	basePath := getBasePath()
	projectPath := filepath.Join(basePath, "Project")

	// Check if Project directory exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		slog.Info("no projects directory found", "path", projectPath)
		return []string{}, nil
	}

	entries, err := os.ReadDir(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read projects directory: %w", err)
	}

	var projectIDs []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".jsonl" {
			continue
		}

		// Remove .jsonl extension to get project ID
		projectID := entry.Name()[:len(entry.Name())-len(".jsonl")]
		projectIDs = append(projectIDs, projectID)
	}

	slog.Info("found projects on disk", "count", len(projectIDs))
	return projectIDs, nil
}

// ReplayAllEvents loads all events from disk and publishes them to the publisher
func ReplayAllEvents(publisher *cqrs.InMemoryPublisher) error {
	events, err := LoadAllEvents()
	if err != nil {
		return fmt.Errorf("failed to load events: %w", err)
	}

	if len(events) == 0 {
		slog.Info("no events to replay")
		return nil
	}

	slog.Info("replaying events", "count", len(events))

	for i, event := range events {
		if _, err := publisher.Publish(&event); err != nil {
			return fmt.Errorf("failed to replay event %d: %w", i, err)
		}
	}

	slog.Info("event replay complete", "count", len(events))
	return nil
}
