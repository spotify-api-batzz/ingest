package main

import "time"

func chunkSlice(toChunk []string, size int) [][]string {
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

func genNiceTime() string {
	timeFormat := "Mon 2 Jan 2006 15-04-05"
	time := time.Now()
	return time.Format(timeFormat)
}
