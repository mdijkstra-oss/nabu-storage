package file

const EntityName = "File"

const (
	CreateFile  = "CreateFile"
	CodeFile    = "CodeFile"
	ClearCoding = "ClearCoding" // Remove all coding from given file
	MergeCodes  = "MergeCodes"  // Merge code S & T into set T
)

type CreateFilePayload = CreateFileData

type CodeFilePayload = CodeFileData
type MergeCodesPayload = MergeCodesData
