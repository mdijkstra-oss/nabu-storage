Here's a function example:

```go
func ProcessLargeDataset(data []string) error {
    for i, item := range data {
        if err := validateItem(item); err != nil {
            return fmt.Errorf("validation failed at index %d: %w", i, err)
        }
        processedItem := transformItem(item)
        if err := saveToDatabase(processedItem); err != nil {
            return fmt.Errorf("database save failed: %w", err)
        }
    }
    return nil
}
```

This code block exceeds 100 bytes and must stay intact.
