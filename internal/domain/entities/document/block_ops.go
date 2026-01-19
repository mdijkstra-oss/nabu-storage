package document

import (
	"hermes-relay/internal/lib/utils"
	"strings"
)

type BlockTree struct {
	Blocks map[string]Block
	HeadID string
	TailID string
}

func NewBlockTree() BlockTree {
	return BlockTree{Blocks: make(map[string]Block)}
}

func ToBlockTree(d DocumentData) BlockTree {
	if d.Blocks == nil {
		return NewBlockTree()
	}
	return BlockTree{Blocks: d.Blocks, HeadID: d.HeadID, TailID: d.TailID}
}

func (t BlockTree) Get(id string) (Block, bool) {
	b, ok := t.Blocks[id]
	return b, ok
}

func (t BlockTree) set(b Block) BlockTree {
	newBlocks := copyBlocks(t.Blocks)
	newBlocks[b.ID] = b
	return BlockTree{Blocks: newBlocks, HeadID: t.HeadID, TailID: t.TailID}
}

func (t BlockTree) setMany(blocks []Block) BlockTree {
	newBlocks := copyBlocks(t.Blocks)
	for _, b := range blocks {
		newBlocks[b.ID] = b
	}
	return BlockTree{Blocks: newBlocks, HeadID: t.HeadID, TailID: t.TailID}
}

func (t BlockTree) remove(id string) BlockTree {
	newBlocks := copyBlocks(t.Blocks)
	delete(newBlocks, id)
	return BlockTree{Blocks: newBlocks, HeadID: t.HeadID, TailID: t.TailID}
}

func (t BlockTree) withHead(id string) BlockTree {
	return BlockTree{Blocks: t.Blocks, HeadID: id, TailID: t.TailID}
}

func (t BlockTree) withTail(id string) BlockTree {
	return BlockTree{Blocks: t.Blocks, HeadID: t.HeadID, TailID: id}
}

func copyBlocks(blocks map[string]Block) map[string]Block {
	result := make(map[string]Block, len(blocks))
	for k, v := range blocks {
		result[k] = v
	}
	return result
}

func InsertBlocksAfter(tree BlockTree, position string, newBlocks []Block) (BlockTree, bool) {
	if len(newBlocks) == 0 {
		return tree, true
	}

	existing, toInsert := partitionByExistence(tree, newBlocks)
	tree = updateBlocksInPlace(tree, existing)

	if len(toInsert) == 0 {
		return tree, true
	}

	toInsert = assignBlockIDs(toInsert)
	op, parentID := utils.ParseBlockPosition(position)

	if parentID == "" {
		if op == PositionHead {
			return insertAtHead(tree, toInsert), true
		}
		return insertAtTail(tree, toInsert), true
	}

	if op == PositionHead {
		return insertAsFirstChild(tree, parentID, toInsert)
	}
	if op == PositionTail {
		return insertAsLastChild(tree, parentID, toInsert)
	}
	return insertAfterID(tree, parentID, toInsert)
}

func partitionByExistence(tree BlockTree, newBlocks []Block) (existing, toInsert []Block) {
	for _, nb := range newBlocks {
		if _, found := tree.Get(nb.ID); found {
			existing = append(existing, nb)
		} else {
			toInsert = append(toInsert, nb)
		}
	}
	return
}

func updateBlocksInPlace(tree BlockTree, updates []Block) BlockTree {
	if len(updates) == 0 {
		return tree
	}
	for _, u := range updates {
		if existing, ok := tree.Get(u.ID); ok {
			updated := u
			updated.NextID = existing.NextID
			updated.PrevID = existing.PrevID
			updated.FirstChildID = existing.FirstChildID
			updated.LastChildID = existing.LastChildID
			updated.ParentID = existing.ParentID
			tree = tree.set(updated)
		}
	}
	return tree
}

func assignBlockIDs(blocks []Block) []Block {
	result := make([]Block, len(blocks))
	for i, b := range blocks {
		if b.ID == "" {
			b.ID = utils.NewBlockID()
		}
		result[i] = b
	}
	return result
}

func linkChain(blocks []Block) []Block {
	if len(blocks) == 0 {
		return blocks
	}
	result := make([]Block, len(blocks))
	for i, b := range blocks {
		if i > 0 {
			b.PrevID = blocks[i-1].ID
		}
		if i < len(blocks)-1 {
			b.NextID = blocks[i+1].ID
		}
		result[i] = b
	}
	return result
}

