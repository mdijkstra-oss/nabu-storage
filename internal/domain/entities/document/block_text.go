package document

const (
	BlockTypeParagraph BlockType = "paragraph"
	BlockTypeHeading   BlockType = "heading"
	BlockTypeQuote     BlockType = "quote"
)

type HeadingExtraProps struct {
	Level int `validate:"required,min=1,max=6"`
}

func init() {
	RegisterBlockType(BlockTypeParagraph, nil)
	RegisterBlockType(BlockTypeHeading, HeadingExtraProps{})
	RegisterBlockType(BlockTypeQuote, nil)
}
