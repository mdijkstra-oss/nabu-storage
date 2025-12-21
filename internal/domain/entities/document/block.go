package document

import (
	"fmt"
	"hermes-relay/internal/lib/utils"
	"reflect"
)

type Block struct {
	ID       string          `json:"id"`
	Type     BlockType       `json:"type"`
	Props    BlockProps      `json:"props,omitempty"`
	Content  []InlineContent `json:"content,omitempty"`
	Children []Block         `json:"children,omitempty"`
}

type BlockType string

type CommonBlockProps struct {
	BackgroundColor string `json:"background_color,omitempty"`
	TextColor       string `json:"text_color,omitempty"`
	TextAlignment   string `json:"text_alignment,omitempty"`
}

type BlockProps struct {
	CommonBlockProps
	HeadingProps
	CheckListProps
	MediaProps
	CodeProps
}

type InlineContent struct {
	Type    InlineType   `json:"type"`
	Text    string       `json:"text,omitempty"`
	Styles  Styles       `json:"styles,omitempty"`
	Href    string       `json:"href,omitempty"`
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
	TextColor       string `json:"text_color,omitempty"`
	BackgroundColor string `json:"background_color,omitempty"`
}

var blockExtraProps = map[BlockType]reflect.Type{}

func RegisterBlockType(t BlockType, extraProps any) {
	if extraProps == nil {
		blockExtraProps[t] = nil
		return
	}
	blockExtraProps[t] = reflect.TypeOf(extraProps)
}

func ValidateBlock(b Block) error {
	if b.ID == "" {
		return fmt.Errorf("block ID is required")
	}

	propsType, exists := blockExtraProps[b.Type]
	if !exists {
		return fmt.Errorf("block %s: unknown type %q", b.ID, b.Type)
	}

	if propsType != nil {
		extra := reflect.New(propsType).Interface()
		if err := utils.CopyMatchingFields(b.Props, extra); err != nil {
			return fmt.Errorf("block %s: %w", b.ID, err)
		}
		if err := utils.Validate.Struct(extra); err != nil {
			return fmt.Errorf("block %s: %w", b.ID, err)
		}
	}

	for _, child := range b.Children {
		if err := ValidateBlock(child); err != nil {
			return err
		}
	}

	return nil
}

func ValidateBlocks(blocks []Block) error {
	for _, b := range blocks {
		if err := ValidateBlock(b); err != nil {
			return err
		}
	}
	return nil
}
