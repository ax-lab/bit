package boot

func SliceSkipParted[S ~[]T, T any](ls S, skip func(v T) bool) (skipCount int) {
	sta, end := 0, len(ls)
	for sta < end {
		mid := sta + (end-sta)/2
		if skip(ls[mid]) {
			sta = mid + 1
		} else {
			end = mid
		}
	}
	return sta
}
