package codeview

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/lib/utils"
)

type CodedSectionView struct {
	ID         string            `json:"id"`
	FileID     string            `json:"file_id"`
	FileName   string            `json:"file_name"`
	ChunkID    string            `json:"chunk_id"`
	CodeID     string            `json:"code_id"`
	Text       string            `json:"text"`
	Reason     string            `json:"reason"`
	Confidence file.Confidence   `json:"confidence"`
	LastActor  commands.Actor    `json:"last_actor"`
}

func (s CodedSectionView) GetID() string {
	return s.ID
}

func (s CodedSectionView) GetActorType() string {
	return string(s.LastActor.ActorType)
}

type SectionFilter struct {
	Confidence string `query:"confidence" validate:"omitempty,oneof=high medium low"`
	ActorType  string `query:"actor_type" validate:"omitempty,oneof=human llm system"`
}

func GetSectionsForCode(proj project.Project, codeID string) []CodedSectionView {
	var sections []CodedSectionView

	for _, f := range proj.Files {
		for _, chunk := range f.Chunks {
			for _, section := range chunk.Codes {
				if section.CodeID == codeID {
					sections = append(sections, toSectionView(section, f, chunk))
				}
			}
		}
	}

	utils.Sort(sections, func(a, b CodedSectionView) bool {
		if a.FileName != b.FileName {
			return a.FileName < b.FileName
		}
		if a.ChunkID != b.ChunkID {
			return a.ChunkID < b.ChunkID
		}
		return a.ID < b.ID
	})

	return sections
}

func FilterSections(sections []CodedSectionView, filter SectionFilter) []CodedSectionView {
	if filter.Confidence != "" {
		sections = utils.Filter(sections, func(s CodedSectionView) bool {
			return string(s.Confidence) == filter.Confidence
		})
	}

	if filter.ActorType != "" {
		sections = utils.Filter(sections, func(s CodedSectionView) bool {
			return s.GetActorType() == filter.ActorType
		})
	}

	return sections
}

func toSectionView(section file.CodedSection, f file.File, chunk file.Chunk) CodedSectionView {
	return CodedSectionView{
		ID:         section.ID,
		FileID:     f.ID,
		FileName:   f.Name,
		ChunkID:    chunk.ID,
		CodeID:     section.CodeID,
		Text:       section.Text,
		Reason:     section.Reason,
		Confidence: section.Confidence,
		LastActor:  section.LastActor,
	}
}
