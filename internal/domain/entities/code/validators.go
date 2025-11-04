package code

import (
	"fmt"
)

func ValidateCreateCode(allCodes []Code, payload CreateCodePayload) error {
	for _, existingCode := range allCodes {
		if existingCode.ProjectID == payload.ProjectID && existingCode.Slug == payload.Slug {
			return fmt.Errorf("code with slug '%s' already exists in project '%s'", payload.Slug, payload.ProjectID)
		}
	}

	return nil
}
