// Package pingo implements the basics for creating and running subprocesses
// as plugins.  The subprocesses will communicate via either TCP or Unix socket
// to implement an interface that mimics the standard RPC package.
package sima

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	errInvalidMessage      = ErrInvalidMessage(errors.New("Invalid ready message"))
	errRegistrationTimeout = ErrRegistrationTimeout(errors.New("Registration timed out"))
)

// Represents a plugin. After being created the plugin is not started or ready to run.
//
// Additional configuration (ErrorHandler and Timeout) can be set after initialization.
//
// Use Start() to make the plugin available.
type Plugin struct {
	exe         string
	proto       string
	unixdir     string
	params      []string
	initTimeout time.Duration
	exitTimeout time.Duration
	handler     ErrorHandler
	running     bool
	meta        meta
	objsCh      chan *objects
	connCh      chan *conn
	killCh      chan *waiter
	exitCh      chan struct{}
}

// NewPlugin create a new plugin ready to be started, or returns an error if the initial setup fails.
//
// The first argument specifies the protocol. It can be either set to "unix" for communication on an
// ephemeral local socket, or "tcp" for network communication on the local host (using a random
// unprivileged port.)
//
// This constructor will panic if the proto argument is neither "unix" nor "tcp".
//
// The path to the plugin executable should be absolute. Any path accepted by the "exec" package in the
// standard library is accepted and the same rules for execution are applied.
//
// Optionally some parameters might be passed to the plugin executable.
func NewPlugin(proto, path string, params ...string) *Plugin {
	if proto != "unix" && proto != "tcp" {
		panic("Invalid protocol. Specify 'unix' or 'tcp'.")
	}
	p := &Plugin{
		exe:         path,
		proto:       proto,
		params:      params,
		initTimeout: 2 * time.Second,
		exitTimeout: 2 * time.Second,
		handler:     NewDefaultErrorHandler(),
		meta:        meta("sima" + randstr(5)),
		objsCh:      make(chan *objects),
		connCh:      make(chan *conn),
		killCh:      make(chan *waiter),
		exitCh:      make(chan struct{}),
	}
	return p
}

// Set the error (and output) handler implementation.  Use this to set a custom implementation.
// By default, standard logging is used.  See ErrorHandler.
//
// Panics if called after Start.
func (p *Plugin) SetErrorHandler(h ErrorHandler) {
	if p.running {
		panic("Cannot call SetErrorHandler after Start")
	}
	p.handler = h
}

// Set the maximum time a plugin is allowed to start up and to shut down.  Empty timeout (zero)
// is not allowed, default will be used.
//
// Default is two seconds.
//
// Panics if called after Start.
func (p *Plugin) SetTimeout(t time.Duration) {
	if p.running {
		panic("Cannot call SetTimeout after Start")
	}
	if t == 0 {
		return
	}
	p.initTimeout = t
	p.exitTimeout = t
}

func (p *Plugin) SetSocketDirectory(dir string) {
	if p.running {
		panic("Cannot call SetSocketDirectory after Start")
	}
	p.unixdir = dir
}

// Default string representation
func (p *Plugin) String() string {
	return fmt.Sprintf("%s %s", p.exe, strings.Join(p.params, " "))
}

// Start will execute the plugin as a subprocess. Start will return immediately. Any first call to the
// plugin will reveal eventual errors occurred at initialization.
//
// Calls subsequent to Start will hang until the plugin has been properly initialized.
func (p *Plugin) Start() {
	p.running = true
	go p.run()
}

// Stop attemps to stop cleanly or kill the running plugin, then will free all resources.
// Stop returns when the plugin as been shut down and related routines have exited.
func (p *Plugin) Stop() {
	wr := newWaiter()
	p.killCh <- wr
	wr.wait()
	p.exitCh <- struct{}{}
}

// Call performs an RPC call to the plugin. Prior to calling Call, the plugin must have been
// initialized by calling Start.
//
// Call will hang until a plugin has been initialized; it will return any error that happens
// either when performing the call or during plugin initialization via Start.
//
// Please refer to the "rpc" package from the standard library for more information on the
// semantics of this function.
func (p *Plugin) Call(name string, args interface{}, resp interface{}) error {
	conn := &conn{wr: newWaiter()}
	p.connCh <- conn
	conn.wr.wait()

	if conn.err != nil {
		return conn.err
	}

	return conn.client.Call(name, args, resp)
}

