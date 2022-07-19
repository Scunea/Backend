package utils

func ExistsFunc[E any](s []E, f func(E) bool) bool {
	for _, v := range s {
		if f(v) {
			return true
		}
	}
	return false
}

func IndexFunc[E any](s []E, f func(E) bool) int {
	for i, v := range s {
		if f(v) {
			return i
		}
	}
	return -1
}

func FilterFunc[E any](s []E, f func(E) bool) []E {
	var filtered []E
	for _, v := range s {
		if f(v) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func Remove[E any](s []E, i int) []E {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
