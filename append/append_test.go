package append

import (
	"sort"
	"testing"
)

var mergeTests = []struct {
	a, b, m []int
}{
	{[]int{1, 2, 3}, []int{}, []int{1, 2, 3}},
	{[]int{1, 2, 3}, []int{1}, []int{1, 2, 3}},
	{[]int{1, 2, 3}, []int{1, 2}, []int{1, 2, 3}},
	{[]int{1, 2, 3}, []int{1, 2, 3}, []int{1, 2, 3}},
	{[]int{1, 2, 3}, []int{1, 2, 3, 4}, []int{1, 2, 3, 4}},

	{[]int{1, 2, 3}, []int{1, 2, 3, 4}, []int{1, 2, 3, 4}},
	{[]int{1, 2, 3}, []int{1, 2, 4, 3}, []int{1, 2, 3, 4}},
	{[]int{1, 2, 3}, []int{1, 4, 2, 3}, []int{1, 2, 3, 4}},
	{[]int{1, 2, 3}, []int{4, 1, 2, 3}, []int{1, 2, 3, 4}},
}

func TestMergeA(t *testing.T) {
	for i, a := range mergeTests {
		m := MergeA(a.a, a.b)
		same := len(m) == len(a.m)
		for i := 0; same && i < len(m); i++ {
			same = m[i] == a.m[i]
		}
		if !same {
			t.Errorf("#%d, MergeA(%v, %v) value is %v; want %v", i, a.a, a.b, m, a.m)
		}
	}
}

func TestMergeB(t *testing.T) {
	for i, a := range mergeTests {
		m := MergeB(a.a, a.b)
		sort.Ints(m)
		same := len(m) == len(a.m)
		for i := 0; same && i < len(m); i++ {
			same = m[i] == a.m[i]
		}
		if !same {
			t.Errorf("#%d, MergeB(%v, %v) value is %v; want %v", i, a.a, a.b, m, a.m)
		}
	}
}

func TestMergeC(t *testing.T) {
	for i, a := range mergeTests {
		m := MergeC(a.a, a.b)
		same := len(m) == len(a.m)
		for i := 0; same && i < len(m); i++ {
			same = m[i] == a.m[i]
		}
		if !same {
			t.Errorf("#%d, MergeC(%v, %v) value is %v; want %v", i, a.a, a.b, m, a.m)
		}
	}
}

//
// BENCHMARKS
//

func benchmarkAppendA(b *testing.B, n int) {
	for i := 0; i < b.N; i++ {
		// b.StopTimer()
		a := make([]int, 0)
		// b.StartTimer()
		for j := 0; j < n; j++ {
			a = append(a, 1)
		}
	}
}

func BenchmarkAppendA10(b *testing.B)       { benchmarkAppendA(b, 10) }
func BenchmarkAppendA100(b *testing.B)      { benchmarkAppendA(b, 100) }
func BenchmarkAppendA1000(b *testing.B)     { benchmarkAppendA(b, 1000) }
func BenchmarkAppendA10000(b *testing.B)    { benchmarkAppendA(b, 10000) }
func BenchmarkAppendA100000(b *testing.B)   { benchmarkAppendA(b, 100000) }
func BenchmarkAppendA1000000(b *testing.B)  { benchmarkAppendA(b, 1000000) }
func BenchmarkAppendA10000000(b *testing.B) { benchmarkAppendA(b, 10000000) }

func benchmarkAppendB(b *testing.B, n int) {
	for i := 0; i < b.N; i++ {
		// b.StopTimer()
		a := make([]int, 0, n)
		// b.StartTimer()
		for j := 0; j < n; j++ {
			a = append(a, 1)
		}
	}
}

func BenchmarkAppendB10(b *testing.B)       { benchmarkAppendB(b, 10) }
func BenchmarkAppendB100(b *testing.B)      { benchmarkAppendB(b, 100) }
func BenchmarkAppendB1000(b *testing.B)     { benchmarkAppendB(b, 1000) }
func BenchmarkAppendB10000(b *testing.B)    { benchmarkAppendB(b, 10000) }
func BenchmarkAppendB100000(b *testing.B)   { benchmarkAppendB(b, 100000) }
func BenchmarkAppendB1000000(b *testing.B)  { benchmarkAppendB(b, 1000000) }
func BenchmarkAppendB10000000(b *testing.B) { benchmarkAppendB(b, 10000000) }

