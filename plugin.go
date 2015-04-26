package sima

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	"strings"
	"time"
)

const internalObject = "SimaRpc"

type conn struct {
	client *rpc.Client
	err    error
	wr     *waiter
}

type waiter struct {
	c chan struct{}
}

func newWaiter() *waiter {
	return &waiter{c: make(chan struct{})}
}

func (wr *waiter) wait() {
	<-wr.c
}

func (wr *waiter) done() {
	close(wr.c)
}

func (wr *waiter) reset() {
	wr.c = make(chan struct{})
}

type objects struct {
	list []string
	err  error
	wr   *waiter
}

type Plugin struct {
	exe         string
	proto       string
	params      []string
	objs        []string
	initTimeout time.Duration
	exitTimeout time.Duration
	header      header
	objsCh      chan *objects
	connCh      chan *conn
	killCh      chan *waiter
	exitCh      chan struct{}
}

func NewPlugin(exe string, params ...string) (*Plugin, error) {
	p := &Plugin{
		exe:    exe,
		proto:  "unix",
		params: params,
		// TODO: Make user configurable
		initTimeout: 2 * time.Second,
		exitTimeout: 1 * time.Second,
		header:      header("sima" + randstr(5)),
		objs:        make([]string, 0),
		objsCh:      make(chan *objects),
		connCh:      make(chan *conn),
		killCh:      make(chan *waiter),
		exitCh:      make(chan struct{}),
	}
	return p, nil
}

func (p *Plugin) String() string {
	return p.exe + " " + strings.Join(p.objs, ", ")
}

func (p *Plugin) Start() {
	params := make([]string, len(p.params)+2)
	params[0] = "-sima:prefix=" + string(p.header)
	params[1] = "-sima:proto=" + p.proto
	for i := 0; i < len(p.params); i++ {
		params[i+2] = p.params[i]
	}
	cmd := exec.Command(p.exe, params...)
	go p.run(cmd)
}

func (p *Plugin) Stop() {
	wr := newWaiter()
	p.killCh <- wr
	wr.wait()
	p.exitCh <- struct{}{}
}

func (p *Plugin) Call(name string, args interface{}, resp interface{}) error {
	conn := &conn{wr: newWaiter()}
	p.connCh <- conn
	conn.wr.wait()

	if conn.err != nil {
		return conn.err
	}

	return conn.client.Call(name, args, resp)
}

func (p *Plugin) Objects() ([]string, error) {
	objects := &objects{wr: newWaiter()}
	p.objsCh <- objects
	objects.wr.wait()

	return objects.list, objects.err
}

type pluginCtrl struct {
	p *Plugin
	// Protocol and address for RPC
	proto, addr string
	// Unrecoverable error is used as response to calls after it happened.
	err error
	// This channel is an alias to p.connCh. It allows to
	// intermittedly process calls (only when we can handle them).
	connCh chan *conn
	// Same as above, but for objects requests
	objsCh chan *objects
	// Timeout on plugin startup time
	timeoutCh <-chan time.Time
	// Get notification from Wait on the subprocess
	waitCh chan error
	// Get output lines from subprocess
	linesCh chan string
	// Respond to a routine waiting for this mail loop to exit.
	over *waiter
	// Executable
	cmd *exec.Cmd
	// RPC client to subprocess
	client *rpc.Client
}

func newPluginCtrl(p *Plugin, cmd *exec.Cmd, t time.Duration) *pluginCtrl {
	return &pluginCtrl{
		p:         p,
		cmd:       cmd,
		timeoutCh: time.After(t),
		linesCh:   make(chan string),
		waitCh:    make(chan error),
	}
}

func (c *pluginCtrl) fatal(err error) {
	c.err = err
	c.open()
	c.kill()
}

func (c *pluginCtrl) isFatal() bool {
	return c.err != nil
}

func (c *pluginCtrl) close() {
	c.connCh = nil
	c.objsCh = nil
}

func (c *pluginCtrl) open() {
	c.connCh = c.p.connCh
	c.objsCh = c.p.objsCh
}

func (c *pluginCtrl) readOutput(r io.Reader) {
	defer close(c.linesCh)

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		c.linesCh <- scanner.Text()
	}
}

