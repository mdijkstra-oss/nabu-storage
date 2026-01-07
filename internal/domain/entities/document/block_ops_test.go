package document

import (
	"testing"

	th "hermes-relay/internal/lib/test-helpers"
)

func b(id string, children ...Block) Block {
	return Block{ID: id, Type: BlockTypeParagraph, Children: children}
}

func bs(ids ...string) []Block {
	result := make([]Block, len(ids))
	for i, id := range ids {
		result[i] = b(id)
	}
	return result
}

func blockIDs(blocks []Block) []string {
	result := make([]string, len(blocks))
	for i, b := range blocks {
		result[i] = b.ID
	}
	return result
}

func bWithContent(id string, content string) Block {
	return Block{ID: id, Type: BlockTypeParagraph, Content: []InlineContent{{Type: InlineTypeText, Text: content}}}
}

func TestInsertBlocksAfter(t *testing.T) {
	tests := []struct {
		name         string
		blocks       []Block
		pos          string
		add          []Block
		want         []string
		wantChildren []string
		notFound     bool
	}{
		{"at head", bs("a", "b"), PositionHead, bs("x"), []string{"x", "a", "b"}, nil, false},
		{"at tail", bs("a", "b"), PositionTail, bs("x"), []string{"a", "b", "x"}, nil, false},
		{"after first", bs("a", "b", "c"), "a", bs("x"), []string{"a", "x", "b", "c"}, nil, false},
		{"after last", bs("a", "b"), "b", bs("x"), []string{"a", "b", "x"}, nil, false},
		{"multiple", bs("a", "b"), "a", bs("x", "y"), []string{"a", "x", "y", "b"}, nil, false},
		{"not found", bs("a", "b"), "z", bs("x"), []string{"a", "b"}, nil, true},
		{"after nested", []Block{b("a", b("a1"), b("a2")), b("b")}, "a1", bs("x"), []string{"a", "b"}, []string{"a1", "x", "a2"}, false},
		{"deeply nested", []Block{b("a", b("a1", b("a1a")))}, "a1a", bs("x"), nil, nil, false},
		{"head:parent", []Block{b("a", b("a1"), b("a2")), b("b")}, "head:a", bs("x"), []string{"a", "b"}, []string{"x", "a1", "a2"}, false},
		{"tail:parent", []Block{b("a", b("a1"), b("a2")), b("b")}, "tail:a", bs("x"), []string{"a", "b"}, []string{"a1", "a2", "x"}, false},
		{"head:empty", bs("a", "b"), "head:a", bs("x"), []string{"a", "b"}, []string{"x"}, false},
		{"head:not found", bs("a", "b"), "head:z", bs("x"), []string{"a", "b"}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, found := InsertBlocksAfter(tt.blocks, tt.pos, tt.add)
			th.AssertEqual(t, found, !tt.notFound, "found")
			if tt.want != nil {
				th.AssertEqual(t, blockIDs(result), tt.want, "root ids")
			}
			if tt.wantChildren != nil {
				th.AssertEqual(t, blockIDs(result[0].Children), tt.wantChildren, "children ids")
			}
		})
	}
}

func TestInsertBlocksAfterUpsert(t *testing.T) {
	tests := []struct {
		name        string
		blocks      []Block
		pos         string
		add         []Block
		wantIDs     []string
		wantContent map[string]string
	}{
		{
			name:    "upsert existing block updates in place",
			blocks:  []Block{bWithContent("a", "old"), bWithContent("b", "old")},
			pos:     PositionTail,
			add:     []Block{bWithContent("a", "new")},
			wantIDs: []string{"a", "b"},
			wantContent: map[string]string{
				"a": "new",
				"b": "old",
			},
		},
		{
			name:    "upsert mixed existing and new",
			blocks:  []Block{bWithContent("a", "old"), bWithContent("b", "old")},
			pos:     PositionTail,
			add:     []Block{bWithContent("a", "new"), bWithContent("x", "new")},
			wantIDs: []string{"a", "b", "x"},
			wantContent: map[string]string{
				"a": "new",
				"b": "old",
				"x": "new",
			},
		},
		{
			name:    "upsert all existing no position change",
			blocks:  []Block{bWithContent("a", "old"), bWithContent("b", "old"), bWithContent("c", "old")},
			pos:     PositionHead,
			add:     []Block{bWithContent("b", "new"), bWithContent("c", "new")},
			wantIDs: []string{"a", "b", "c"},
			wantContent: map[string]string{
				"a": "old",
				"b": "new",
				"c": "new",
			},
		},
		{
			name:    "upsert nested block in place",
			blocks:  []Block{b("a", bWithContent("a1", "old")), bWithContent("b", "old")},
			pos:     PositionTail,
			add:     []Block{bWithContent("a1", "new")},
			wantIDs: []string{"a", "b"},
			wantContent: map[string]string{
				"a1": "new",
				"b":  "old",
			},
		},
		{
			name:    "duplicate insert same id twice",
			blocks:  []Block{bWithContent("a", "v1")},
			pos:     PositionTail,
			add:     []Block{bWithContent("a", "v2")},
			wantIDs: []string{"a"},
			wantContent: map[string]string{
				"a": "v2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := InsertBlocksAfter(tt.blocks, tt.pos, tt.add)
			th.AssertEqual(t, blockIDs(result), tt.wantIDs, "ids")

			for id, wantText := range tt.wantContent {
				block, found := FindBlock(result, id)
				th.AssertEqual(t, found, true, "block "+id+" found")
				gotText := ""
				if len(block.Content) > 0 {
					gotText = block.Content[0].Text
				}
				th.AssertEqual(t, gotText, wantText, "block "+id+" content")
			}
		})
	}
}

