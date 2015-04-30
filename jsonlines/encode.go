package main

import (
	"encoding/json"
	"fmt"
	"io"
)

func writeStruct(w io.Writer, v interface{}) error {
	if b, err := json.Marshal(v); err == nil {
		if _, err := fmt.Fprintf(w, "$%d\n", len(b)); err != nil {
			return err
		}
		if _, err := w.Write(b); err != nil {
			return err
		}
		if _, err := w.Write([]byte{'\n'}); err != nil {
			return err
		}
	}
	return nil
}
