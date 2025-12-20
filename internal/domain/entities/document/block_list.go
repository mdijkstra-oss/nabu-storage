package document

const (
	BlockTypeBulletList   BlockType = "bulletListItem"
	BlockTypeNumberedList BlockType = "numberedListItem"
	BlockTypeCheckList    BlockType = "checkListItem"
	BlockTypeToggleList   BlockType = "toggleListItem"
)

type CheckListExtraProps struct {
	Checked *bool `validate:"required"`
}

func init() {
	RegisterBlockType(BlockTypeBulletList, nil)
	RegisterBlockType(BlockTypeNumberedList, nil)
	RegisterBlockType(BlockTypeCheckList, CheckListExtraProps{})
	RegisterBlockType(BlockTypeToggleList, nil)
}
