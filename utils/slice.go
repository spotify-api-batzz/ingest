package utils

func ChunkSlice[T any](toChunk []T, size int) [][]T {
	var divided [][]T

	for i := 0; i < len(toChunk); i += size {
		end := i + size

		if end > len(toChunk) {
			end = len(toChunk)
		}

		divided = append(divided, toChunk[i:end])
	}
	return divided
}

func RemoveExcludedFromSlice(slice []string, exclude []string) []string {
	excludedSlice := []string{}
	for _, item := range slice {
		valid := true
		for _, excludedItem := range exclude {
			if excludedItem == item {
				valid = false
				break
			}
		}
		if valid {
			excludedSlice = append(excludedSlice, item)
		}
	}
	return excludedSlice
}

func KeepIncludedInSlice(slice []string, include []string) []string {
	includedSlice := []string{}
	for _, item := range slice {
		valid := false
		for _, includedItem := range include {
			if includedItem == item {
				valid = true
				break
			}
		}
		if valid {
			includedSlice = append(includedSlice, item)
		}
	}
	return includedSlice
}
