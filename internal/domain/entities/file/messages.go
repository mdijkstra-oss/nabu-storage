package file

type AddSectionOp struct {
	CodeSlug string `json:"code_slug" validate:"required,code_slug"`
	CodeID   string `json:"code_id" validate:"required"`
	Text     string `json:"text" validate:"required,min=1,max=1500"`
	Reason   string `json:"reason" validate:"max=1500"`
}

type UpdateSectionOp struct {
	ID     string `json:"id" validate:"required,valid_id"`
	Text   string `json:"text,omitempty" validate:"omitempty,min=1,max=1500"`
	Reason string `json:"reason,omitempty" validate:"max=1500"`
}

type AddedSection struct {
	ID       string `json:"id" validate:"required,valid_id"`
	CodeSlug string `json:"code_slug" validate:"required,code_slug"`
	CodeID   string `json:"code_id" validate:"required"`
	Text     string `json:"text" validate:"required,min=1,max=1500"`
	Reason   string `json:"reason" validate:"max=1500"`
}

type AddCodeSectionsPayload struct {
	ChunkID  string         `json:"chunk_id" validate:"required"`
	Sections []AddSectionOp `json:"sections" validate:"min=1,dive"`
}

type AddedCodeSectionsPayload struct {
	ChunkID  string          `json:"chunk_id" validate:"required"`
	Sections []AddedSection  `json:"sections" validate:"min=1,dive"`
}

type UpdateCodeSectionsPayload struct {
	ChunkID  string            `json:"chunk_id" validate:"required"`
	Sections []UpdateSectionOp `json:"sections" validate:"min=1,dive"`
}

type RemoveCodeSectionsPayload struct {
	ChunkID    string   `json:"chunk_id" validate:"required"`
	SectionIDs []string `json:"section_ids" validate:"min=1,dive,required,valid_id"`
}
