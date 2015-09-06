package append

import "sort"

func MergeA(a, b []int) []int {
	c := make([]int, len(a))
	copy(c, a)
	for _, bv := range b {
		found := false
		for _, cv := range c {
			if bv == cv {
				found = true
				break
			}
		}
		if !found {
			c = append(c, bv)
		}
	}
	return c
}

func MergeC(a, b []int) []int {
	c := make([]int, len(a))
	copy(c, a)
	for _, bv := range b {
		i := sort.Search(len(c), func(i int) bool { return c[i] >= bv })
		if i >= len(c) || c[i] != bv {
			c = append(c, bv)
		}
	}
	sort.Ints(c)
	return c
}

func MergeB(a, b []int) []int {
	// Create a map that holds the values from each slice
	unique := make(map[int]struct{}) // zero byte payload

	for _, v := range a {
		unique[v] = struct{}{} // zero byte payload
	}
	for _, v := range b {
		unique[v] = struct{}{} // zero byte payload
	}

	final := make([]int, len(unique)) // allocate to right size
	i := 0
	for k := range unique {
		final[i] = k
		i++ // index not a feature of map ranges
	}
	// sort.Ints(final) // non-decreasing order
	return final
}
