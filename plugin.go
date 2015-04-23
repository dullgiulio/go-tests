package sima

import (
	"bufio"
	"io"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	"strings"
	"time"
)

type pluginStatus int

const (
	pluginStatusNone pluginStatus = iota
	pluginStatusRunning
)

type plugin struct {
	exe    string
	proto  string
	addr   string
	objs   []string
	killCh <-chan time.Time
	status pluginStatus
	header *headerReader
}

func NewPlugin(exe string) *plugin {
	return &plugin{
		exe:    exe,
		header: newHeaderReader(),
		objs:   make([]string, 0),
		proto:  "unix",
		addr:   "test",
	}
}

func (p *plugin) String() string {
	return p.exe + " " + strings.Join(p.objs, ", ")
}

func (p *plugin) start() error {
	cmd := exec.Command(p.exe, "-sima:proto=unix", "-sima:addr=test")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	ready := make(chan struct{})

	go func(r io.Reader) {
		scanner := bufio.NewScanner(r)

		for scanner.Scan() {
			log.Print("subprocess: ", scanner.Text())

			if ready != nil {
				ready <- struct{}{}
				close(ready)
				ready = nil
			}
		}
	}(stdout)

	cmd.Start()
	p.killCh = nil

	go p.wait(cmd)

	<-ready

	return nil
}

func (p *plugin) stop() error {
	respCh := make(chan resp)
	p.call("SimaRpc.Exit", 1, SimaNil, respCh)

	p.killCh = time.After(1 * time.Second)

	resp := <-respCh
	close(respCh)

	return resp.err
}

func (p *plugin) call(name string, args interface{}, response interface{}, respCh chan<- resp) {
	// TODO: Find which plugin contains the object, start it if necessary.
	// The client object should come from the plugin
	client, err := rpc.DialHTTP(p.proto, p.addr)
	if err != nil {
		log.Print("dial failed: ", err)
		respCh <- resp{err: err}
		return
	}

	if p.proto == "unix" {
		os.Remove(p.addr)
	}

	go func() {
		var t string
		err = client.Call(name, args, &t)
		log.Print("response: ", t)
		client.Close()
		respCh <- resp{data: t, err: err}
		close(respCh)
	}()

	return
}

func (p *plugin) register() error {
	cmd := exec.Command(p.exe, "-sima:discover")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Start()

	p.header.readAll(stdout)
	if val := p.header.get("objects"); val != "" {
		p.objs = strings.Split(val, ", ")
	}

	debug.Printf("Set kill after one second in register")
	p.killCh = time.After(1 * time.Second)
	return p.wait(cmd)
}

func (p *plugin) wait(cmd *exec.Cmd) error {
	errCh := make(chan error)
	go func(cmd *exec.Cmd) {
		errCh <- cmd.Wait()
		debug.Printf("Subprocess exited")
		close(errCh)
	}(cmd)

	for {
		select {
		case err := <-errCh:
			return err
		case <-p.killCh:
			cmd.Process.Kill()
			// TODO: Report error if this fails
		}
	}
}
