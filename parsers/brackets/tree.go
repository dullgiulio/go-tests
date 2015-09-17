package main

import "fmt"

type entry interface {
	String() string
	Children() []entry
}

func printEntries(e entry, tabs int) {
	for i := 0; i < tabs; i++ {
		fmt.Print("\t")
	}
	fmt.Printf("%s", e.String())
	children := e.Children()
	if children != nil {
		for i := range children {
			printEntries(children[i], tabs+1)
		}
	}
}

type text string

func (t *text) String() string {
	return string(*t)
}

func (t *text) Children() []entry {
	return nil
}

type list struct {
	entries []entry
}

func (l *list) newList() *list {
	return &list{entries: make([]entry, 0)}
}

func (l *list) add(e entry) {
	l.entries = append(l.entries, e)
}

func (l *list) String() string {
	return fmt.Sprintf("[list %d]", len(l.entries))
}

func (l *list) Children() []entry {
	return l.entries
}
