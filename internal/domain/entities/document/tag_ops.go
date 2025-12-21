package document

import (
	"hermes-relay/internal/lib/normalizer"
	"hermes-relay/internal/lib/utils"
	"slices"
)

func NormalizeTag(tag string) string {
	return normalizer.NormalizeValue(tag, normalizer.Trim, normalizer.Collapse, normalizer.Lowercase)
}

func NormalizeTags(tags []string) []string {
	return utils.Map(tags, NormalizeTag)
}

func AddTags(current []string, add []string) []string {
	set := utils.ToSet(filterEmpty(NormalizeTags(current)))
	for _, tag := range NormalizeTags(add) {
		if tag != "" {
			set[tag] = true
		}
	}
	result := utils.Keys(set)
	slices.Sort(result)
	return result
}

func filterEmpty(tags []string) []string {
	return utils.Filter(tags, func(tag string) bool { return tag != "" })
}

func RemoveTags(current []string, remove []string) []string {
	removeSet := utils.ToSet(NormalizeTags(remove))
	result := utils.Filter(filterEmpty(NormalizeTags(current)), func(tag string) bool {
		return !removeSet[tag]
	})
	slices.Sort(result)
	return result
}
