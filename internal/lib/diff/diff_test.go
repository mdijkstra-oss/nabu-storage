package diff

import (
	"testing"

	th "hermes-relay/internal/lib/test-helpers"
)

func TestApply(t *testing.T) {
	tests := []struct {
		Name     string
		Content  string
		Patch    string
		Expected Result
	}{
		{
			Name:    "creates new file from Add File",
			Content: "",
			Patch: `*** Add File: test.md
hello
world`,
			Expected: Ok("hello\nworld"),
		},
		{
			Name:    "updates file with single hunk",
			Content: "hello\nworld",
			Patch: `*** Update File: test.md
@@
-hello
+goodbye
world`,
			Expected: Ok("goodbye\nworld"),
		},
		{
			Name:    "applies multiple hunks",
			Content: "aaa\nbbb\nccc",
			Patch: `*** Update File: test.md
@@
-aaa
+AAA
bbb
@@
bbb
-ccc
+CCC`,
			Expected: Ok("AAA\nbbb\nCCC"),
		},
		{
			Name:    "preserves context lines",
			Content: "line1\nline2\nline3",
			Patch: `*** Update File: test.md
@@
line1
-line2
+replaced
line3`,
			Expected: Ok("line1\nreplaced\nline3"),
		},
		{
			Name:    "fails when patch context not found",
			Content: "hello",
			Patch: `*** Update File: test.md
@@
-nonexistent
+replacement`,
			Expected: Fail(`patch context not found: "nonexistent"`),
		},
		{
			Name:    "handles function rename example from spec",
			Content: "def fib(n):\n    if n <= 1:\n        return n\n    return fib(n-1) + fib(n-2)",
			Patch: `@@
-def fib(n):
+def fibonacci(n):
    if n <= 1:
        return n
-    return fib(n-1) + fib(n-2)
+    return fibonacci(n-1) + fibonacci(n-2)`,
			Expected: Ok("def fibonacci(n):\n    if n <= 1:\n        return n\n    return fibonacci(n-1) + fibonacci(n-2)"),
		},
		{
			Name:    "appends to content when old text is empty",
			Content: "existing",
			Patch: `*** Add File: test.md
appended`,
			Expected: Ok("existing\nappended"),
		},
		{
			Name:    "implicit hunk start with + lines",
			Content: "",
			Patch: `+# Hello
+World`,
			Expected: Ok("# Hello\nWorld"),
		},
		{
			Name:    "implicit hunk start with - and + lines",
			Content: "old content",
			Patch: `-old content
+new content`,
			Expected: Ok("new content"),
		},
		{
			Name:    "append with @@ and + lines only",
			Content: "# Title",
			Patch: `@@
+
+New paragraph here.`,
			Expected: Ok("# Title\n\nNew paragraph here."),
		},
		{
			Name:    "append to empty file with @@ and + lines",
			Content: "",
			Patch: `@@
+# Title`,
			Expected: Ok("# Title"),
		},
		{
			Name:    "append multiple sections incrementally",
			Content: "# Title",
			Patch: `@@
+
+Section one content.`,
			Expected: Ok("# Title\n\nSection one content."),
		},
		{
			Name:    "real scenario: append with empty + line",
			Content: "# Codebook",
			Patch: `@@
+
+This is a *sample* qualitative codebook for analyzing texts.`,
			Expected: Ok("# Codebook\n\nThis is a *sample* qualitative codebook for analyzing texts."),
		},
		{
			Name:    "append json block without anchor",
			Content: "# Doc\n\nIntro text.",
			Patch: `@@
+
+` + "```json-callout" + `
+{"id": "test", "type": "codebook"}
+` + "```",
			Expected: Ok("# Doc\n\nIntro text.\n\n```json-callout\n{\"id\": \"test\", \"type\": \"codebook\"}\n```"),
		},
		{
			Name:    "create file with @@ and + lines",
			Content: "",
			Patch: `@@
+# Coffee Bean Research Codebook`,
			Expected: Ok("# Coffee Bean Research Codebook"),
		},
		{
			Name:    "malformed: LLM prefixes @@ with +",
			Content: "",
			Patch: `+@@
++# Coffee Bean Research Codebook`,
			Expected: Ok("# Coffee Bean Research Codebook"),
		},
		{
			Name:    "forgiving: LLM adds leading space to context lines",
			Content: "# Coffee Bean Research Codebook",
			Patch: `@@
 # Coffee Bean Research Codebook
+
+This is a paragraph.`,
			Expected: Ok("# Coffee Bean Research Codebook\n\nThis is a paragraph."),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := Apply(tt.Content, tt.Patch)
			th.AssertEqual(t, result, tt.Expected, "result")
		})
	}
}
