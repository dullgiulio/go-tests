package sima

import (
	"bufio"
	"errors"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type plugin struct {
	exe        string
	proto      string
	addr       string
	objs       []string
	cmd        *exec.Cmd
	client     *rpc.Client
	finishedCh chan error
	inputCh    chan string
	callsCh    chan *call
	killCh     chan *waitRequest
	over       *waitRequest
	reg        *registration
	ready      sync.RWMutex
	calls      sync.WaitGroup
	header     header
}

func NewPlugin(exe string) *plugin {
	p := &plugin{
		exe:        exe,
		objs:       make([]string, 0),
		proto:      "unix",
		inputCh:    make(chan string),
		finishedCh: make(chan error),
		callsCh:    make(chan *call),
		killCh:     make(chan *waitRequest),
		over:       newWaitRequest(),
		header:     header("sima" + randstr(5)),
	}
	// Keep this locked exclusively until the plugin is ready.
	p.ready.Lock()
	go p.run()
	return p
}

func (p *plugin) String() string {
	return p.exe + " " + strings.Join(p.objs, ", ")
}

func (p *plugin) start(reg *registration) {
	p.reg = reg
	go p.wait(exec.Command(p.exe, "-sima:prefix="+string(p.header), "-sima:proto="+p.proto))
}

func (p *plugin) call(c *call) {
	p.callsCh <- c
}

func (p *plugin) doCall(c *call) {
	// Stop and wait until ready.  This prevents calls to be
	// fired before the plugin has actually started.
	// If the start has failed, an error is returned.
	p.ready.RLock()
	defer p.ready.RUnlock()

	p.calls.Add(1)

	err := p.client.Call(c.obj+"."+c.function, c.args, c.resp)

	// TODO: This is a potential data race?
	if c.respCh != nil {
		c.respCh <- err
		close(c.respCh)
	}

	p.calls.Done()

	return
}

func (p *plugin) stop() {
	// TODO: If "graceful", wait on p.calls
	wr := newWaitRequest()
	p.killCh <- wr
	wr.wait()
	p.over.wait()
}

func (p *plugin) wait(cmd *exec.Cmd) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		p.finishedCh <- err
		return
	}
	if err := cmd.Start(); err != nil {
		p.finishedCh <- err
		return
	}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		p.inputCh <- scanner.Text()
	}

	err = cmd.Wait()
	p.finishedCh <- err
}

func (p *plugin) doKill() {
	if p.cmd == nil {
		return
	}
	if err := p.cmd.Process.Kill(); err != nil {
		log.Print("Cannot kill: ", err)
	}
}

func (p *plugin) parseReady(str string) error {
	errInvalid := errors.New("invalid ready message")
	if !strings.HasPrefix(str, "proto=") {
		return errInvalid
	}
	str = str[6:]
	s := strings.IndexByte(str, ' ')
	if s < 0 {
		return errInvalid
	}
	p.proto = str[0:s]
	str = str[s+1:]
	if !strings.HasPrefix(str, "addr=") {
		return errInvalid
	}
	p.addr = str[5:]
	return nil
}

func (p *plugin) run() {
	var errPluginUnrecoverable error

	// This channel is an alias to p.callsCh. It allows to
	// intermittedly process calls (only when we can handle them).
	var callsCh chan *call

	for {
		select {
		case c := <-callsCh:
			if errPluginUnrecoverable != nil {
				c.respCh <- errPluginUnrecoverable
			}

			// If the plugin is not ready, the call will just hang waiting.
			go p.doCall(c)
		case line := <-p.inputCh:
			if key, val := p.header.parse(line); key != "" {
				switch key {
				case "fatal":
					log.Print("Fatal: ", val)
					p.doKill()
				case "error":
					log.Print("Subprocess error: ", val)
				case "objects":
					p.objs = strings.Split(val, ", ")
					// Send the objects that we have in this plugin
					p.reg.objs = p.objs
					p.reg.done()
					p.reg = nil // XXX: Cannot reuse?
				case "ready":
					var err error

					if p.parseReady(val); err != nil {
						log.Print("Cannot parse plugin connection data: ", err)
						p.doKill()
						// TODO: Unrecoverable?
						continue
					}

					p.client, err = rpc.DialHTTP(p.proto, p.addr)
					if err != nil {
						// If we get an error after the plugin declared itself as ready,
						// the plugin is lying or there has been another problem.  In any case
						// this plugin needs to be killed and started again.
						log.Print("dial failed: ", err)

						p.doKill()

						// TODO: restart this plugin
						continue
					}

					// Remove the temp socket now that we are connected
					if p.proto == "unix" {
						if err := os.Remove(p.addr); err != nil {
							log.Print("Cannot remove temporary socket: ", err)
						}
					}

					// Start accepting calls in this loop
					callsCh = p.callsCh

					// Broadcast that we are ready to process call requests.
					p.ready.Unlock()
				}
			}
		case wr := <-p.killCh:
			// If we don't accept calls, kill immediately
			if callsCh == nil {
				p.doKill()
			} else {
				/*
					go func() {
						<-time.After(1 * time.Second)
						log.Print("Killing now")
						p.doKill()
					}()
				*/

				go p.doCall(&call{obj: "SimaRpc", function: "Exit", args: 0})
			}
			// Do not accept calls
			callsCh = nil
			// TODO: Set that we were killed
			wr.done()
		case err := <-p.finishedCh:
			if err != nil {
				if _, ok := err.(*exec.ExitError); !ok {
					log.Print("Generic error: ", err)
				}
			}

			// If we get calls now, they must hang
			p.ready.Lock()
			// And must not be served but this loop
			callsCh = nil

			p.over.done()

			// TODO: If we were not killed restart plugin
			return
		}
	}
}
