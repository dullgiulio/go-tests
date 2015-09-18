package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
)

func toInt64Slice(buf []byte) ([]int64, error) {
	if len(buf)%8 != 0 {
		return nil, errors.New("trailing bytes")
	}
	vals := make([]int64, len(buf)/8)
	for i := 0; i < len(vals); i++ {
		val := binary.LittleEndian.Uint64(buf[i*8:])
		vals[i] = int64(val)
	}
	return vals, nil
}

func toByteSlice(vals []int64) []byte {
	buf := make([]byte, len(vals)*8)
	for i, v := range vals {
		binary.LittleEndian.PutUint64(buf[i*8:], uint64(v))
	}
	return buf
}

func main() {
	var err error
	vals := []int64{100, -101, 102, -103}
	fmt.Printf("%v\n", vals)
	data := toByteSlice(vals)
	vals, err = toInt64Slice(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v\n", vals)
}
