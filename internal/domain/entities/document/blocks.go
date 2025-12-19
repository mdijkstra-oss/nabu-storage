package document

type Block struct {
	ID       string         `json:"id"`
	Type     BlockType      `json:"type"`
	Props    BlockProps     `json:"props,omitempty"`
	Content  []InlineContent `json:"content,omitempty"`
	Children []Block        `json:"children,omitempty"`
}

type BlockType string

const (
	BlockTypeParagraph    BlockType = "paragraph"
	BlockTypeHeading      BlockType = "heading"
	BlockTypeBulletList   BlockType = "bulletListItem"
	BlockTypeNumberedList BlockType = "numberedListItem"
	BlockTypeCheckList    BlockType = "checkListItem"
	BlockTypeTable        BlockType = "table"
	BlockTypeImage        BlockType = "image"
	BlockTypeVideo        BlockType = "video"
	BlockTypeAudio        BlockType = "audio"
	BlockTypeFile         BlockType = "file"
	BlockTypeCodeBlock    BlockType = "codeBlock"
)

type BlockProps struct {
	Level           int    `json:"level,omitempty"`
	BackgroundColor string `json:"backgroundColor,omitempty"`
	TextColor       string `json:"textColor,omitempty"`
	TextAlignment   string `json:"textAlignment,omitempty"`
	Checked         *bool  `json:"checked,omitempty"`
	Language        string `json:"language,omitempty"`
	URL             string `json:"url,omitempty"`
	Caption         string `json:"caption,omitempty"`
	Width           int    `json:"width,omitempty"`
}

type InlineContent struct {
	Type    InlineType `json:"type"`
	Text    string     `json:"text,omitempty"`
	Styles  Styles     `json:"styles,omitempty"`
	Href    string     `json:"href,omitempty"`
	Content []StyledText `json:"content,omitempty"`
}

type InlineType string

const (
	InlineTypeText InlineType = "text"
	InlineTypeLink InlineType = "link"
)

type StyledText struct {
	Type   InlineType `json:"type"`
	Text   string     `json:"text"`
	Styles Styles     `json:"styles,omitempty"`
}

type Styles struct {
	Bold            *bool  `json:"bold,omitempty"`
	Italic          *bool  `json:"italic,omitempty"`
	Underline       *bool  `json:"underline,omitempty"`
	Strikethrough   *bool  `json:"strikethrough,omitempty"`
	Code            *bool  `json:"code,omitempty"`
	TextColor       string `json:"textColor,omitempty"`
	BackgroundColor string `json:"backgroundColor,omitempty"`
}
