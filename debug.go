package sima

import (
	"log"
)

type debugT bool

var debug = debugT(true) // XXX: Remember to set to false

func (d debugT) Printf(format string, args ...interface{}) {
    if d {
        log.Printf(format, args...)
    }
}
