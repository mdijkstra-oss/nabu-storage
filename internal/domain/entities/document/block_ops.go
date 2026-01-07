package document

import "hermes-relay/internal/lib/utils"

func InsertBlocksAfter(blocks []Block, position string, newBlocks []Block) ([]Block, bool) {
	// Upsert: update existing blocks in place, only insert truly new ones
	existing, toInsert := partitionByExistence(blocks, newBlocks)
	blocks = updateBlocksInPlace(blocks, existing)

	if len(toInsert) == 0 {
		return blocks, true
	}

	op, parentID := utils.ParseBlockPosition(position)

	if parentID == "" {
		if op == PositionHead {
			return insertAtHead(blocks, toInsert), true
		}
		return insertAtTail(blocks, toInsert), true
	}

	if op == PositionHead {
		return insertAsFirstChild(blocks, parentID, toInsert)
	}
	if op == PositionTail {
		return insertAsLastChild(blocks, parentID, toInsert)
	}
	return insertAfterID(blocks, parentID, toInsert)
}

func partitionByExistence(tree []Block, newBlocks []Block) (existing, toInsert []Block) {
	for _, nb := range newBlocks {
		if _, found := FindBlock(tree, nb.ID); found {
			existing = append(existing, nb)
		} else {
			toInsert = append(toInsert, nb)
		}
	}
	return
}

func updateBlocksInPlace(blocks []Block, updates []Block) []Block {
	if len(updates) == 0 {
		return blocks
	}
	updateMap := make(map[string]Block, len(updates))
	for _, u := range updates {
		updateMap[u.ID] = u
	}
	return updateInTree(blocks, updateMap)
}

func updateInTree(blocks []Block, updates map[string]Block) []Block {
	result := make([]Block, len(blocks))
	for i, b := range blocks {
		if update, found := updates[b.ID]; found {
			result[i] = update
		} else if len(b.Children) > 0 {
			updated := b
			updated.Children = updateInTree(b.Children, updates)
			result[i] = updated
		} else {
			result[i] = b
		}
	}
	return result
}

func insertAtHead(blocks []Block, newBlocks []Block) []Block {
	result := make([]Block, 0, len(blocks)+len(newBlocks))
	result = append(result, newBlocks...)
	result = append(result, blocks...)
	return result
}

func insertAtTail(blocks []Block, newBlocks []Block) []Block {
	result := make([]Block, 0, len(blocks)+len(newBlocks))
	result = append(result, blocks...)
	result = append(result, newBlocks...)
	return result
}

func insertAsFirstChild(blocks []Block, parentID string, newBlocks []Block) ([]Block, bool) {
	for i, b := range blocks {
		if b.ID == parentID {
			newChildren := insertAtHead(b.Children, newBlocks)
			return replaceChildrenAtIndex(blocks, i, newChildren), true
		}
		if len(b.Children) > 0 {
			if children, found := insertAsFirstChild(b.Children, parentID, newBlocks); found {
				return replaceChildrenAtIndex(blocks, i, children), true
			}
		}
	}
	return blocks, false
}

func insertAsLastChild(blocks []Block, parentID string, newBlocks []Block) ([]Block, bool) {
	for i, b := range blocks {
		if b.ID == parentID {
			newChildren := insertAtTail(b.Children, newBlocks)
			return replaceChildrenAtIndex(blocks, i, newChildren), true
		}
		if len(b.Children) > 0 {
			if children, found := insertAsLastChild(b.Children, parentID, newBlocks); found {
				return replaceChildrenAtIndex(blocks, i, children), true
			}
		}
	}
	return blocks, false
}

func insertAfterID(blocks []Block, blockID string, newBlocks []Block) ([]Block, bool) {
	for i, b := range blocks {
		if b.ID == blockID {
			return insertAtIndex(blocks, i+1, newBlocks), true
		}
		if len(b.Children) > 0 {
			if children, found := insertAfterID(b.Children, blockID, newBlocks); found {
				return replaceChildrenAtIndex(blocks, i, children), true
			}
		}
	}
	return blocks, false
}

