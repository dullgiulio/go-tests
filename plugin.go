package sima

import (
	"net/rpc"
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
	status pluginStatus
	header *headerReader
}

func NewPlugin(exe string) *plugin {
	return &plugin{
		exe:    exe,
		header: newHeaderReader(),
		objs:   make([]string, 0),
	}
}

func (p *plugin) String() string {
	return p.exe + " " + strings.Join(p.objs, ", ")
}

func (p *plugin) start() error {
	// TODO: Start and make killable via chan
	return nil
}

func (p *plugin) stop() error {
	// TODO: Make sure it's actually dead
	respCh := make(chan resp)
	if err := p.call("SimaRpc.Exit", 1, SimaNil, respCh); err != nil {
		close(respCh)
		return err
	}

	resp := <-respCh
	return resp.err
}

func (p *plugin) call(name string, n int, data interface{}, respCh chan<- resp) error {
	// TODO: Find which plugin contains the object, start it if necessary.
	// The client object should come from the plugin
	client, err := rpc.DialHTTP(p.proto, p.addr)
	if err != nil {
		return err
	}

	go func() {
		err = client.Call(name, n, data)
		client.Close()
		respCh <- resp{data: data, err: err}
		close(respCh)
	}()

	return nil
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

	// TODO: Must exit immediately, or kill it on timeout.

	return p.wait(cmd, 1*time.Second)
}

func (p *plugin) wait(cmd *exec.Cmd, t time.Duration) error {
	errCh := make(chan error)
	go func(cmd *exec.Cmd) {
		errCh <- cmd.Wait()
		close(errCh)
	}(cmd)

	for {
		select {
		case err := <-errCh:
			return err
		case <-time.After(t):
			cmd.Process.Kill()
			// TODO: What happens if this failed?
		}
	}
}
