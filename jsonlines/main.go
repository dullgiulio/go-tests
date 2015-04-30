package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
)

type Test struct {
	Names  []string
	Values []string
}

var _letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-")

func randstr(n int) string {
	b := make([]rune, n)
	l := len(_letters)

	for i := range b {
		b[i] = _letters[rand.Intn(l)]
	}

	return string(b)
}

func makeTest() *Test {
	t := new(Test)
	t.Names = make([]string, 0)
	for i := 0; i < rand.Intn(10); i++ {
		t.Names = append(t.Names, randstr(rand.Intn(10)))
	}
	t.Values = make([]string, 0)
	for i := 0; i < rand.Intn(10); i++ {
		t.Values = append(t.Values, randstr(rand.Intn(10)))
	}
	return t
}

func main() {
	file := "test.json"
	output, err := os.Create(file)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < rand.Intn(10); i++ {
		t := makeTest()
		fmt.Printf("%v\n", t)
		if err := writeStruct(output, t); err != nil {
			log.Fatal(err)
		}
	}
	output.Close()

	input, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	err = readStructs(input)
	if err != nil {
		log.Fatal(err)
	}

	os.Remove(file)
}
