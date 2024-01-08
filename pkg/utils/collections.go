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

func Count[T any](collection []T, predicate Predicate[T]) int {
	return len(Filter(collection, predicate))
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

func Contains[T comparable](collection []T, element T) bool {
	for _, item := range collection {
		if item == element {
			return true
		}
	}
	return false
}

func ContainsAll[T comparable](container []T, elements []T) bool {
	if len(container) < len(elements) {
		return false
	}
	for _, element := range elements {
		if !Contains(container, element) {
			return false
		}
	}
	return true
}

func MutualContainment[T comparable](a []T, b []T) bool {
	return ContainsAll(a, b) && ContainsAll(b, a)
}
