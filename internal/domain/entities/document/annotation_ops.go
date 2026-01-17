package document

import (
	"hermes-relay/internal/lib/text-search/find"
	"hermes-relay/internal/lib/utils"
	"maps"
	"strings"
)

func normalizeForComparison(text string) string {
	return strings.ToLower(find.NormalizeText(text))
}

func hasMatchingAnnotation(current map[string]Annotation, text string) bool {
	normalized := normalizeForComparison(text)
	for _, ann := range current {
		if normalizeForComparison(ann.Text) == normalized {
			return true
		}
	}
	return false
}

func AddAnnotation(current map[string]Annotation, ann Annotation) map[string]Annotation {
	if hasMatchingAnnotation(current, ann.Text) {
		return current
	}
	result := maps.Clone(current)
	if result == nil {
		result = make(map[string]Annotation)
	}
	result[ann.ID] = ann
	return result
}

func RemoveAnnotations(current map[string]Annotation, removeIDs []string) map[string]Annotation {
	removeSet := utils.ToSet(removeIDs)
	result := make(map[string]Annotation, len(current))
	for k, v := range current {
		if !removeSet[k] {
			result[k] = v
		}
	}
	return result
}

func UpdateAnnotationProps(current map[string]Annotation, ids []string, props AnnotationPropsUpdate) map[string]Annotation {
	updateSet := utils.ToSet(ids)
	result := make(map[string]Annotation, len(current))
	for k, v := range current {
		if updateSet[k] {
			if props.Color != nil {
				v.Color = *props.Color
			}
			if props.Reason != nil {
				v.Reason = *props.Reason
			}
			if props.Payload != nil {
				v.Payload = props.Payload
			}
		}
		result[k] = v
	}
	return result
}
