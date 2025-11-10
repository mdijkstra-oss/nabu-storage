package fileview

// ToSummary strips chunks from File for context-efficient responses
func ToSummary(f File) FileSummary {
	return FileSummary{BaseFile: f.BaseFile}
}
