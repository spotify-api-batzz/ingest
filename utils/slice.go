package utils

func ChunkSlice(toChunk []string, size int) [][]string {
	var divided [][]string

	for i := 0; i < len(toChunk); i += size {
		end := i + size

		if end > len(toChunk) {
			end = len(toChunk)
		}

		divided = append(divided, toChunk[i:end])
	}
	return divided
}
