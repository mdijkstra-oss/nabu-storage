package entities

import (
	_ "embed"
)

//go:embed project/schema.json
var ProjectSchema []byte

//go:embed file/schema.json
var FileSchema []byte

//go:embed code/schema.json
var CodeSchema []byte
