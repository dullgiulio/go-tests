package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

func readStruct(r *bufio.Reader) error {
	var numbuf bytes.Buffer
	var err error

	for {
		var b byte
		if b, err = r.ReadByte(); err != nil {
			break
		}
		if b == '\n' {
			break
		}
		if err = numbuf.WriteByte(b); err != nil {
			break
		}
	}

	if err != nil {
		return err
	}

	var nextlen int64
	num := numbuf.String()
	nextlen, err = strconv.ParseInt(num[1:], 10, 32)
	if err != nil || nextlen <= 0 {
		return err
	}
	jsonbuf := make([]byte, int(nextlen))
	if _, err := io.ReadFull(io.Reader(r), jsonbuf); err != nil {
		return err
	}

	var t Test
	if err := json.Unmarshal(jsonbuf, &t); err != nil {
		return err
	}
	fmt.Printf("%v\n", &t)

	// Consume newline
	_, err = r.ReadByte()
	return err
}

func readStructs(r io.Reader) error {
	bufr := bufio.NewReader(r)
	for {
		if err := readStruct(bufr); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	return nil
}
