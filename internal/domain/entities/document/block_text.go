package document

const (
	BlockTypeParagraph BlockType = "paragraph"
	BlockTypeHeading   BlockType = "heading"
	BlockTypeQuote     BlockType = "quote"
)

type HeadingProps struct {
	Level int `json:"level,omitempty" validate:"required,min=1,max=6"`
}

func init() {
	RegisterBlockType(BlockTypeParagraph, nil)
	RegisterBlockType(BlockTypeHeading, HeadingProps{})
	RegisterBlockType(BlockTypeQuote, nil)
}
