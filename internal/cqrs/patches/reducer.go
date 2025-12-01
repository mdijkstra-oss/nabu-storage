package patches

import (
	"bytes"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/lib/utils"
	"log/slog"
)

const (
	ActionTypeNone     = "none"
	ActionTypeSnapshot = "snapshot"
	ActionTypePatch    = "patch"
)

type PatchAction struct {
	Type     string
	Patch    []byte
	Snapshot any
}

func DecidePatch(before, after *project.Project, isActive bool) (PatchAction, error) {
	if !isActive {
		return PatchAction{Type: ActionTypeNone}, nil
	}

	if after == nil || !after.IsPatchable() {
		return PatchAction{Type: ActionTypeNone}, nil
	}

	if before == nil {
		return PatchAction{Type: ActionTypeSnapshot, Snapshot: after}, nil
	}

	patch, err := GeneratePatch(before, after)
	if err != nil {
		return PatchAction{Type: ActionTypeNone}, err
	}

	if bytes.Equal(patch, []byte("null")) {
		return PatchAction{Type: ActionTypeNone}, nil
	}

	return PatchAction{Type: ActionTypePatch, Patch: patch}, nil
}

func EmitPatchAction(publish dispatch.PublishFunc, projectID string, action PatchAction) {
	utils.GuardWith(func() {
		switch action.Type {
		case ActionTypeSnapshot:
			if _, err := publish(NewSnapshotEvent(projectID, action.Snapshot)); err != nil {
				slog.Error("failed to publish snapshot event", "projectID", projectID, "error", err)
			}
		case ActionTypePatch:
			if _, err := publish(NewPatchEvent(projectID, action.Patch)); err != nil {
				slog.Error("failed to publish patch event", "projectID", projectID, "error", err)
			}
		case ActionTypeNone:
			// Nothing to emit
		default:
			panic("EmitPatchAction: unknown action type: " + action.Type)
		}
	}, "projectID", projectID, "actionType", action.Type, "operation", "emitPatchAction")
}
