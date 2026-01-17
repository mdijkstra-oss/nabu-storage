package document

import (
	"testing"

	th "hermes-relay/internal/lib/test-helpers"
)

func block(id string) Block {
	return Block{ID: id, Type: BlockTypeParagraph}
}

func blockWithContent(id string, text string) Block {
	return Block{ID: id, Type: BlockTypeParagraph, Content: []InlineContent{{Type: InlineTypeText, Text: text}}}
}

func tree(blocks ...Block) BlockTree {
	return FromArray(blocks)
}

func orderedIDs(t BlockTree) []string {
	blocks := ToArray(t)
	ids := make([]string, len(blocks))
	for i, b := range blocks {
		ids[i] = b.ID
	}
	return ids
}

func childIDs(t BlockTree, parentID string) []string {
	parent, ok := t.Get(parentID)
	if !ok || parent.FirstChildID == "" {
		return nil
	}
	var ids []string
	childID := parent.FirstChildID
	for childID != "" {
		child, _ := t.Get(childID)
		ids = append(ids, child.ID)
		childID = child.NextID
	}
	return ids
}

func treeWithChild(parentID string, childIDs ...string) BlockTree {
	t := NewBlockTree()
	parent := block(parentID)
	t = t.set(parent)
	t = t.withHead(parentID)
	t = t.withTail(parentID)

	for i, cid := range childIDs {
		child := block(cid)
		child.ParentID = parentID
		if i > 0 {
			child.PrevID = childIDs[i-1]
		}
		if i < len(childIDs)-1 {
			child.NextID = childIDs[i+1]
		}
		t = t.set(child)
	}

	if len(childIDs) > 0 {
		parent.FirstChildID = childIDs[0]
		parent.LastChildID = childIDs[len(childIDs)-1]
		t = t.set(parent)
	}

	return t
}

func TestInsertBlocksAfter(t *testing.T) {
	tests := []struct {
		name     string
		tree     BlockTree
		pos      string
		add      []Block
		wantRoot []string
		notFound bool
	}{
		{"at head", tree(block("a"), block("b")), PositionHead, []Block{block("x")}, []string{"x", "a", "b"}, false},
		{"at tail", tree(block("a"), block("b")), PositionTail, []Block{block("x")}, []string{"a", "b", "x"}, false},
		{"after first", tree(block("a"), block("b"), block("c")), "a", []Block{block("x")}, []string{"a", "x", "b", "c"}, false},
		{"after last", tree(block("a"), block("b")), "b", []Block{block("x")}, []string{"a", "b", "x"}, false},
		{"multiple", tree(block("a"), block("b")), "a", []Block{block("x"), block("y")}, []string{"a", "x", "y", "b"}, false},
		{"not found", tree(block("a"), block("b")), "z", []Block{block("x")}, []string{"a", "b"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, found := InsertBlocksAfter(tt.tree, tt.pos, tt.add)
			th.AssertEqual(t, found, !tt.notFound, "found")
			th.AssertEqual(t, orderedIDs(result), tt.wantRoot, "root ids")
		})
	}
}

func TestInsertBlocksAfterNested(t *testing.T) {
	tests := []struct {
		name         string
		tree         BlockTree
		pos          string
		add          []Block
		wantRoot     []string
		wantChildren []string
		parentID     string
		notFound     bool
	}{
		{
			name:         "head:parent inserts as first child",
			tree:         treeWithChild("a", "a1", "a2"),
			pos:          "head:a",
			add:          []Block{block("x")},
			wantRoot:     []string{"a"},
			wantChildren: []string{"x", "a1", "a2"},
			parentID:     "a",
		},
		{
			name:         "tail:parent inserts as last child",
			tree:         treeWithChild("a", "a1", "a2"),
			pos:          "tail:a",
			add:          []Block{block("x")},
			wantRoot:     []string{"a"},
			wantChildren: []string{"a1", "a2", "x"},
			parentID:     "a",
		},
		{
			name:         "after sibling inserts between",
			tree:         treeWithChild("a", "a1", "a2"),
			pos:          "a1",
			add:          []Block{block("x")},
			wantRoot:     []string{"a"},
			wantChildren: []string{"a1", "x", "a2"},
			parentID:     "a",
		},
		{
			name:         "head:empty parent",
			tree:         tree(block("a"), block("b")),
			pos:          "head:a",
			add:          []Block{block("x")},
			wantRoot:     []string{"a", "b"},
			wantChildren: []string{"x"},
			parentID:     "a",
		},
		{
			name:     "head:not found",
			tree:     tree(block("a"), block("b")),
			pos:      "head:z",
			add:      []Block{block("x")},
			wantRoot: []string{"a", "b"},
			notFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, found := InsertBlocksAfter(tt.tree, tt.pos, tt.add)
			th.AssertEqual(t, found, !tt.notFound, "found")
			th.AssertEqual(t, orderedIDs(result), tt.wantRoot, "root ids")
			if tt.wantChildren != nil {
				th.AssertEqual(t, childIDs(result, tt.parentID), tt.wantChildren, "children ids")
			}
		})
	}
}

