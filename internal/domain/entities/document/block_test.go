package document

import (
	"testing"
)

func TestValidateBlock(t *testing.T) {
	tests := []struct {
		name    string
		block   Block
		wantErr bool
	}{
		{
			name:    "valid paragraph",
			block:   Block{ID: "b1", Type: BlockTypeParagraph},
			wantErr: false,
		},
		{
			name:    "missing ID",
			block:   Block{Type: BlockTypeParagraph},
			wantErr: true,
		},
		{
			name:    "valid heading level 1",
			block:   Block{ID: "b1", Type: BlockTypeHeading, Props: BlockProps{Level: 1}},
			wantErr: false,
		},
		{
			name:    "valid heading level 6",
			block:   Block{ID: "b1", Type: BlockTypeHeading, Props: BlockProps{Level: 6}},
			wantErr: false,
		},
		{
			name:    "heading missing level",
			block:   Block{ID: "b1", Type: BlockTypeHeading},
			wantErr: true,
		},
		{
			name:    "heading level too high",
			block:   Block{ID: "b1", Type: BlockTypeHeading, Props: BlockProps{Level: 7}},
			wantErr: true,
		},
		{
			name:    "valid checklist checked",
			block:   Block{ID: "b1", Type: BlockTypeCheckList, Props: BlockProps{Checked: boolPtr(true)}},
			wantErr: false,
		},
		{
			name:    "valid checklist unchecked",
			block:   Block{ID: "b1", Type: BlockTypeCheckList, Props: BlockProps{Checked: boolPtr(false)}},
			wantErr: false,
		},
		{
			name:    "checklist missing checked",
			block:   Block{ID: "b1", Type: BlockTypeCheckList},
			wantErr: true,
		},
		{
			name:    "valid image with URL",
			block:   Block{ID: "b1", Type: BlockTypeImage, Props: BlockProps{URL: "https://example.com/img.png"}},
			wantErr: false,
		},
		{
			name:    "image missing URL",
			block:   Block{ID: "b1", Type: BlockTypeImage},
			wantErr: true,
		},
		{
			name:    "valid code block",
			block:   Block{ID: "b1", Type: BlockTypeCodeBlock},
			wantErr: false,
		},
		{
			name:    "valid bullet list",
			block:   Block{ID: "b1", Type: BlockTypeBulletList},
			wantErr: false,
		},
		{
			name:    "valid table",
			block:   Block{ID: "b1", Type: BlockTypeTable},
			wantErr: false,
		},
		{
			name: "valid nested children",
			block: Block{
				ID:   "b1",
				Type: BlockTypeParagraph,
				Children: []Block{
					{ID: "b2", Type: BlockTypeParagraph},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid nested child",
			block: Block{
				ID:   "b1",
				Type: BlockTypeParagraph,
				Children: []Block{
					{ID: "b2", Type: BlockTypeHeading},
				},
			},
			wantErr: true,
		},
		{
			name:    "unknown block type",
			block:   Block{ID: "b1", Type: "unknown"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBlock(tt.block)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}
