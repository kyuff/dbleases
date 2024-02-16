package split

func By[T any, S ~[]T, K comparable](items S, fn func(value T) K) map[K]S {
	var splits = make(map[K]S)

	for _, item := range items {
		key := fn(item)
		splits[key] = append(splits[key], item)
	}

	return splits
}
