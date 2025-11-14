package patches

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/domain/projections/registry"
	"log/slog"
)

func WrapRegistryWithPatching(
	registryState *registry.RegistryState,
	publisher *dispatch.InMemoryPublisher,
	activeChecker ActiveProjectChecker,
) func(*commands.AnyMessage) {
	return func(event *commands.AnyMessage) {
		projectID := extractProjectIDForEvent(registryState, event)
		shouldGeneratePatches := shouldGeneratePatchesForEvent(event, projectID, activeChecker)

		var beforeProject *project.Project
		if shouldGeneratePatches {
			beforeProject = registryState.GetProject(projectID)
		}

		registryState.ApplyEvent(event)

		if shouldGeneratePatches {
			afterProject := registryState.GetProject(projectID)
			emitPatchOrSnapshot(publisher, projectID, beforeProject, afterProject)
		}
	}
}

func shouldGeneratePatchesForEvent(event *commands.AnyMessage, projectID string, activeChecker ActiveProjectChecker) bool {
	if event.Type != commands.DomainEvent {
		return false
	}
	if projectID == "" {
		return false
	}
	return activeChecker.IsActive(projectID)
}

func emitPatchOrSnapshot(publisher *dispatch.InMemoryPublisher, projectID string, before, after *project.Project) {
	if after == nil || !after.IsPatchable() {
		return
	}

	if before == nil {
		emitSnapshot(publisher, projectID, after)
		return
	}

	emitPatch(publisher, projectID, before, after)
}

func emitSnapshot(publisher *dispatch.InMemoryPublisher, projectID string, snapshot *project.Project) {
	if _, err := publisher.Publish(NewSnapshotEvent(projectID, snapshot)); err != nil {
		slog.Error("failed to publish snapshot event", "projectID", projectID, "error", err)
	}
}

func emitPatch(publisher *dispatch.InMemoryPublisher, projectID string, before, after *project.Project) {
	patch, err := GeneratePatch(before, after)
	if err != nil {
		slog.Error("failed to generate patch", "projectID", projectID, "error", err)
		return
	}

	if _, err := publisher.Publish(NewPatchEvent(projectID, patch)); err != nil {
		slog.Error("failed to publish patch event", "projectID", projectID, "error", err)
	}
}

func extractProjectIDForEvent(registryState *registry.RegistryState, event *commands.AnyMessage) string {
	if event.AggregateType == "Project" {
		return event.AggregateID
	}

	projectID := commands.ExtractProjectID(event)
	if projectID != "" {
		return projectID
	}

	return registryState.GetProjectIDForEntity(event.AggregateType, event.AggregateID)
}
