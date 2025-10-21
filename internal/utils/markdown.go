package utils

import (
	"strings"
)

func SplitMarkdownBlocks(content string) []string {
	rawBlocks := strings.Split(content, "\n\n")

	var blocks []string
	for i := 0; i < len(rawBlocks); i++ {
		block := strings.TrimSpace(rawBlocks[i])
		if block == "" {
			continue
		}

		if isHeader(block) && i+1 < len(rawBlocks) {
			next := strings.TrimSpace(rawBlocks[i+1])
			if next != "" {
				blocks = append(blocks, block+"\n\n"+next)
				i++
				continue
			}
		}

		blocks = append(blocks, block)
	}

	return blocks
}

func CombineBlocks(blocks []string, n int) []string {
	if n <= 0 {
		n = 1
	}

	var combined []string
	for i := 0; i < len(blocks); i += n {
		end := i + n
		if end > len(blocks) {
			end = len(blocks)
		}

		chunk := strings.Join(blocks[i:end], "\n\n")
		combined = append(combined, chunk)
	}

	return combined
}

func isHeader(block string) bool {
	line := strings.Split(block, "\n")[0]
	return strings.HasPrefix(strings.TrimSpace(line), "#")
}
