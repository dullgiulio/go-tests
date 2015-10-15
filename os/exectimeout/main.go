package main

import (
    "bytes"
    "errors"
    "fmt"
    "os/exec"
    "time"
)

// GetCommandOutput returns stdout of command as a string, error and stderr
// if error we return like this: nil, error and concatenated stdout and stderr ( some commands show the errors on stdout )
func GetCommandOutput(command string, timeout time.Duration, arg ...string) ([]byte, error, []byte) {
    var err error
    var out, errOut bytes.Buffer
    var c = make(chan []byte)

    cmd := exec.Command(command, arg...)
    cmd.Stdout = &out
    cmd.Stderr = &errOut
    if err = cmd.Start(); err != nil {
        message := append(out.Bytes(), errOut.Bytes()...)
        return nil, err, message
    }
    go func() {
        err = cmd.Wait()
        c <- out.Bytes()
    }()
    time.AfterFunc(timeout, func() {
        cmd.Process.Kill()
        if err == nil {
            err = errors.New("Maxruntime exceeded executing command")
        } else {
            err = fmt.Errorf("Maxruntime exceeded executing command, kill process error: %s", err)
        }
    })
    data := <-c
    if err != nil {
        message := append(out.Bytes(), errOut.Bytes()...)
        return nil, err, message
    }
    return data, nil, nil
}

func main() {
    data, _, _ := GetCommandOutput("./sleepecho.sh", 1 * time.Second, "")
    fmt.Printf("Data: %s\n", data)
}
