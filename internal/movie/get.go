package movie

func GetValue[T any](v Value[T]) T {
	return v.Value
}

func GetValueList[T any](v []Value[T]) []T {
	return convertValueListToArray(v)
}

func ConvertArrayToValueList[T any](items []T, entryType string) []Value[T] {
	out := make([]Value[T], len(items))
	for i, v := range items {
		out[i] = Value[T]{Value: v, EntryType: entryType}
	}
	return out
}

func convertValueListToArray[T any](items []Value[T]) []T {
	var values []T
	for _, v := range items {
		values = append(values, v.Value)
	}
	return values
}
