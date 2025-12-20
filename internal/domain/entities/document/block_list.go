package document

const (
	BlockTypeBulletList   BlockType = "bulletListItem"
	BlockTypeNumberedList BlockType = "numberedListItem"
	BlockTypeCheckList    BlockType = "checkListItem"
	BlockTypeToggleList   BlockType = "toggleListItem"
)

type CheckListProps struct {
	Checked *bool `json:"checked,omitempty" validate:"required"`
}

func init() {
	RegisterBlockType(BlockTypeBulletList, nil)
	RegisterBlockType(BlockTypeNumberedList, nil)
	RegisterBlockType(BlockTypeCheckList, CheckListProps{})
	RegisterBlockType(BlockTypeToggleList, nil)
}
