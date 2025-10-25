# DRY

JFC Junior, listen up. Code you generate has duplication issues. We're not doing that again.

No temporary files. No code duplication. No copy-paste.
If I see the same logic written twice, you're rewriting it.
DRY means Don't Repeat Yourself. Not "Don't Repeat Yourself Except In Tests." Not "Don't Repeat Yourself Unless It's Easier."
Every time you copy-paste, an engineer dies inside. Don't make me read duplicated test code.
Write it once. Write it right. Use it everywhere.

# Tests
- WRITE DRY TESTS - NO REPETITION
- Use table-driven tests where possible

Sample:  
```go
tests := []struct {
    name     string
    query    string
    expected int  // number of matches
}{
    {"exact match", "some text", 1},
    {"extra spaces", "the  dog", 1},
    {"with punctuation", "Hello world", 1},
    // etc
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // ONE test implementation, runs for ALL cases
    })
}
```
Write test helper functions! 

