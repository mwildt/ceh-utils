package utils

func Map[T any, R any](collection []T, mapper func(T) R) (result []R) {
	for _, element := range collection {
		result = append(result, mapper(element))
	}
	return result
}
