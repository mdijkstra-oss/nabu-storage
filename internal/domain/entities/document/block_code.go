package document

const BlockTypeCodeBlock BlockType = "codeBlock"

type CodeProps struct {
	Language string `json:"language,omitempty"`
}

func init() {
	RegisterBlockType(BlockTypeCodeBlock, CodeProps{})
}
