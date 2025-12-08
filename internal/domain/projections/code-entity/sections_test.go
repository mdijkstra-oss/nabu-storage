package codeview

import (
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
)

var getSectionsTests = []struct {
	Name     string
	Input    string
	Expected []string
}{
	{Name: "Gets all sections for code-1 across files", Input: "code-1", Expected: []string{"f1-c1-s1", "f1-c2-s1", "f2-c1-s1"}},
	{Name: "Gets all sections for code-2", Input: "code-2", Expected: []string{"f1-c1-s2"}},
	{Name: "Returns empty for nonexistent code", Input: "code-99", Expected: []string{}},
}

var filterTests = []struct {
	Name     string
	Input    SectionFilter
	Expected []string
}{
	{Name: "No filter returns all", Input: SectionFilter{}, Expected: []string{"f1-c1-s1", "f1-c2-s1", "f2-c1-s1"}},
	{Name: "Filter by high confidence", Input: SectionFilter{Confidence: "high"}, Expected: []string{"f1-c1-s1", "f2-c1-s1"}},
	{Name: "Filter by low confidence", Input: SectionFilter{Confidence: "low"}, Expected: []string{"f1-c2-s1"}},
	{Name: "Filter by human actor", Input: SectionFilter{ActorType: "human"}, Expected: []string{"f1-c1-s1", "f2-c1-s1"}},
	{Name: "Filter by llm actor", Input: SectionFilter{ActorType: "llm"}, Expected: []string{"f1-c2-s1"}},
	{Name: "Filter by both confidence and actor", Input: SectionFilter{Confidence: "high", ActorType: "human"}, Expected: []string{"f1-c1-s1", "f2-c1-s1"}},
	{Name: "Filter with no matches", Input: SectionFilter{Confidence: "low", ActorType: "human"}, Expected: []string{}},
}

var queryTests = []struct {
	Name     string
	Input    SectionsQuery
	Expected []string
}{
	{Name: "Returns all sections", Input: SectionsQuery{PaginationQuery: projection.PaginationQuery{Page: 1, PageSize: 10}, ID: "code-1"}, Expected: []string{"f1-c1-s1", "f1-c2-s1", "f2-c1-s1"}},
	{Name: "Pagination limits results", Input: SectionsQuery{PaginationQuery: projection.PaginationQuery{Page: 1, PageSize: 2}, ID: "code-1"}, Expected: []string{"f1-c1-s1", "f1-c2-s1"}},
	{Name: "Page 2", Input: SectionsQuery{PaginationQuery: projection.PaginationQuery{Page: 2, PageSize: 2}, ID: "code-1"}, Expected: []string{"f2-c1-s1"}},
	{Name: "Filter by confidence", Input: SectionsQuery{PaginationQuery: projection.PaginationQuery{Page: 1, PageSize: 10}, SectionFilter: SectionFilter{Confidence: "high"}, ID: "code-1"}, Expected: []string{"f1-c1-s1", "f2-c1-s1"}},
	{Name: "Filter by actor", Input: SectionsQuery{PaginationQuery: projection.PaginationQuery{Page: 1, PageSize: 10}, SectionFilter: SectionFilter{ActorType: "llm"}, ID: "code-1"}, Expected: []string{"f1-c2-s1"}},
	{Name: "Nonexistent code", Input: SectionsQuery{PaginationQuery: projection.PaginationQuery{Page: 1, PageSize: 10}, ID: "code-99"}, Expected: nil},
}

func TestSections(t *testing.T) {
	proj := project.BuildTestProjectWithData()
	allSections := GetSectionsForCode(proj, "code-1")

	th.RunFunctionTests(t, getSectionsTests, func(codeID string) []CodedSectionView {
		return GetSectionsForCode(proj, codeID)
	}, extractSectionIDs)

	th.RunFunctionTests(t, filterTests, func(filter SectionFilter) []CodedSectionView {
		return FilterSections(allSections, filter)
	}, extractSectionIDs)

	th.RunFunctionTests(t, queryTests, func(q SectionsQuery) *projection.PaginationResult[CodedSectionView] {
		return QuerySections(q, proj)
	}, extractResultIDs)
}

func TestCodedSectionViewContainsFileContext(t *testing.T) {
	proj := project.BuildTestProjectWithData()
	sections := GetSectionsForCode(proj, "code-1")

	s1 := findSection(sections, "f1-c1-s1")
	th.AssertEqual(t, s1.FileID, "file-1", "file_id")
	th.AssertEqual(t, s1.FileName, "Interview A", "file_name")
	th.AssertEqual(t, s1.Text, "worried about", "text")
	th.AssertEqual(t, s1.Confidence, file.ConfidenceHigh, "confidence")

	s4 := findSection(sections, "f2-c1-s1")
	th.AssertEqual(t, s4.FileID, "file-2", "file_id")
	th.AssertEqual(t, s4.FileName, "Interview B", "file_name")
}

func extractSectionIDs(sections []CodedSectionView) []string {
	ids := make([]string, len(sections))
	for i, s := range sections {
		ids[i] = s.ID
	}
	return ids
}

func extractResultIDs(result *projection.PaginationResult[CodedSectionView]) []string {
	if result == nil {
		return nil
	}
	return extractSectionIDs(result.Items)
}

func findSection(sections []CodedSectionView, id string) CodedSectionView {
	for _, s := range sections {
		if s.ID == id {
			return s
		}
	}
	return CodedSectionView{}
}
