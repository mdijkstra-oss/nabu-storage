package document

const (
	BlockTypeImage BlockType = "image"
	BlockTypeVideo BlockType = "video"
	BlockTypeAudio BlockType = "audio"
	BlockTypeFile  BlockType = "file"
)

type MediaProps struct {
	URL          string `json:"url,omitempty" validate:"required"`
	Caption      string `json:"caption,omitempty"`
	PreviewWidth int    `json:"previewWidth,omitempty"`
	Name         string `json:"name,omitempty"`
}

func init() {
	RegisterBlockType(BlockTypeImage, MediaProps{})
	RegisterBlockType(BlockTypeVideo, MediaProps{})
	RegisterBlockType(BlockTypeAudio, MediaProps{})
	RegisterBlockType(BlockTypeFile, MediaProps{})
}
