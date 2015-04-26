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

var (
	ErrInvalidMessage      = errors.New("Invalid ready message")
	ErrRegistrationTimeout = errors.New("Registration timed out")
)

type Plugin struct {
	exe         string
	proto       string
	params      []string
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
		objsCh:      make(chan *objects),
		connCh:      make(chan *conn),
		killCh:      make(chan *waiter),
		exitCh:      make(chan struct{}),
	}
	return p, nil
}

func (p *Plugin) String() string {
	return p.exe
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

type ctrl struct {
	p    *Plugin
	objs []string
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

func newCtrl(p *Plugin, cmd *exec.Cmd, t time.Duration) *ctrl {
	return &ctrl{
		p:         p,
		cmd:       cmd,
		timeoutCh: time.After(t),
		linesCh:   make(chan string),
		waitCh:    make(chan error),
	}
}

func (c *ctrl) fatal(err error) {
	c.err = err
	c.open()
	c.kill()
}

func (c *ctrl) isFatal() bool {
	return c.err != nil
}

func (c *ctrl) close() {
	c.connCh = nil
	c.objsCh = nil
}

func (c *ctrl) open() {
	c.connCh = c.p.connCh
	c.objsCh = c.p.objsCh
}

func (c *ctrl) ready(val string) bool {
	var err error

	if err := c.parseReady(val); err != nil {
		c.fatal(err)
		return false
	}

	c.client, err = rpc.DialHTTP(c.proto, c.addr)
	if err != nil {
		c.fatal(err)
		return false
	}

	// Remove the temp socket now that we are connected
	if c.proto == "unix" {
		if err := os.Remove(c.addr); err != nil {
			log.Print("Cannot remove temporary socket: ", err)
		}
	}

	// Defuse the timeout on ready
	c.timeoutCh = nil

	return true
}

func (c *ctrl) readOutput(r io.Reader) {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		c.linesCh <- scanner.Text()
	}
}

func (c *ctrl) wait(cmd *exec.Cmd) {
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

func (c *ctrl) kill() {
	if c.cmd == nil {
		return
	}
	if err := c.cmd.Process.Kill(); err != nil {
		// TODO: Send to error handler
		log.Print("Cannot kill: ", err)
	}
	c.cmd = nil
}

func (c *ctrl) parseReady(str string) error {
	if !strings.HasPrefix(str, "proto=") {
		return ErrInvalidMessage
	}
	str = str[6:]
	s := strings.IndexByte(str, ' ')
	if s < 0 {
		return ErrInvalidMessage
	}
	c.proto = str[0:s]
	str = str[s+1:]
	if !strings.HasPrefix(str, "addr=") {
		return ErrInvalidMessage
	}
	c.addr = str[5:]
	return nil
}

// Copy the list of objects for the requestor
func (c *ctrl) objects() []string {
	list := make([]string, len(c.objs)-1)
	for i, j := 0, 0; i < len(c.objs); i++ {
		if c.objs[i] == internalObject {
			continue
		}
		list[j] = c.objs[i]
		j = j + 1
	}
	return list
}

func (p *Plugin) run(cmd *exec.Cmd) {
	c := newCtrl(p, cmd, p.initTimeout)
	go c.wait(cmd)

	for {
		select {
		case <-c.timeoutCh:
			c.fatal(ErrRegistrationTimeout)
		case r := <-c.connCh:
			if c.isFatal() {
				r.err = c.err
				r.wr.done()
				continue
			}

			r.client = c.client
			r.wr.done()
		case o := <-c.objsCh:
			if c.isFatal() {
				o.err = c.err
				o.wr.done()
				continue
			}

			o.list = c.objects()
			o.wr.done()
		case line := <-c.linesCh:
			key, val := p.header.parse(line)
			if key == "" {
				if val != "" {
					// TODO: Send to error handler
					log.Print("Subprocess error: ", val)
				}
				continue
			}

			switch key {
			case "fatal":
				c.fatal(errors.New(val))
			case "error":
				// TODO: Send to error handler
				log.Print("Subprocess error: ", val)
			case "objects":
				c.objs = strings.Split(val, ", ")
			case "ready":
				if !c.ready(val) {
					continue
				}
				// Start accepting calls
				c.open()
			}
		case wr := <-p.killCh:
			if c.waitCh == nil {
				wr.done()
				continue
			}

			// If we don't accept calls, kill immediately
			if c.connCh == nil || c.client == nil {
				c.kill()
			} else {
				// TODO: Improve this, should be in same routine
				go func(t time.Duration) {
					<-time.After(t)
					c.kill()
				}(p.exitTimeout)

				c.client.Call(internalObject+".Exit", 0, nil)
			}

			// Do not accept calls
			c.close()

			// When wait on the subprocess is exited, signal back via "over"
			c.over = wr
		case err := <-c.waitCh:
			if err != nil {
				if _, ok := err.(*exec.ExitError); !ok {
					log.Print("Generic error: ", err)
				} else {
					c.fatal(err)
				}
			}

			// Signal to whoever killed us (via killCh) that we are done
			if c.over != nil {
				c.over.done()
			}

			c.cmd = nil
			c.waitCh = nil
			c.linesCh = nil
		case <-p.exitCh:
			return
		}
	}
}
