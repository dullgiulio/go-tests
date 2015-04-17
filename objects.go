package sima

import (
	"strings"
)

type Objects struct {
	Names []string
}

func NewObjects() *Objects {
	return &Objects{Names: make([]string, 0)}
}

func (o *Objects) Add(name string) {
	o.Names = append(o.Names, name)
}

func (o *Objects) String() string {
	return strings.Join(o.Names, ", ")
}
