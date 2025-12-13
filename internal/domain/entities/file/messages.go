package file

type AddSectionOp struct {
	CodeID     string     `json:"code_id" validate:"required,valid_id_or_slug"`
	Text       string     `json:"text" validate:"required,min=1,max=1500"`
	Reason     string     `json:"reason" validate:"max=1500"`
	Confidence Confidence `json:"confidence" validate:"required,oneof=high medium low"`
}

type UpdateSectionOp struct {
	ID         string      `json:"id" validate:"required,valid_id"`
	CodeID     string      `json:"code_id,omitempty" validate:"omitempty,valid_id_or_slug"`
	Text       string      `json:"text,omitempty" validate:"omitempty,min=1,max=1500"`
	Reason     string      `json:"reason,omitempty" validate:"max=1500"`
	Confidence *Confidence `json:"confidence,omitempty" validate:"omitempty,oneof=high medium low"`
}

type AddedSection struct {
	ID         string     `json:"id" validate:"required,valid_id"`
	CodeID     string     `json:"code_id" validate:"required,valid_id"`
	Text       string     `json:"text" validate:"required,min=1,max=1500"`
	Reason     string     `json:"reason" validate:"max=1500"`
	Confidence Confidence `json:"confidence" validate:"required,oneof=high medium low"`
}

type AddCodeSectionsPayload struct {
	Sections []AddSectionOp `json:"sections" validate:"min=1,dive"`
	Failures map[int]string `json:"failures,omitempty"`
}

type AddedCodeSectionsPayload struct {
	Sections []AddedSection `json:"sections" validate:"min=1,dive"`
	Failures map[int]string `json:"failures,omitempty"`
}

func (p AddedCodeSectionsPayload) GetFailures() map[int]string {
	return p.Failures
}

func (p AddedCodeSectionsPayload) WithoutFailures() any {
	return AddedCodeSectionsPayload{
		Sections: p.Sections,
	}
}

type UpdateCodeSectionsPayload struct {
	Sections []UpdateSectionOp `json:"sections" validate:"min=1,dive"`
	Failures map[int]string    `json:"failures,omitempty"`
}

type UpdatedCodeSectionsPayload struct {
	Sections []UpdateSectionOp `json:"sections" validate:"min=1,dive"`
	Failures map[int]string    `json:"failures,omitempty"`
}

func (p UpdatedCodeSectionsPayload) GetFailures() map[int]string {
	return p.Failures
}

func (p UpdatedCodeSectionsPayload) WithoutFailures() any {
	return UpdatedCodeSectionsPayload{
		Sections: p.Sections,
	}
}

type RemoveCodeSectionsPayload struct {
	SectionIDs []string `json:"section_ids" validate:"min=1,dive,required,valid_id"`
}

type SectionOp struct {
	Op         string      `json:"op" validate:"required,oneof=add update delete"`
	ID         string      `json:"id,omitempty" validate:"omitempty,valid_id"`
	CodeID     string      `json:"code_id,omitempty" validate:"omitempty,valid_id"`
	Text       string      `json:"text,omitempty" validate:"omitempty,min=1,max=1500"`
	Reason     string      `json:"reason,omitempty" validate:"max=1500"`
	Confidence *Confidence `json:"confidence,omitempty" validate:"omitempty,oneof=high medium low"`
}

type ModifiedCodeSectionsPayload struct {
	Operations []SectionOp    `json:"operations" validate:"min=1,dive"`
	Failures   map[int]string `json:"failures,omitempty"`
}

func (p ModifiedCodeSectionsPayload) GetFailures() map[int]string {
	return p.Failures
}

func (p ModifiedCodeSectionsPayload) WithoutFailures() any {
	return ModifiedCodeSectionsPayload{
		Operations: p.Operations,
	}
}
