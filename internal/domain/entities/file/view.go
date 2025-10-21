package file

type File struct {
	ID      string `json:"id" validate:"required"`
	Content string `json:"content"`
	Attributes
}

type Attributes struct {
	Codes map[string][]string `json:"codes"`
}

// Claude: Used in /file/{id}/chunk/{chunkIndex} query
type FileChunk struct {
	ID             string `json:"id" validate:"required"`
	Chunk          string `json:"chunk"`
	ChunkIndex     int    `json:"chunk_index" validate:"required"`
	NextChunkIndex int    `json:"next_chunk_index"`
}