func (c *pluginCtrl) wait(cmd *exec.Cmd) {
	if cmd == nil {
		// TODO: Shouldn't happen, but reply with error
		return
	}
	defer close(c.waitCh)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		c.waitCh <- err
		return
	}
	if err := cmd.Start(); err != nil {
		c.waitCh <- err
		return
	}

	c.readOutput(stdout)

	c.waitCh <- cmd.Wait()
}

func (c *pluginCtrl) kill() {
	if c.cmd == nil {
		return
	}
	if err := c.cmd.Process.Kill(); err != nil {
		log.Print("Cannot kill: ", err)
	}
	c.cmd = nil
}

func (c *pluginCtrl) parseReady(str string) error {
	errInvalid := errors.New("invalid ready message")
	if !strings.HasPrefix(str, "proto=") {
		return errInvalid
	}
	str = str[6:]
	s := strings.IndexByte(str, ' ')
	if s < 0 {
		return errInvalid
	}
	c.proto = str[0:s]
	str = str[s+1:]
	if !strings.HasPrefix(str, "addr=") {
		return errInvalid
	}
	c.addr = str[5:]
	return nil
}

func (p *Plugin) run(cmd *exec.Cmd) {
	ctrl := newPluginCtrl(p, cmd, p.initTimeout)
	go ctrl.wait(cmd)

	for {
		select {
		case <-ctrl.timeoutCh:
			ctrl.fatal(errors.New("Registration timed out"))
		case c := <-ctrl.connCh:
			if ctrl.isFatal() {
				c.err = ctrl.err
				c.wr.done()
				continue
			}

			c.client = ctrl.client
			c.wr.done()
		case o := <-ctrl.objsCh:
			if ctrl.isFatal() {
				o.err = ctrl.err
				o.wr.done()
				continue
			}

			// Copy the list of objects for the requestor
			o.list = make([]string, len(p.objs)-1)
			for i, j := 0, 0; i < len(p.objs); i++ {
				if p.objs[i] == internalObject {
					continue
				}
				o.list[j] = p.objs[i]
				j = j + 1
			}
			o.wr.done()
		case line := <-ctrl.linesCh:
			if key, val := p.header.parse(line); key != "" {
				switch key {
				case "fatal":
					ctrl.fatal(errors.New(val))
				case "error":
					// TODO: Send to error handler
					log.Print("Subprocess error: ", val)
				case "objects":
					p.objs = strings.Split(val, ", ")
					// We will start replyin to "Objects" requests only when
					// the subprocess is ready
				case "ready":
					var err error

					if err := ctrl.parseReady(val); err != nil {
						ctrl.fatal(err)
						continue
					}

					ctrl.client, err = rpc.DialHTTP(ctrl.proto, ctrl.addr)
					if err != nil {
						ctrl.fatal(err)
						continue
					}

					// Remove the temp socket now that we are connected
					if ctrl.proto == "unix" {
						if err := os.Remove(ctrl.addr); err != nil {
							log.Print("Cannot remove temporary socket: ", err)
						}
					}

					// Defuse the timeout on ready
					ctrl.timeoutCh = nil

					// Start accepting calls
					ctrl.open()
				}
			}
		case wr := <-p.killCh:
			if ctrl.waitCh == nil {
				wr.done()
				continue
			}

			// If we don't accept calls, kill immediately
			if ctrl.connCh == nil || ctrl.client == nil {
				ctrl.kill()
			} else {
				// TODO: Improve this, should be in same routine
				go func(t time.Duration) {
					<-time.After(t)
					ctrl.kill()
				}(p.exitTimeout)

				ctrl.client.Call(internalObject+".Exit", 0, nil)
			}

			// Do not accept calls
			ctrl.close()

			// When wait on the subprocess is exited, signal back via "over"
			ctrl.over = wr
		case err := <-ctrl.waitCh:
			if err != nil {
				if _, ok := err.(*exec.ExitError); !ok {
					log.Print("Generic error: ", err)
				} else {
					ctrl.fatal(err)
				}
			}

			// Signal to whoever killed us (via killCh) that we are done
			if ctrl.over != nil {
				ctrl.over.done()
			}

			ctrl.cmd = nil
			ctrl.waitCh = nil
		case <-p.exitCh:
			return
		}
	}
}
