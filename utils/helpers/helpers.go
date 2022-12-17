package helpers

// SliceIncludes is used to check if T is in a slice.
func SliceIncludes[T comparable](slice []T, t T) bool {
	for _, v := range slice {
		if v == t {
			return true
		}
	}
	return false
}

// MapValues is used to get the values of a map.
func MapValues[K comparable, V any](m map[K]V) []V {
	values := make([]V, len(m))
	i := 0
	for _, v := range m {
		values[i] = v
		i++
	}
	return values
}