func TestInsertBlocksAfterUpsert(t *testing.T) {
	tests := []struct {
		name        string
		tree        BlockTree
		pos         string
		add         []Block
		wantIDs     []string
		wantContent map[string]string
	}{
		{
			name:    "upsert existing block updates in place",
			tree:    tree(blockWithContent("a", "old"), blockWithContent("b", "old")),
			pos:     PositionTail,
			add:     []Block{blockWithContent("a", "new")},
			wantIDs: []string{"a", "b"},
			wantContent: map[string]string{
				"a": "new",
				"b": "old",
			},
		},
		{
			name:    "upsert mixed existing and new",
			tree:    tree(blockWithContent("a", "old"), blockWithContent("b", "old")),
			pos:     PositionTail,
			add:     []Block{blockWithContent("a", "new"), blockWithContent("x", "new")},
			wantIDs: []string{"a", "b", "x"},
			wantContent: map[string]string{
				"a": "new",
				"b": "old",
				"x": "new",
			},
		},
		{
			name:    "upsert all existing no position change",
			tree:    tree(blockWithContent("a", "old"), blockWithContent("b", "old"), blockWithContent("c", "old")),
			pos:     PositionHead,
			add:     []Block{blockWithContent("b", "new"), blockWithContent("c", "new")},
			wantIDs: []string{"a", "b", "c"},
			wantContent: map[string]string{
				"a": "old",
				"b": "new",
				"c": "new",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := InsertBlocksAfter(tt.tree, tt.pos, tt.add)
			th.AssertEqual(t, orderedIDs(result), tt.wantIDs, "ids")

			for id, wantText := range tt.wantContent {
				b, found := result.Get(id)
				th.AssertEqual(t, found, true, "block "+id+" found")
				gotText := ""
				if len(b.Content) > 0 {
					gotText = b.Content[0].Text
				}
				th.AssertEqual(t, gotText, wantText, "block "+id+" content")
			}
		})
	}
}

func TestDeleteBlocksByID(t *testing.T) {
	tests := []struct {
		name         string
		tree         BlockTree
		del          []string
		wantRoot     []string
		wantChildren []string
		parentID     string
	}{
		{"single", tree(block("a"), block("b"), block("c")), []string{"b"}, []string{"a", "c"}, nil, ""},
		{"multiple", tree(block("a"), block("b"), block("c")), []string{"a", "c"}, []string{"b"}, nil, ""},
		{"nested child", treeWithChild("a", "a1", "a2"), []string{"a1"}, []string{"a"}, []string{"a2"}, "a"},
		{"parent deletes children", treeWithChild("a", "a1", "a2"), []string{"a"}, []string{}, nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DeleteBlocksByID(tt.tree, tt.del)
			th.AssertEqual(t, orderedIDs(result), tt.wantRoot, "ids")
			if tt.wantChildren != nil {
				th.AssertEqual(t, childIDs(result, tt.parentID), tt.wantChildren, "children")
			}
		})
	}
}

