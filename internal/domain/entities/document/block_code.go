package document

const BlockTypeCodeBlock BlockType = "codeBlock"

func init() {
	RegisterBlockType(BlockTypeCodeBlock, nil)
}
