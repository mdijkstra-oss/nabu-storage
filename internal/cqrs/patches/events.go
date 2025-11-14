package patches

import (
	"encoding/json"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/lib/utils"
)

const (
	ProjectPatched   = "ProjectPatched"
	ProjectSnapshot  = "ProjectSnapshot"
	PatchesAggregate = "Patches"
)

type ProjectEventPayload interface {
	GetProjectID() string
	GetData() any
}

type PatchEventPayload struct {
	ProjectID string `json:"project_id"`
	Patch     []byte `json:"patch"`
}

func (p PatchEventPayload) GetProjectID() string {
	return p.ProjectID
}

func (p PatchEventPayload) GetData() any {
	var data any
	utils.ShouldWork(json.Unmarshal(p.Patch, &data))
	return data
}

type SnapshotEventPayload struct {
	ProjectID string `json:"project_id"`
	Snapshot  any    `json:"snapshot"`
}

func (p SnapshotEventPayload) GetProjectID() string {
	return p.ProjectID
}

func (p SnapshotEventPayload) GetData() any {
	return p.Snapshot
}

func NewPatchEvent(projectID string, patch []byte) *commands.AnyMessage {
	return commands.ToAny(commands.NewSystemEvent[PatchEventPayload, any](
		ProjectPatched,
		PatchEventPayload{
			ProjectID: projectID,
			Patch:     patch,
		},
		PatchesAggregate,
		projectID,
		nil,
	))
}

func NewSnapshotEvent(projectID string, snapshot any) *commands.AnyMessage {
	return commands.ToAny(commands.NewSystemEvent[SnapshotEventPayload, any](
		ProjectSnapshot,
		SnapshotEventPayload{
			ProjectID: projectID,
			Snapshot:  snapshot,
		},
		PatchesAggregate,
		projectID,
		nil,
	))
}
