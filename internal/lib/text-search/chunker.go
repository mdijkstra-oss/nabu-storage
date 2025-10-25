package textsearch

import (
	"regexp"
	"strings"
)

type BlockSize = int

const (
	HalfPage BlockSize = 1500
	FullPage BlockSize = HalfPage * 2
)

func ChunkBlocks(text string, minBlockSize BlockSize, maxBlockSize BlockSize) []string {
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

	appendLine := func(line string) {
		newContent := current + line + "\n"

		// Never split block-level elements (tables, lists, code, blockquotes, rules)
		// Only apply size-based splitting to paragraphs for frontend rendering compatibility
		isBlockLevel := currentType != "paragraph" && currentType != ""

		if len(newContent) > int(maxBlockSize) && current != "" && !isBlockLevel {
			// Three-tier fallback for splitting long paragraph content
			splitPos := int(maxBlockSize)
			if splitPos > len(current) {
				splitPos = len(current)
			}

			// First priority: split at newline
			lastNewline := strings.LastIndex(current[:splitPos], "\n")
			if lastNewline > 0 {
				splitPos = lastNewline + 1
			} else {
				// Second priority: split at space (word boundary)
				lastSpace := strings.LastIndex(current[:splitPos], " ")
				if lastSpace > 0 {
					splitPos = lastSpace + 1
				}
				// Third priority: split mid-word (splitPos already set to maxBlockSize)
			}

			blocks = append(blocks, current[:splitPos])
			current = current[splitPos:] + line + "\n"
		} else {
			current = newContent
		}
	}

	for _, line := range lines {
		if match := regexp.MustCompile("^(```|~~~)").FindString(line); match != "" {
			if codeMarker == "" {
				startBlock("code", line+"\n")
				codeMarker = match
			} else if strings.HasPrefix(line, codeMarker) {
				appendLine(line)
				addBlock()
				codeMarker = ""
			} else {
				appendLine(line)
			}
			continue
		}

		if codeMarker != "" {
			appendLine(line)
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
					appendLine(line)
				}
				matched = true
				break
			}
		}

		if !matched {
			if current == "" || currentType != "paragraph" {
				startBlock("paragraph", line+"\n")
			} else {
				appendLine(line)
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

func getHeadingLevel(block string) int {
	match := regexp.MustCompile(`^(#+)\s`).FindStringSubmatch(block)
	if match != nil {
		return len(match[1]) // Number of # characters
	}
	return 0 // Not a heading
}

func mergeSmallBlocks(blocks []string, minSize int) []string {
	if len(blocks) == 0 {
		return blocks
	}

	result := []string{}
	accumulated := blocks[0]
	accumulatedHeadingLevel := getHeadingLevel(accumulated)

	for i := 1; i < len(blocks); i++ {
		nextHeadingLevel := getHeadingLevel(blocks[i])

		// Don't merge if next block has a "bigger" heading (lower level number)
		// e.g., ### (level 3) followed by ## (level 2) should start new chunk
		isBiggerHeading := nextHeadingLevel > 0 && accumulatedHeadingLevel > 0 && nextHeadingLevel < accumulatedHeadingLevel

		if len(accumulated) < minSize && !isBiggerHeading {
			accumulated += "\n" + blocks[i]
			// Update to track the smallest (most important) heading level
			if nextHeadingLevel > 0 && (accumulatedHeadingLevel == 0 || nextHeadingLevel < accumulatedHeadingLevel) {
				accumulatedHeadingLevel = nextHeadingLevel
			}
		} else {
			result = append(result, accumulated)
			accumulated = blocks[i]
			accumulatedHeadingLevel = nextHeadingLevel
		}
	}
	result = append(result, accumulated)

	return result
}
