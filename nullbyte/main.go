package main

import (
    "bytes"
    "bufio"
    "fmt"
    "log"
    "os"
)

type EOL byte

// dropCR drops a terminal \r from the data.
func (e EOL) dropCR(data []byte) []byte {
    if byte(e) == '\n' && len(data) > 0 && data[len(data)-1] == '\r' {
        return data[0 : len(data)-1]
    }
    return data
}

func (e EOL) ScanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
    if atEOF && len(data) == 0 {
        return 0, nil, nil
    }
    if i := bytes.IndexByte(data, byte(e)); i >= 0 {
        // We have a full eol-terminated line.
        return i + 1, e.dropCR(data[0:i]), nil
    }
    // If we're at EOF, we have a final, non-terminated line. Return it.
    if atEOF {
        return len(data), e.dropCR(data), nil
    }
    // Request more data.
    return 0, nil, nil
}

func main() {
    eol := EOL('\000')
    sc := bufio.NewScanner(os.Stdin)
    sc.Split(eol.ScanLines)
    for sc.Scan() {
        fmt.Printf("%s\n", sc.Text())
    }
    if err := sc.Err(); err != nil {
        log.Fatal(err)
    }
}