func insertAtHead(tree BlockTree, newBlocks []Block) BlockTree {
	newBlocks = linkChain(newBlocks)
	tree = tree.setMany(newBlocks)

	firstNew := newBlocks[0]
	lastNew := newBlocks[len(newBlocks)-1]

	if tree.HeadID != "" {
		oldHead, _ := tree.Get(tree.HeadID)
		oldHead.PrevID = lastNew.ID
		tree = tree.set(oldHead)

		lastNew.NextID = tree.HeadID
		tree = tree.set(lastNew)
	} else {
		tree = tree.withTail(lastNew.ID)
	}

	return tree.withHead(firstNew.ID)
}

func insertAtTail(tree BlockTree, newBlocks []Block) BlockTree {
	newBlocks = linkChain(newBlocks)
	tree = tree.setMany(newBlocks)

	firstNew := newBlocks[0]
	lastNew := newBlocks[len(newBlocks)-1]

	if tree.TailID != "" {
		oldTail, _ := tree.Get(tree.TailID)
		oldTail.NextID = firstNew.ID
		tree = tree.set(oldTail)

		firstNew.PrevID = tree.TailID
		tree = tree.set(firstNew)
	} else {
		tree = tree.withHead(firstNew.ID)
	}

	return tree.withTail(lastNew.ID)
}

func insertAfterID(tree BlockTree, blockID string, newBlocks []Block) (BlockTree, bool) {
	target, ok := tree.Get(blockID)
	if !ok {
		return tree, false
	}

	newBlocks = linkChain(newBlocks)

	for i := range newBlocks {
		newBlocks[i].ParentID = target.ParentID
	}

	newBlocks[0].PrevID = target.ID

	oldNextID := target.NextID
	if oldNextID != "" {
		newBlocks[len(newBlocks)-1].NextID = oldNextID
	}

	tree = tree.setMany(newBlocks)

	target.NextID = newBlocks[0].ID
	tree = tree.set(target)

	if oldNextID != "" {
		next, _ := tree.Get(oldNextID)
		next.PrevID = newBlocks[len(newBlocks)-1].ID
		tree = tree.set(next)
	} else {
		if target.ParentID != "" {
			parent, _ := tree.Get(target.ParentID)
			parent.LastChildID = newBlocks[len(newBlocks)-1].ID
			tree = tree.set(parent)
		} else {
			tree = tree.withTail(newBlocks[len(newBlocks)-1].ID)
		}
	}

	return tree, true
}

func insertAsFirstChild(tree BlockTree, parentID string, newBlocks []Block) (BlockTree, bool) {
	parent, ok := tree.Get(parentID)
	if !ok {
		return tree, false
	}

	newBlocks = linkChain(newBlocks)
	for i := range newBlocks {
		newBlocks[i].ParentID = parentID
	}
	tree = tree.setMany(newBlocks)

	firstNew := newBlocks[0]
	lastNew := newBlocks[len(newBlocks)-1]

	if parent.FirstChildID != "" {
		oldFirst, _ := tree.Get(parent.FirstChildID)
		oldFirst.PrevID = lastNew.ID
		tree = tree.set(oldFirst)

		lastNew.NextID = parent.FirstChildID
		tree = tree.set(lastNew)
	} else {
		parent.LastChildID = lastNew.ID
	}

	parent.FirstChildID = firstNew.ID
	tree = tree.set(parent)

	return tree, true
}

func insertAsLastChild(tree BlockTree, parentID string, newBlocks []Block) (BlockTree, bool) {
	parent, ok := tree.Get(parentID)
	if !ok {
		return tree, false
	}

	newBlocks = linkChain(newBlocks)
	for i := range newBlocks {
		newBlocks[i].ParentID = parentID
	}
	tree = tree.setMany(newBlocks)

	firstNew := newBlocks[0]
	lastNew := newBlocks[len(newBlocks)-1]

	if parent.LastChildID != "" {
		oldLast, _ := tree.Get(parent.LastChildID)
		oldLast.NextID = firstNew.ID
		tree = tree.set(oldLast)

		firstNew.PrevID = parent.LastChildID
		tree = tree.set(firstNew)
	} else {
		parent.FirstChildID = firstNew.ID
	}

	parent.LastChildID = lastNew.ID
	tree = tree.set(parent)

	return tree, true
}

func DeleteBlocksByID(tree BlockTree, ids []string) BlockTree {
	idSet := toSet(ids)
	for id := range idSet {
		tree = unlinkAndRemove(tree, id)
	}
	return tree
}

