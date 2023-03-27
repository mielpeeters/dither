package needle

import (
	"math"
)

// ChunkSlice splits the given slice in n chunks.
// The last chunk might be smaller if chunkSize is not a factor of the original slice's length
func ChunkSlice[T any](slice []T, n int) [][]T {
	var chunks [][]T
	var size int
	var checkSize bool = true

	for i := 0; i < len(slice); i += size {
		if checkSize {
			if (len(slice)-i)%n == 0 {
				checkSize = false
			}
			size = int(math.Ceil(float64(len(slice)-i) / float64(n)))
			n--
		}

		end := i + size

		// necessary check to avoid slicing beyond
		// slice capacity
		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

// ChunkMap splits the given map in n chunks.
// The last chunk might be smaller if chunkSize is not a factor of the original slice's length
func ChunkMap[T comparable, U any](slice map[T]U, n int) []map[T]U {
	var chunk map[T]U
	var chunks []map[T]U
	var keys []T
	chunks = make([]map[T]U, 0)

	for i := range slice {
		keys = append(keys, i)
	}

	chunkSize := int(math.Ceil(float64(len(slice)) / float64(n)))

	for i := 0; i < len(keys); i += chunkSize {
		end := i + chunkSize
		chunk = make(map[T]U)

		// necessary check to avoid slicing beyond
		// keys capacity
		if end > len(keys) {
			end = len(keys)
		}

		for j := i; j < end; j++ {
			chunk[keys[j]] = slice[keys[j]]
		}

		chunks = append(chunks, chunk)
	}

	return chunks
}
