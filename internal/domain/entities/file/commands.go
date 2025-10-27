package file

const EntityName = "File"

const (
	Create      = "Create" // point to file prob in future, not whole contents
	CodeFile    = "CodeFile"
	ClearCoding = "ClearCoding"
	MergeCodes  = "MergeCodes"
)

type CreateFilePayload = CreateFileData
type CodeFilePayload = CodeFileData
type MergeCodesPayload = MergeCodesData