func unlinkAndRemove(tree BlockTree, id string) BlockTree {
	block, ok := tree.Get(id)
	if !ok {
		return tree
	}

	childID := block.FirstChildID
	for childID != "" {
		child, _ := tree.Get(childID)
		nextChildID := child.NextID
		tree = unlinkAndRemove(tree, childID)
		childID = nextChildID
	}

	if block.PrevID != "" {
		prev, _ := tree.Get(block.PrevID)
		prev.NextID = block.NextID
		tree = tree.set(prev)
	} else if block.ParentID != "" {
		parent, _ := tree.Get(block.ParentID)
		parent.FirstChildID = block.NextID
		tree = tree.set(parent)
	} else if tree.HeadID == id {
		tree = tree.withHead(block.NextID)
	}

	if block.NextID != "" {
		next, _ := tree.Get(block.NextID)
		next.PrevID = block.PrevID
		tree = tree.set(next)
	} else if block.ParentID != "" {
		parent, _ := tree.Get(block.ParentID)
		parent.LastChildID = block.PrevID
		tree = tree.set(parent)
	} else if tree.TailID == id {
		tree = tree.withTail(block.PrevID)
	}

	return tree.remove(id)
}

func toSet(ids []string) map[string]bool {
	set := make(map[string]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set
}

func FindBlock(tree BlockTree, id string) (Block, bool) {
	return tree.Get(id)
}

func FindBlockDepth(tree BlockTree, id string) int {
	block, ok := tree.Get(id)
	if !ok {
		return -1
	}
	depth := 0
	for block.ParentID != "" {
		depth++
		block, _ = tree.Get(block.ParentID)
	}
	return depth
}

func MoveBlocksAfter(tree BlockTree, ids []string, position string) BlockTree {
	moving := ExtractBlocksByID(tree, ids)
	tree = DeleteBlocksByID(tree, ids)
	tree, _ = InsertBlocksAfter(tree, position, moving)
	return tree
}

func ExtractBlocksByID(tree BlockTree, ids []string) []Block {
	idSet := toSet(ids)
	var result []Block
	for id := range idSet {
		if b, ok := tree.Get(id); ok {
			result = append(result, b)
		}
	}
	return result
}

func ApplyBlockUpdate(tree BlockTree, update *UpdateBlockPayload) BlockTree {
	block, ok := tree.Get(update.BlockID)
	if !ok {
		return tree
	}

	if update.Type != nil {
		block.Type = *update.Type
	}
	if update.Props != nil {
		if update.Props.BackgroundColor != nil {
			block.Props.BackgroundColor = *update.Props.BackgroundColor
		}
		if update.Props.Level != nil {
			block.Props.Level = *update.Props.Level
		}
		if update.Props.Checked != nil {
			block.Props.Checked = update.Props.Checked
		}
	}
	if update.Content != nil {
		block.Content = update.Content
	}

	return tree.set(block)
}

func ToArray(tree BlockTree) []Block {
	if tree.HeadID == "" {
		return []Block{}
	}
	return collectSiblings(tree, tree.HeadID)
}

func collectSiblings(tree BlockTree, startID string) []Block {
	var result []Block
	currentID := startID
	for currentID != "" {
		block, ok := tree.Get(currentID)
		if !ok {
			break
		}
		blockCopy := block
		blockCopy.NextID = ""
		blockCopy.PrevID = ""
		blockCopy.ParentID = ""
		blockCopy.FirstChildID = ""
		blockCopy.LastChildID = ""
		blockCopy.Children = nil

		result = append(result, blockCopy)
		currentID = block.NextID
	}
	return result
}

func FromArray(blocks []Block) BlockTree {
	tree := NewBlockTree()
	if len(blocks) == 0 {
		return tree
	}
	return insertNestedBlocks(tree, PositionHead, blocks)
}

func insertNestedBlocks(tree BlockTree, position string, blocks []Block) BlockTree {
	if len(blocks) == 0 {
		return tree
	}

	flatBlocks := make([]Block, len(blocks))
	for i, b := range blocks {
		flat := b
		flat.Children = nil
		flatBlocks[i] = flat
	}

	tree, _ = InsertBlocksAfter(tree, position, flatBlocks)

	for _, b := range blocks {
		if len(b.Children) > 0 {
			tree = insertNestedBlocks(tree, "head:"+b.ID, b.Children)
		}
	}

	return tree
}

func ExtractBlockText(block Block) string {
	var parts []string
	for _, inline := range block.Content {
		parts = append(parts, extractInlineText(inline))
	}
	return strings.Join(parts, "")
}

func extractInlineText(inline InlineContent) string {
	if len(inline.Content) > 0 {
		var parts []string
		for _, styled := range inline.Content {
			parts = append(parts, styled.Text)
		}
		return strings.Join(parts, "")
	}
	return inline.Text
}

func ExtractDocumentText(tree BlockTree) string {
	blocks := ToArray(tree)
	var parts []string
	for _, block := range blocks {
		parts = append(parts, ExtractBlockText(block))
	}
	return strings.Join(parts, "\n")
}
