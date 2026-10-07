// Package utils holds small, dependency-free helpers shared across the project.
package utils

// ChunkSlice splits items into consecutive batches of at most size elements.
// A size <= 0 returns a single batch containing all items. The returned
// sub-slices share backing storage with items (no copy).
func ChunkSlice[T any](items []T, size int) [][]T {
	if size <= 0 {
		return [][]T{items}
	}
	batches := make([][]T, 0, (len(items)+size-1)/size)
	for start := 0; start < len(items); start += size {
		end := start + size
		if end > len(items) {
			end = len(items)
		}
		batches = append(batches, items[start:end])
	}
	return batches
}