func TestDeleteBlocksByID(t *testing.T) {
	tests := []struct {
		name         string
		blocks       []Block
		del          []string
		want         []string
		wantChildren []string
	}{
		{"single", bs("a", "b", "c"), []string{"b"}, []string{"a", "c"}, nil},
		{"multiple", bs("a", "b", "c"), []string{"a", "c"}, []string{"b"}, nil},
		{"nested", []Block{b("a", b("a1"), b("a2")), b("b")}, []string{"a1"}, []string{"a", "b"}, []string{"a2"}},
		{"parent", []Block{b("a", b("a1")), b("b")}, []string{"a"}, []string{"b"}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DeleteBlocksByID(tt.blocks, tt.del)
			th.AssertEqual(t, blockIDs(result), tt.want, "ids")
			if tt.wantChildren != nil {
				th.AssertEqual(t, blockIDs(result[0].Children), tt.wantChildren, "children")
			}
		})
	}
}

func TestFindBlockDepth(t *testing.T) {
	th.RunMapTests(t, map[string]int{
		"a":   0,
		"a1":  1,
		"a1a": 2,
		"z":   -1,
	}, func(id string) int {
		blocks := []Block{b("a", b("a1", b("a1a"))), b("b")}
		return FindBlockDepth(blocks, id)
	})
}

func TestMoveBlocksAfter(t *testing.T) {
	result := MoveBlocksAfter(bs("a", "b", "c", "d"), []string{"b"}, "c")
	th.AssertEqual(t, blockIDs(result), []string{"a", "c", "b", "d"}, "ids")
}

func TestExtractBlocksByID(t *testing.T) {
	blocks := []Block{b("a", b("a1")), b("b"), b("c")}
	result := ExtractBlocksByID(blocks, []string{"b", "a1"})
	th.AssertEqual(t, len(result), 2, "count")
}

func TestReplaceBlocksByID(t *testing.T) {
	result := ReplaceBlocksByID(bs("a", "b", "c"), []string{"b"}, bs("x", "y"))
	th.AssertEqual(t, blockIDs(result), []string{"a", "x", "y", "c"}, "ids")
}

func TestFindBlock(t *testing.T) {
	blocks := []Block{b("a", b("a1", b("a1a"))), b("b")}
	th.RunMapTests(t, map[string]bool{
		"a": true, "a1": true, "a1a": true, "z": false,
	}, func(id string) bool {
		_, found := FindBlock(blocks, id)
		return found
	})
}

func TestUpdateBlocksProps(t *testing.T) {
	checked, unchecked := true, false

	tests := []struct {
		name   string
		blocks []Block
		ids    []string
		props  BlockProps
		check  func(t *testing.T, result []Block)
	}{
		{
			"toggle checked",
			[]Block{
				{ID: "a", Type: BlockTypeCheckList, Props: BlockProps{CheckListProps: CheckListProps{Checked: &checked}}},
				{ID: "b", Type: BlockTypeCheckList},
				{ID: "c", Type: BlockTypeParagraph},
			},
			[]string{"a", "b"},
			BlockProps{CheckListProps: CheckListProps{Checked: &unchecked}},
			func(t *testing.T, result []Block) {
				th.AssertEqual(t, *result[0].Props.Checked, false, "a.checked")
				th.AssertEqual(t, *result[1].Props.Checked, false, "b.checked")
				th.AssertEqual(t, result[2].Props.Checked, (*bool)(nil), "c.checked")
			},
		},
		{
			"nested heading level",
			[]Block{b("a", Block{ID: "a1", Type: BlockTypeHeading, Props: BlockProps{HeadingProps: HeadingProps{Level: 1}}}), b("b")},
			[]string{"a1"},
			BlockProps{HeadingProps: HeadingProps{Level: 2}},
			func(t *testing.T, result []Block) {
				th.AssertEqual(t, result[0].Children[0].Props.Level, 2, "level")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UpdateBlocksProps(tt.blocks, tt.ids, tt.props)
			tt.check(t, result)
		})
	}
}
