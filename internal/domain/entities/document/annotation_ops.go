package document

import "hermes-relay/internal/lib/utils"

func AddAnnotations(current []Annotation, add []Annotation) []Annotation {
	existingIDs := utils.ToSet(utils.Map(current, func(a Annotation) string { return a.ID }))
	result := make([]Annotation, len(current))
	copy(result, current)

	for _, ann := range add {
		if !existingIDs[ann.ID] {
			result = append(result, ann)
		}
	}
	return result
}

func RemoveAnnotations(current []Annotation, removeIDs []string) []Annotation {
	removeSet := utils.ToSet(removeIDs)
	return utils.Filter(current, func(a Annotation) bool {
		return !removeSet[a.ID]
	})
}

func UpdateAnnotationProps(current []Annotation, ids []string, props AnnotationPropsUpdate) []Annotation {
	updateSet := utils.ToSet(ids)
	return utils.Map(current, func(a Annotation) Annotation {
		if !updateSet[a.ID] {
			return a
		}
		if props.Color != nil {
			a.Color = *props.Color
		}
		if props.Reason != nil {
			a.Reason = *props.Reason
		}
		if props.Payload != nil {
			a.Payload = props.Payload
		}
		return a
	})
}
