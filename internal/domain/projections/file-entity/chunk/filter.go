package chunk

import (
	"hermes-relay/internal/domain/entities/file"
	textsearch "hermes-relay/internal/lib/text-search"
	"hermes-relay/internal/lib/utils"
)

type ChunkFilter struct {
	SearchText  string   `query:"searchText"`
	MinCoverage *float64 `query:"minCoverage"`
	MaxCoverage *float64 `query:"maxCoverage"`
	CodeIDs     []string `query:"codeIds"`
}

func CalculateChunkCoverage(chunk file.Chunk, codeIDs []string) float64 {
	sections := chunk.Codes
	if len(codeIDs) > 0 {
		sections = utils.Filter(sections, func(cs file.CodedSection) bool {
			return utils.Contains(codeIDs, cs.CodeID)
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

func FilterChunksByCoverage(chunks []file.Chunk, minCoverage, maxCoverage *float64, codeIDs []string) []file.Chunk {
	return utils.Filter(chunks, func(chunk file.Chunk) bool {
		coverage := CalculateChunkCoverage(chunk, codeIDs)

		if minCoverage != nil && coverage < *minCoverage {
			return false
		}
		if maxCoverage != nil && coverage > *maxCoverage {
			return false
		}
		return true
	})
}

func ApplyFilter(chunks []file.Chunk, filter ChunkFilter) []file.Chunk {
	if filter.SearchText != "" {
		chunks = FilterChunksByText(chunks, filter.SearchText)
	}

	if filter.MinCoverage != nil || filter.MaxCoverage != nil || len(filter.CodeIDs) > 0 {
		chunks = FilterChunksByCoverage(chunks, filter.MinCoverage, filter.MaxCoverage, filter.CodeIDs)
	}

	return chunks
}
