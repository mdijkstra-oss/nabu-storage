package chunk

import (
	"hermes-relay/internal/domain/entities/file"
	textsearch "hermes-relay/internal/lib/text-search"
	"hermes-relay/internal/lib/utils"
)

func CalculateChunkCoverage(chunk file.Chunk, codeSlugs []string) float64 {
	sections := chunk.Codes
	if len(codeSlugs) > 0 {
		sections = utils.Filter(sections, func(cs file.CodedSection) bool {
			return utils.Contains(codeSlugs, cs.CodeSlug)
		})
	}

	subTexts := utils.Map(sections, func(cs file.CodedSection) string {
		return cs.Text
	})

	return textsearch.CalculateCoverage(chunk.Content, subTexts)
}

func FilterChunksByText(chunks []file.Chunk, searchText string) []file.Chunk {
	if searchText == "" {
		return chunks
	}
	return utils.Filter(chunks, func(chunk file.Chunk) bool {
		return textsearch.ContainsText(chunk.Content, searchText)
	})
}

func FilterChunksByCoverage(chunks []file.Chunk, minCoverage, maxCoverage *float64, codeSlugs []string) []file.Chunk {
	return utils.Filter(chunks, func(chunk file.Chunk) bool {
		coverage := CalculateChunkCoverage(chunk, codeSlugs)

		if minCoverage != nil && coverage < *minCoverage {
			return false
		}
		if maxCoverage != nil && coverage > *maxCoverage {
			return false
		}
		return true
	})
}