// Objects returns a list of the exported objects from the plugin. Exported objects used
// internally are not reported.
//
// Like Call, Objects returns any error happened on initialization if called after Start.
func (p *Plugin) Objects() ([]string, error) {
	objects := &objects{wr: newWaiter()}
	p.objsCh <- objects
	objects.wr.wait()

	return objects.list, objects.err
}

// ErrorHandler is the interface used by Plugin to report non-fatal errors and any other
// output from the plugin.
//
// A default implementation is provided and used if none is specified on plugin creation.
type ErrorHandler interface {
	// Error is called whenever a non-fatal error occurs in the plugin subprocess.
	Error(error)
	// Print is called for each line of output received from the plugin subprocess.
	Print(interface{})
}

// Default error handler implementation. Uses the default logging facility from the
// Go standard library.
type DefaultErrorHandler struct{}

// Constructor for default error handler.
func NewDefaultErrorHandler() *DefaultErrorHandler {
	return &DefaultErrorHandler{}
}

// Log via default standard library facility prepending the "error: " string.
func (e *DefaultErrorHandler) Error(err error) {
	log.Print("error: ", err)
}

// Log via default standard library facility.
func (e *DefaultErrorHandler) Print(s interface{}) {
	log.Print(s)
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
	proc *os.Process
	// RPC client to subprocess
	client *rpc.Client
}

func newCtrl(p *Plugin, t time.Duration) *ctrl {
	return &ctrl{
		p:         p,
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

func (c *ctrl) wait(pidCh chan<- int, exe string, params ...string) {
	defer close(c.waitCh)
	defer close(pidCh)

	cmd := exec.Command(exe, params...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		c.waitCh <- err
		return
	}
	if err := cmd.Start(); err != nil {
		c.waitCh <- err
		return
	}
	pidCh <- cmd.Process.Pid

	c.readOutput(stdout)

	c.waitCh <- cmd.Wait()
}

func (c *ctrl) kill() {
	if c.proc == nil {
		return
	}
	// Ignore errors here because Kill might have been called after
	// process has ended.
	c.proc.Kill()
	c.proc = nil
}

func (c *ctrl) parseReady(str string) error {
	if !strings.HasPrefix(str, "proto=") {
		return errInvalidMessage
	}
	str = str[6:]
	s := strings.IndexByte(str, ' ')
	if s < 0 {
		return errInvalidMessage
	}
	c.proto = str[0:s]
	str = str[s+1:]
	if !strings.HasPrefix(str, "addr=") {
		return errInvalidMessage
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

func (p *Plugin) run() {
	params := make([]string, len(p.params)+2)
	params[0] = "-sima:prefix=" + string(p.meta)
	params[1] = "-sima:proto=" + p.proto
	if p.proto == "unix" && p.unixdir != "" {
		params = append(params, "-sima:unixdir="+p.unixdir)
	}
	offset := len(params)
	for i := 0; i < len(p.params); i++ {
		params[i+offset] = p.params[i]
	}
	c := newCtrl(p, p.initTimeout)

	pidCh := make(chan int)
	go c.wait(pidCh, p.exe, params...)
	pid := <-pidCh

	if pid != 0 {
		if proc, err := os.FindProcess(pid); err == nil {
			c.proc = proc
		}
	}

	for {
		select {
		case <-c.timeoutCh:
			c.fatal(errRegistrationTimeout)
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
			key, val := p.meta.parse(line)
			if key == "" {
				if val != "" {
					p.handler.Print(val)
				}
				continue
			}

			switch key {
			case "fatal":
				if err := parseError(val); err != nil {
					c.fatal(err)
				} else {
					c.fatal(errors.New(val))
				}
			case "error":
				if err := parseError(val); err != nil {
					p.handler.Print(err)
				} else {
					p.handler.Print(errors.New(val))
				}
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
				// Be sure to kill the process if it doesn't obey Exit.
				go func(pid int, t time.Duration) {
					<-time.After(t)

					if proc, err := os.FindProcess(pid); err == nil {
						proc.Kill()
					}
				}(pid, p.exitTimeout)

				c.client.Call(internalObject+".Exit", 0, nil)
			}

			// Do not accept calls
			c.close()

			// When wait on the subprocess is exited, signal back via "over"
			c.over = wr
		case err := <-c.waitCh:
			if err != nil {
				if _, ok := err.(*exec.ExitError); !ok {
					p.handler.Error(err)
				} else {
					c.fatal(err)
				}
			}

			// Signal to whoever killed us (via killCh) that we are done
			if c.over != nil {
				c.over.done()
			}

			c.proc = nil
			c.waitCh = nil
			c.linesCh = nil
		case <-p.exitCh:
			return
		}
	}
}