func TestFindBlockDepth(t *testing.T) {
	deepTree := treeWithChild("a", "a1")
	a1, _ := deepTree.Get("a1")
	a1.FirstChildID = "a1a"
	a1.LastChildID = "a1a"
	deepTree = deepTree.set(a1)
	deepChild := block("a1a")
	deepChild.ParentID = "a1"
	deepTree = deepTree.set(deepChild)

	th.RunMapTests(t, map[string]int{
		"a":   0,
		"a1":  1,
		"a1a": 2,
		"z":   -1,
	}, func(id string) int {
		return FindBlockDepth(deepTree, id)
	})
}

func TestMoveBlocksAfter(t *testing.T) {
	result := MoveBlocksAfter(tree(block("a"), block("b"), block("c"), block("d")), []string{"b"}, "c")
	th.AssertEqual(t, orderedIDs(result), []string{"a", "c", "b", "d"}, "ids")
}

func TestExtractBlocksByID(t *testing.T) {
	tr := treeWithChild("a", "a1")
	tr, _ = InsertBlocksAfter(tr, "a", []Block{block("b"), block("c")})

	result := ExtractBlocksByID(tr, []string{"b", "a1"})
	th.AssertEqual(t, len(result), 2, "count")
}

func TestReplaceBlocksByID(t *testing.T) {
	result := ReplaceBlocksByID(tree(block("a"), block("b"), block("c")), []string{"b"}, []Block{block("x"), block("y")})
	th.AssertEqual(t, orderedIDs(result), []string{"a", "x", "y", "c"}, "ids")
}

func TestFindBlock(t *testing.T) {
	deepTree := treeWithChild("a", "a1")
	a1, _ := deepTree.Get("a1")
	a1.FirstChildID = "a1a"
	a1.LastChildID = "a1a"
	deepTree = deepTree.set(a1)
	deepChild := block("a1a")
	deepChild.ParentID = "a1"
	deepTree = deepTree.set(deepChild)
	deepTree, _ = InsertBlocksAfter(deepTree, "a", []Block{block("b")})

	th.RunMapTests(t, map[string]bool{
		"a": true, "a1": true, "a1a": true, "b": true, "z": false,
	}, func(id string) bool {
		_, found := FindBlock(deepTree, id)
		return found
	})
}

func TestUpdateBlocksProps(t *testing.T) {
	checked, unchecked := true, false

	tests := []struct {
		name  string
		tree  BlockTree
		ids   []string
		props BlockProps
		check func(t *testing.T, result BlockTree)
	}{
		{
			"toggle checked",
			tree(
				Block{ID: "a", Type: BlockTypeCheckList, Props: BlockProps{CheckListProps: CheckListProps{Checked: &checked}}},
				Block{ID: "b", Type: BlockTypeCheckList},
				Block{ID: "c", Type: BlockTypeParagraph},
			),
			[]string{"a", "b"},
			BlockProps{CheckListProps: CheckListProps{Checked: &unchecked}},
			func(t *testing.T, result BlockTree) {
				a, _ := result.Get("a")
				b, _ := result.Get("b")
				c, _ := result.Get("c")
				th.AssertEqual(t, *a.Props.Checked, false, "a.checked")
				th.AssertEqual(t, *b.Props.Checked, false, "b.checked")
				th.AssertEqual(t, c.Props.Checked, (*bool)(nil), "c.checked")
			},
		},
		{
			"nested heading level",
			func() BlockTree {
				tr := treeWithChild("a", "a1")
				a1, _ := tr.Get("a1")
				a1.Type = BlockTypeHeading
				a1.Props = BlockProps{HeadingProps: HeadingProps{Level: 1}}
				return tr.set(a1)
			}(),
			[]string{"a1"},
			BlockProps{HeadingProps: HeadingProps{Level: 2}},
			func(t *testing.T, result BlockTree) {
				a1, _ := result.Get("a1")
				th.AssertEqual(t, a1.Props.Level, 2, "level")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UpdateBlocksProps(tt.tree, tt.ids, tt.props)
			tt.check(t, result)
		})
	}
}

