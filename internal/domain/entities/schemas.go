package entities

import (
	"bytes"
	_ "embed"
	"encoding/json"
)

//go:embed project/schema.json
var projectSchemaRaw []byte

//go:embed file/schema.json
var FileSchema []byte

//go:embed code/schema.json
var CodeSchema []byte

var ProjectSchema []byte

func init() {
	ProjectSchema = expandProjectSchema(projectSchemaRaw, CodeSchema, FileSchema)
}

func expandProjectSchema(project, code, file []byte) []byte {
	mustUnmarshal := func(data []byte, v any) {
		if err := json.Unmarshal(data, v); err != nil {
			panic("invalid embedded schema: " + err.Error())
		}
	}

	var schema map[string]any
	mustUnmarshal(project, &schema)

	var codeSchema map[string]any
	mustUnmarshal(code, &codeSchema)
	delete(codeSchema, "$schema")

	var fileSchema map[string]any
	mustUnmarshal(file, &fileSchema)
	delete(fileSchema, "$schema")

	props := schema["properties"].(map[string]any)
	codes := props["codes"].(map[string]any)
	codes["items"] = codeSchema

	files := props["files"].(map[string]any)
	files["items"] = fileSchema

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(schema); err != nil {
		panic("failed to encode expanded schema: " + err.Error())
	}
	return buf.Bytes()
}
