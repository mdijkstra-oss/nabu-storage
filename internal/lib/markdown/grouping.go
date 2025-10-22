package markdown

import (
	"regexp"
	"strings"
)

type BlockSize = int

const (
	HalfPage BlockSize = 1500
	FullPage BlockSize = 3000
)

func ParseBlocks(text string, minBlockSize BlockSize) []string {
	lines := strings.Split(text, "\n")
	blocks := []string{}

	var current string
	var currentType string
	var codeMarker string

	matchers := []struct {
		pattern   *regexp.Regexp
		blockType string
	}{
		{regexp.MustCompile(`^\s*\|`), "table"},
		{regexp.MustCompile(`^(\s*)([*\-+]|\d+\.)\s`), "list"},
		{regexp.MustCompile(`^\s*>`), "blockquote"},
		{regexp.MustCompile(`^(    |\t)`), "code"},
		{regexp.MustCompile(`^(\*\*\*|---|___)(\s*)$`), "rule"},
	}

	addBlock := func() {
		if current != "" {
			blocks = append(blocks, current)
			current = ""
			currentType = ""
		}
	}

	startBlock := func(bType, content string) {
		addBlock()
		current = content
		currentType = bType
	}

	for _, line := range lines {
		if match := regexp.MustCompile("^(```|~~~)").FindString(line); match != "" {
			if codeMarker == "" {
				startBlock("code", line+"\n")
				codeMarker = match
			} else if strings.HasPrefix(line, codeMarker) {
				current += line + "\n"
				addBlock()
				codeMarker = ""
			} else {
				current += line + "\n"
			}
			continue
		}

		if codeMarker != "" {
			current += line + "\n"
			continue
		}

		if strings.TrimSpace(line) == "" {
			addBlock()
			continue
		}

		matched := false
		for _, m := range matchers {
			if m.pattern.MatchString(line) {
				if m.blockType == "rule" {
					addBlock()
					blocks = append(blocks, line+"\n")
				} else if current == "" || currentType != m.blockType {
					startBlock(m.blockType, line+"\n")
				} else {
					current += line + "\n"
				}
				matched = true
				break
			}
		}

		if !matched {
			if current == "" || currentType != "paragraph" {
				startBlock("paragraph", line+"\n")
			} else {
				current += line + "\n"
			}
		}
	}

	addBlock()
	return mergeSmallBlocks(mergeHeaders(blocks), minBlockSize)
}

func mergeHeaders(blocks []string) []string {
	result := []string{}
	for i := 0; i < len(blocks); i++ {
		if regexp.MustCompile(`^#+\s`).MatchString(blocks[i]) && i+1 < len(blocks) {
			result = append(result, blocks[i]+"\n"+blocks[i+1])
			i++
		} else {
			result = append(result, blocks[i])
		}
	}
	return result
}

func mergeSmallBlocks(blocks []string, minSize int) []string {
	if len(blocks) == 0 {
		return blocks
	}

	result := []string{}
	accumulated := blocks[0]

	for i := 1; i < len(blocks); i++ {
		if len(accumulated) < minSize {
			accumulated += "\n" + blocks[i]
		} else {
			result = append(result, accumulated)
			accumulated = blocks[i]
		}
	}
	result = append(result, accumulated)

	return result
}
