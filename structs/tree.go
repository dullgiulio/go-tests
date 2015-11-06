package main

import (
	"fmt"
	"math/rand"
	"time"
)

type leaf struct {
	text     string
	cnt      int64
	children []*leaf
}

func newLeaf(s string) *leaf {
	return &leaf{
		text:     s,
		children: make([]*leaf, 0),
	}
}

func (l *leaf) String() string {
	return l.text
}

func (l *leaf) add(c *leaf) *leaf {
	l.children = append(l.children, c)
	return c
}

func (l *leaf) find(s string) int {
	for i := range l.children {
		if l.children[i].text == s {
			return i
		}
	}
	return -1
}

func (l *leaf) increment() {
	l.cnt++
}

/*

// Implement sortable for children
func (l *list) Len() int {
	return len(l.children)
}

func (l *list) Swap(i, j int) {
	l.children[i], l.children[j] = l.children[j], l.children[i]
}

func (l *list) Less(i, j int) bool {
	return l.children[i].cnt < l.children[j].cnt
}

// TODO: make concurrent, non-recursive.
func sort(lf *leaf) {
	if len(lf.children) == 0 {
		return
	}
	sort.Sort(lf)
	for i := range lf.children {
		sort(lf.children[i])
	}
}

*/

func insert(lf *leaf, in <-chan string) {
	for s := range in {
		if pos := lf.find(s); pos != -1 {
			lf = lf.children[pos]
			lf.increment()
			continue
		}
		lf = lf.add(newLeaf(s))
	}
}

type walker struct {
	root  *leaf
	rnd   *rand.Rand
	depth int
}

func newWalker(root *leaf) *walker {
	return &walker{
		root:  root,
		rnd:   rand.New(rand.NewSource(time.Now().UnixNano())),
		depth: depth(root),
	}
}

// TODO: make concurrent, non-recursive.
func depth(lf *leaf) int {
	if len(lf.children) == 0 {
		return 1
	}
	max := 0
	for i := range lf.children {
		d := depth(lf.children[i])
		if d > max {
			max = d
		}
	}
	return max + 1
}

// TODO: this works with numbers that might easily overflow.
func (w *walker) weightedRand(lf *leaf) *leaf {
	// Sum all the counts.
	var total int64
	for i := range lf.children {
		total += lf.children[i].cnt + 1
	}
	// Get a number n, 0 <= n < total
	n := w.rnd.Int63n(total)
	// Find which leaf i this number corresponds to in the sum.
	total = 0
	for i := range lf.children {
		ntotal := total + lf.children[i].cnt + 1
		if total <= n && n < ntotal {
			return lf.children[i]
		}
		total = ntotal
	}
	panic("not reachable")
}

func (w *walker) randWalk(depth int) {
	if depth <= 0 {
		depth = w.depth
	}
	if len(w.root.children) == 0 {
		return
	}
	lf := w.weightedRand(w.root)
	for i := 0; i < depth; i++ {
		fmt.Printf("%s ", lf) // TODO: Emit node
		if len(lf.children) == 0 {
			return
		}
		lf = w.weightedRand(lf)
	}
	fmt.Printf("\n")
}