func benchmarkAppendC(b *testing.B, n int) {
	for i := 0; i < b.N; i++ {
		// b.StopTimer()
		a := make([]int, n)
		// b.StartTimer()
		for j := 0; j < n; j++ {
			a[j] = 1
		}
	}
}

func BenchmarkAppendC10(b *testing.B)       { benchmarkAppendC(b, 10) }
func BenchmarkAppendC100(b *testing.B)      { benchmarkAppendC(b, 100) }
func BenchmarkAppendC1000(b *testing.B)     { benchmarkAppendC(b, 1000) }
func BenchmarkAppendC10000(b *testing.B)    { benchmarkAppendC(b, 10000) }
func BenchmarkAppendC100000(b *testing.B)   { benchmarkAppendC(b, 100000) }
func BenchmarkAppendC1000000(b *testing.B)  { benchmarkAppendC(b, 1000000) }
func BenchmarkAppendC10000000(b *testing.B) { benchmarkAppendC(b, 10000000) }

const LIMIT = 100000

var iota [LIMIT + 1]int

func init() {
	for i := range iota {
		iota[i] = i
	}
}

func benchmarkMergeA(b *testing.B, n int) {
	for i := 0; i < b.N; i++ {
		_ = MergeA(iota[0:n], iota[1:n+1])
	}
}

func BenchmarkMergeA10(b *testing.B)    { benchmarkMergeA(b, 10) }
func BenchmarkMergeA100(b *testing.B)   { benchmarkMergeA(b, 100) }
func BenchmarkMergeA200(b *testing.B)   { benchmarkMergeA(b, 200) }
func BenchmarkMergeA300(b *testing.B)   { benchmarkMergeA(b, 300) }
func BenchmarkMergeA400(b *testing.B)   { benchmarkMergeA(b, 400) }
func BenchmarkMergeA500(b *testing.B)   { benchmarkMergeA(b, 500) }
func BenchmarkMergeA1000(b *testing.B)  { benchmarkMergeA(b, 1000) }
func BenchmarkMergeA10000(b *testing.B) { benchmarkMergeA(b, 10000) }

func benchmarkMergeB(b *testing.B, n int) {
	for i := 0; i < b.N; i++ {
		_ = MergeB(iota[0:n], iota[1:n+1])
	}
}

func BenchmarkMergeB10(b *testing.B)    { benchmarkMergeB(b, 10) }
func BenchmarkMergeB100(b *testing.B)   { benchmarkMergeB(b, 100) }
func BenchmarkMergeB200(b *testing.B)   { benchmarkMergeB(b, 200) }
func BenchmarkMergeB300(b *testing.B)   { benchmarkMergeB(b, 300) }
func BenchmarkMergeB400(b *testing.B)   { benchmarkMergeB(b, 400) }
func BenchmarkMergeB500(b *testing.B)   { benchmarkMergeB(b, 500) }
func BenchmarkMergeB1000(b *testing.B)  { benchmarkMergeB(b, 1000) }
func BenchmarkMergeB10000(b *testing.B) { benchmarkMergeB(b, 10000) }

func benchmarkMergeC(b *testing.B, n int) {
	for i := 0; i < b.N; i++ {
		_ = MergeC(iota[0:n], iota[1:n+1])
	}
}

func BenchmarkMergeC10(b *testing.B)    { benchmarkMergeC(b, 10) }
func BenchmarkMergeC100(b *testing.B)   { benchmarkMergeC(b, 100) }
func BenchmarkMergeC200(b *testing.B)   { benchmarkMergeC(b, 200) }
func BenchmarkMergeC300(b *testing.B)   { benchmarkMergeC(b, 300) }
func BenchmarkMergeC400(b *testing.B)   { benchmarkMergeC(b, 400) }
func BenchmarkMergeC500(b *testing.B)   { benchmarkMergeC(b, 500) }
func BenchmarkMergeC1000(b *testing.B)  { benchmarkMergeC(b, 1000) }
func BenchmarkMergeC10000(b *testing.B) { benchmarkMergeC(b, 10000) }
