package document

const BlockTypeTable BlockType = "table"

func init() {
	RegisterBlockType(BlockTypeTable, nil)
}
