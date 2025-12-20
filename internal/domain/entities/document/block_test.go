package document

import (
	"testing"

	th "hermes-relay/internal/lib/test-helpers"
)

func TestValidateBlock(t *testing.T) {
	T, F := true, false
	h := func(level int) BlockProps { return BlockProps{HeadingProps: HeadingProps{Level: level}} }
	chk := func(v bool) BlockProps { return BlockProps{CheckListProps: CheckListProps{Checked: &v}} }
	img := func(url string) BlockProps { return BlockProps{MediaProps: MediaProps{URL: url}} }

	th.RunFunctionTests(t, []struct {
		Name     string
		Input    Block
		Expected bool
	}{
		{"valid paragraph", Block{ID: "b1", Type: BlockTypeParagraph}, F},
		{"missing ID", Block{Type: BlockTypeParagraph}, T},
		{"valid heading level 1", Block{ID: "b1", Type: BlockTypeHeading, Props: h(1)}, F},
		{"valid heading level 6", Block{ID: "b1", Type: BlockTypeHeading, Props: h(6)}, F},
		{"heading missing level", Block{ID: "b1", Type: BlockTypeHeading}, T},
		{"heading level too high", Block{ID: "b1", Type: BlockTypeHeading, Props: h(7)}, T},
		{"valid checklist checked", Block{ID: "b1", Type: BlockTypeCheckList, Props: chk(true)}, F},
		{"valid checklist unchecked", Block{ID: "b1", Type: BlockTypeCheckList, Props: chk(false)}, F},
		{"checklist missing checked", Block{ID: "b1", Type: BlockTypeCheckList}, T},
		{"valid image with URL", Block{ID: "b1", Type: BlockTypeImage, Props: img("https://example.com/img.png")}, F},
		{"image missing URL", Block{ID: "b1", Type: BlockTypeImage}, T},
		{"valid code block", Block{ID: "b1", Type: BlockTypeCodeBlock}, F},
		{"valid bullet list", Block{ID: "b1", Type: BlockTypeBulletList}, F},
		{"valid table", Block{ID: "b1", Type: BlockTypeTable}, F},
		{"valid nested children", Block{ID: "b1", Type: BlockTypeParagraph, Children: []Block{{ID: "b2", Type: BlockTypeParagraph}}}, F},
		{"invalid nested child", Block{ID: "b1", Type: BlockTypeParagraph, Children: []Block{{ID: "b2", Type: BlockTypeHeading}}}, T},
		{"unknown block type", Block{ID: "b1", Type: "unknown"}, T},
	}, ValidateBlock, func(err error) bool { return err != nil })
}
