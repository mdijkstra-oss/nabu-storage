package document

const (
	BlockTypeImage BlockType = "image"
	BlockTypeVideo BlockType = "video"
	BlockTypeAudio BlockType = "audio"
	BlockTypeFile  BlockType = "file"
)

type MediaExtraProps struct {
	URL string `validate:"required"`
}

func init() {
	RegisterBlockType(BlockTypeImage, MediaExtraProps{})
	RegisterBlockType(BlockTypeVideo, MediaExtraProps{})
	RegisterBlockType(BlockTypeAudio, MediaExtraProps{})
	RegisterBlockType(BlockTypeFile, MediaExtraProps{})
}
