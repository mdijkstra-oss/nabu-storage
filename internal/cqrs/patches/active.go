package patches

type ActiveProjectChecker interface {
	IsActive(projectID string) bool
}