func insertAtIndex(blocks []Block, index int, newBlocks []Block) []Block {
	result := make([]Block, 0, len(blocks)+len(newBlocks))
	result = append(result, blocks[:index]...)
	result = append(result, newBlocks...)
	result = append(result, blocks[index:]...)
	return result
}

func replaceChildrenAtIndex(blocks []Block, index int, newChildren []Block) []Block {
	result := make([]Block, len(blocks))
	copy(result, blocks)
	updated := result[index]
	updated.Children = newChildren
	result[index] = updated
	return result
}

func DeleteBlocksByID(blocks []Block, ids []string) []Block {
	idSet := toSet(ids)
	return deleteFromTree(blocks, idSet)
}

func deleteFromTree(blocks []Block, ids map[string]bool) []Block {
	result := make([]Block, 0, len(blocks))
	for _, b := range blocks {
		if ids[b.ID] {
			continue
		}
		if len(b.Children) > 0 {
			updated := b
			updated.Children = deleteFromTree(b.Children, ids)
			result = append(result, updated)
		} else {
			result = append(result, b)
		}
	}
	return result
}

func toSet(ids []string) map[string]bool {
	set := make(map[string]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set
}

func FindBlockDepth(blocks []Block, id string) int {
	return findDepth(blocks, id, 0)
}

func findDepth(blocks []Block, id string, depth int) int {
	for _, b := range blocks {
		if b.ID == id {
			return depth
		}
		if len(b.Children) > 0 {
			if found := findDepth(b.Children, id, depth+1); found >= 0 {
				return found
			}
		}
	}
	return -1
}

func MoveBlocksAfter(blocks []Block, ids []string, position string) []Block {
	moving := ExtractBlocksByID(blocks, ids)
	remaining := DeleteBlocksByID(blocks, ids)
	result, _ := InsertBlocksAfter(remaining, position, moving)
	return result
}

func ExtractBlocksByID(blocks []Block, ids []string) []Block {
	idSet := toSet(ids)
	return extractFromTree(blocks, idSet)
}

func extractFromTree(blocks []Block, ids map[string]bool) []Block {
	var result []Block
	for _, b := range blocks {
		if ids[b.ID] {
			result = append(result, b)
		}
		if len(b.Children) > 0 {
			result = append(result, extractFromTree(b.Children, ids)...)
		}
	}
	return result
}

func ReplaceBlocksByID(blocks []Block, ids []string, newBlocks []Block) []Block {
	idSet := toSet(ids)
	return replaceInTree(blocks, idSet, newBlocks)
}

func FindBlock(blocks []Block, id string) (Block, bool) {
	for _, b := range blocks {
		if b.ID == id {
			return b, true
		}
		if len(b.Children) > 0 {
			if found, ok := FindBlock(b.Children, id); ok {
				return found, true
			}
		}
	}
	return Block{}, false
}

func UpdateBlocksProps(blocks []Block, ids []string, props BlockProps) []Block {
	idSet := toSet(ids)
	return updatePropsInTree(blocks, idSet, props)
}

func updatePropsInTree(blocks []Block, ids map[string]bool, props BlockProps) []Block {
	result := make([]Block, len(blocks))
	for i, b := range blocks {
		if ids[b.ID] {
			updated := b
			updated.Props = utils.ApplyPartialUpdate(b.Props, props)
			result[i] = updated
		} else if len(b.Children) > 0 {
			updated := b
			updated.Children = updatePropsInTree(b.Children, ids, props)
			result[i] = updated
		} else {
			result[i] = b
		}
	}
	return result
}

func replaceInTree(blocks []Block, ids map[string]bool, newBlocks []Block) []Block {
	result := make([]Block, 0, len(blocks))
	replaced := false
	for _, b := range blocks {
		if ids[b.ID] {
			if !replaced {
				result = append(result, newBlocks...)
				replaced = true
			}
			continue
		}
		if len(b.Children) > 0 {
			updated := b
			updated.Children = replaceInTree(b.Children, ids, newBlocks)
			result = append(result, updated)
		} else {
			result = append(result, b)
		}
	}
	return result
}
