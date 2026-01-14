package document

import "hermes-relay/internal/lib/utils"

func AssignAnnotationIDs(annotations []Annotation) []Annotation {
	return utils.Map(annotations, func(a Annotation) Annotation {
		if a.ID == "" {
			a.ID = utils.NewAnnotationID()
		}
		return a
	})
}

func AddAnnotation(current []Annotation, ann Annotation) []Annotation {
	result := make([]Annotation, len(current)+1)
	copy(result, current)
	result[len(current)] = ann
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
