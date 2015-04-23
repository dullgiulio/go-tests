package sima

import (
	"bufio"
	"io"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type pluginStatus int

const (
	pluginStatusNone pluginStatus = iota
	pluginStatusStarting
	pluginStatusRunning
)

type plugin struct {
	exe        string
	proto      string
	addr       string
	objs       []string
	cmd        *exec.Cmd
	killCh     <-chan time.Time
	finishedCh chan error
	inputCh    chan string
	readyCh    chan struct{}
	callsCh    chan call
	status     pluginStatus
	wg         sync.WaitGroup
}

func NewPlugin(exe string) *plugin {
	p := &plugin{
		exe:        exe,
		objs:       make([]string, 0),
		proto:      "unix",
		addr:       "test",
		inputCh:    make(chan string),
		finishedCh: make(chan error),
		readyCh:    make(chan struct{}),
		callsCh:    make(chan call),
	}
	go p.run()
	return p
}

func (p *plugin) String() string {
	return p.exe + " " + strings.Join(p.objs, ", ")
}

func (p *plugin) readOutput(r io.Reader) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		p.inputCh <- scanner.Text()
	}
}

func (p *plugin) start() error {
	return p.exec("-sima:proto="+p.proto, "-sima:addr="+p.addr)
}

func (p *plugin) register() error {
	p.killCh = time.After(1 * time.Second)
	return p.exec("-sima:discover")
}

func (p *plugin) exec(args ...string) error {
	p.cmd = exec.Command(p.exe, args...)
	stdout, err := p.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := p.cmd.Start(); err != nil {
		return err
	}

	p.killCh = nil

	go p.readOutput(stdout)
	go p.wait()

	return nil
}

func (p *plugin) call(c call) {
	p.callsCh <- c
}

func (p *plugin) doCall(c call) {
	<-p.readyCh

	p.wg.Add(1)

	// TODO: Find which plugin contains the object, start it if necessary.
	// The client object should come from the plugin
	client, err := rpc.DialHTTP(p.proto, p.addr)
	if err != nil {
		log.Print("dial failed: ", err)
		c.respCh <- resp{err: err}
		p.wg.Done()
		return
	}

	// Remove the temp socket now that we are connected
	if p.proto == "unix" {
		os.Remove(p.addr)
	}

	var r string

	err = client.Call(c.obj+"."+c.function, c.args, &r)
	log.Print(r)
	client.Close()

	c.respCh <- resp{data: r, err: err}
	close(c.respCh)

	p.wg.Done()

	return
}

func (p *plugin) stop() error {
	respCh := make(chan resp)
	p.call(call{obj: "SimaRpc", function: "Exit", args: 0, respCh: respCh})

	p.killCh = time.After(1 * time.Second)

	resp := <-respCh
	close(respCh)

	return resp.err
}

func (p *plugin) wait() {
	p.finishedCh <- p.cmd.Wait()
}

func (p *plugin) run() {
	header := header("sima")

	for {
		select {
		case c := <-p.callsCh:
			log.Print("got call")

			if p.status == pluginStatusNone {
				if err := p.start(); err != nil {
					log.Print(err)
					continue
				}

				p.status = pluginStatusStarting
			}

			go p.doCall(c)
		case line := <-p.inputCh:
			log.Print("line: ", line)
			if key, val := header.parse(line); key != "" {
				switch key {
				case "error":
					log.Print("Subprocess error: ", val)
				case "objects":
					log.Print("Objects: ", val)
					p.objs = strings.Split(val, ", ")
				case "ready":
					// Set plugin ready
					p.status = pluginStatusRunning

					log.Print("ready!")

					// Broadcast that we are ready to process call requests.
					close(p.readyCh)
				}
			}
		case <-p.killCh:
			// TODO: Should we p.wg.Wait() here to avoid closing on calls?
			log.Print("Killing subprocess")
			if err := p.cmd.Process.Kill(); err != nil {
				log.Print("Cannot kill: ", err)
			}
		case err := <-p.finishedCh:
			log.Print("finished")
			if err != nil {
				log.Print("Generic error: ", err)
			}
			p.status = pluginStatusNone
		}
	}
}
