package main

import (
    "os"
    "fmt"
    "log"
)

func main() {
    fi, err := os.Stdin.Stat()
    if err != nil {
        log.Fatal(err)
    }

    if fi.Mode() & os.ModeNamedPipe == os.ModeNamedPipe {
        fmt.Printf("Input is a pipe")
    }
}