func TestExtractBlockText(t *testing.T) {
	tests := []struct {
		name  string
		block Block
		want  string
	}{
		{
			name:  "empty block",
			block: Block{ID: "a", Type: BlockTypeParagraph},
			want:  "",
		},
		{
			name:  "single text inline",
			block: Block{ID: "a", Type: BlockTypeParagraph, Content: []InlineContent{{Type: InlineTypeText, Text: "hello"}}},
			want:  "hello",
		},
		{
			name: "multiple text inlines",
			block: Block{ID: "a", Type: BlockTypeParagraph, Content: []InlineContent{
				{Type: InlineTypeText, Text: "hello "},
				{Type: InlineTypeText, Text: "world"},
			}},
			want: "hello world",
		},
		{
			name: "link inline extracts styled text",
			block: Block{ID: "a", Type: BlockTypeParagraph, Content: []InlineContent{
				{Type: InlineTypeLink, Href: "http://example.com", Content: []StyledText{{Text: "click here"}}},
			}},
			want: "click here",
		},
		{
			name: "mixed text and link",
			block: Block{ID: "a", Type: BlockTypeParagraph, Content: []InlineContent{
				{Type: InlineTypeText, Text: "see "},
				{Type: InlineTypeLink, Href: "http://example.com", Content: []StyledText{{Text: "this link"}}},
				{Type: InlineTypeText, Text: " for details"},
			}},
			want: "see this link for details",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractBlockText(tt.block)
			th.AssertEqual(t, got, tt.want, "text")
		})
	}
}

func TestExtractDocumentText(t *testing.T) {
	tests := []struct {
		name string
		tree BlockTree
		want string
	}{
		{
			name: "empty document",
			tree: NewBlockTree(),
			want: "",
		},
		{
			name: "single block",
			tree: tree(Block{ID: "a", Type: BlockTypeParagraph, Content: []InlineContent{{Type: InlineTypeText, Text: "hello"}}}),
			want: "hello",
		},
		{
			name: "multiple blocks joined with newline",
			tree: tree(
				Block{ID: "a", Type: BlockTypeParagraph, Content: []InlineContent{{Type: InlineTypeText, Text: "first"}}},
				Block{ID: "b", Type: BlockTypeParagraph, Content: []InlineContent{{Type: InlineTypeText, Text: "second"}}},
			),
			want: "first\nsecond",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractDocumentText(tt.tree)
			th.AssertEqual(t, got, tt.want, "text")
		})
	}
}

func TestFromArrayWithNestedChildren(t *testing.T) {
	blocks := []Block{
		{
			ID:   "a",
			Type: BlockTypeParagraph,
			Children: []Block{
				{ID: "a1", Type: BlockTypeParagraph},
				{ID: "a2", Type: BlockTypeParagraph},
			},
		},
		{ID: "b", Type: BlockTypeParagraph},
	}

	tree := FromArray(blocks)

	th.AssertEqual(t, orderedIDs(tree), []string{"a", "b"}, "root ids")
	th.AssertEqual(t, childIDs(tree, "a"), []string{"a1", "a2"}, "children of a")

	a1, _ := tree.Get("a1")
	th.AssertEqual(t, a1.ParentID, "a", "a1 parent")

	a2, _ := tree.Get("a2")
	th.AssertEqual(t, a2.ParentID, "a", "a2 parent")
}

func TestFromArrayDeeplyNested(t *testing.T) {
	blocks := []Block{
		{
			ID:   "a",
			Type: BlockTypeParagraph,
			Children: []Block{
				{
					ID:   "a1",
					Type: BlockTypeParagraph,
					Children: []Block{
						{ID: "a1a", Type: BlockTypeParagraph},
					},
				},
			},
		},
	}

	tree := FromArray(blocks)

	th.AssertEqual(t, orderedIDs(tree), []string{"a"}, "root")
	th.AssertEqual(t, childIDs(tree, "a"), []string{"a1"}, "children of a")
	th.AssertEqual(t, childIDs(tree, "a1"), []string{"a1a"}, "children of a1")

	a1a, _ := tree.Get("a1a")
	th.AssertEqual(t, a1a.ParentID, "a1", "a1a parent")
}
