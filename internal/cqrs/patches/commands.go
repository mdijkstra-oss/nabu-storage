package patches

const RequestSnapshot = "RequestSnapshot"

type RequestSnapshotPayload struct {
	ProjectID string `json:"project_id" validate:"required"`
}
