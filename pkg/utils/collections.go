package utils

func Map[T any, R any](collection []T, mapper func(T) R) (result []R) {
	for _, element := range collection {
		result = append(result, mapper(element))
	}
	return result
}

func FilterValues[K comparable, T any](collection map[K]T, predicate Predicate[T]) (result []T) {
	result = make([]T, 0)
	for _, element := range collection {
		if predicate(element) {
			result = append(result, element)
		}
	}
	return result
}

func Filter[T any](collection []T, predicate Predicate[T]) (result []T) {
	result = make([]T, 0)
	for _, element := range collection {
		if predicate(element) {
			result = append(result, element)
		}
	}
	return result
}

func AnyMatch[T any](collection []T, predicate Predicate[T]) bool {
	for _, element := range collection {
		if predicate(element) {
			return true
		}
	}
	return false
}

type Predicate[T any] func(q T) bool

func True[T any]() Predicate[T] {
	return func(t T) bool {
		return true
	}
}
