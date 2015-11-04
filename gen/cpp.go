// +build ignore
//!go:generate sh -c "sed 's,^//\\#,\\#,g' main.go | cpp -E -P -o ../../main1.go && go fmt ../../main.go"
package main

import (
    "fmt"
)

//#define errc()  if err != nil {return err}

func test() (err error) {
    _, err := fmt.Println("1213")
    errc()
}

func main() {
    test()
}

