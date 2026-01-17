package document

import (
	"hermes-relay/internal/lib/normalizer"
	"hermes-relay/internal/lib/utils"
	"maps"
)

func NormalizeTag(tag string) string {
	return normalizer.NormalizeValue(tag, normalizer.Trim, normalizer.Collapse, normalizer.Lowercase)
}

func NormalizeTags(tags []string) []string {
	return utils.Map(tags, NormalizeTag)
}

func AddTags(current map[string]Tag, add []string) map[string]Tag {
	result := cloneOrInit(current)
	for _, tag := range NormalizeTags(add) {
		if tag != "" {
			result[tag] = Tag{ID: tag}
		}
	}
	return result
}

func RemoveTags(current map[string]Tag, remove []string) map[string]Tag {
	result := cloneOrInit(current)
	for _, tag := range NormalizeTags(remove) {
		delete(result, tag)
	}
	return result
}

func cloneOrInit[K comparable, V any](m map[K]V) map[K]V {
	if m == nil {
		return make(map[K]V)
	}
	return maps.Clone(m)
}
