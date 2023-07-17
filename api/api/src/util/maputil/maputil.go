package maputil

// Array converts the given map into an array with m's values.
func Array[K comparable, V any](m map[K]V) []V {
	arr := make([]V, 0, len(m))
	for _, v := range m {
		arr = append(arr, v)
	}
	return arr
}
