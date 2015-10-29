package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	flag.Parse()
	for _, arg := range flag.Args() {
		fi, err := os.Stat(arg)
		if err != nil {
			log.Fatal(err)
			continue
		}
		fmt.Printf("%s\tstat\t%v\n", arg, fi)
		fi, err = os.Lstat(arg)
		if err != nil {
			log.Print(err)
			continue
		}
		fmt.Printf("%s\tlstat\t%v\n", arg, fi)
	}
}
