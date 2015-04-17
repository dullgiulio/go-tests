package sima

import (
	"strings"
)

type Methods struct {
    Names []string
}

func NewMethods() *Methods {
    return &Methods{ Names: make([]string, 0) }
}

func (m *Methods) Add(name string) {
    m.Names = append(m.Names, name)
}

func (m *Methods) String() string {
	return strings.Join(m.Names, ", ")
}
