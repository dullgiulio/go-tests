package sima

import (
	"errors"
	"log"
	"strings"
)

type call struct {
	obj, function string
	respCh        chan error
	args          interface{}
	resp          interface{}
}

type waitRequest struct {
	c chan struct{}
}

func newWaitRequest() *waitRequest {
	return &waitRequest{c: make(chan struct{})}
}

func (wr *waitRequest) wait() {
	<-wr.c
}

func (wr *waitRequest) done() {
	close(wr.c)
}

type registration struct {
	plugin *plugin
	objs   []string
	done   chan struct{}
}

type Manager struct {
	cancel     chan *waitRequest
	calls      chan *call
	registerCh chan registration
	plugins    map[*plugin]struct{}
	objects    map[string]*plugin
}

func NewManager() *Manager {
	m := &Manager{
		cancel:     make(chan *waitRequest),
		calls:      make(chan *call),
		registerCh: make(chan registration),
		plugins:    make(map[*plugin]struct{}),
		objects:    make(map[string]*plugin),
	}
	go m.run()
	return m
}

func (m *Manager) run() {
	for {
		select {
		case c := <-m.cancel:
			finished := make(chan struct{}, len(m.plugins))

			// Shut down all plugins
			for p := range m.plugins {
				finished <- struct{}{}

				go p.stop()
			}

			for i := 0; i < len(m.plugins); i++ {
				<-finished
			}
			c.done()
			return
		case r := <-m.registerCh:
			m.plugins[r.plugin] = struct{}{}

			// Remove all objects of this plugin if present already
			for obj, p := range m.objects {
				if p == r.plugin {
					delete(m.objects, obj)
				}
			}
			// Insert the exported objects. This way we support change of objects on upgrade.
			for _, obj := range r.objs {
				if p, ok := m.objects[obj]; ok {
					log.Print("Object ", obj, " already registered in ", p.String())
				} else {
					m.objects[obj] = r.plugin
				}
			}

			close(r.done)
		case c := <-m.calls:
			p, ok := m.objects[c.obj]
			if !ok {
				log.Print("Object ", c.obj, " not found")
				continue
			}

			p.call(c)
		}
	}
}

func (m *Manager) Call(name string, args interface{}, resp interface{}) error {
	parts := strings.SplitN(name, ".", 2)
	if parts[0] == "" || parts[1] == "" {
		return errors.New("Invalid object name")
	}

	respCh := make(chan error)
	m.calls <- &call{obj: parts[0], function: parts[1], args: args, resp: resp, respCh: respCh}
	return <-respCh
}

func (m *Manager) Stop() {
	wr := newWaitRequest()
	m.cancel <- wr
	wr.wait()
}

func (m *Manager) Register(p *plugin) {
	done := make(chan struct{})
	p.start(m.registerCh, done)
	<-done
}
