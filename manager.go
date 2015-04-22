package sima

import (
	"errors"
	"log"
	"strings"
)

type action int

const (
	actionPluginDebug action = iota
	actionPluginRegister
	actionPluginUnregister
	actionStop
)

type resp struct {
	data interface{}
	err  error
}

type call struct {
	obj, function string
	respCh        chan resp
	data          interface{}
	n             int
}

type event struct {
	plugin *plugin
	action action
}

type Manager struct {
	events   chan event
	calls    chan call
	plugins  map[string]*plugin
	callback func()
}

func NewManager(c func()) *Manager {
	m := &Manager{
		events:   make(chan event),
		calls:    make(chan call),
		plugins:  make(map[string]*plugin),
		callback: c,
	}
	go m.run()
	return m
}

func (m *Manager) run() {
	for {
		select {
		case e := <-m.events:
			var err error

			switch e.action {
			case actionPluginRegister:
				debug.Printf("Register plugin %s", e.plugin.exe)
				err = e.plugin.register()
				for _, obj := range e.plugin.objs {
					if p, ok := m.plugins[obj]; ok {
						log.Print("Object ", obj, " already registered in ", p.String())
					}
					m.plugins[obj] = e.plugin
				}
			case actionPluginUnregister:
				debug.Printf("Unregister plugin %s", e.plugin.exe)
				e.plugin.stop()

				for _, obj := range e.plugin.objs {
					delete(m.plugins, obj)
				}
			case actionStop:
				if m.callback != nil {
					debug.Printf("callback and exit")
					m.callback()
				}
				return
			default:
				debug.Printf("Plugin: %v", e.plugin)
			}

			if err != nil {
				log.Print(err)
				err = nil
			}
		case c := <-m.calls:
			p, ok := m.plugins[c.obj]
			if !ok {
				log.Print("Object ", c.obj, " not found")
				continue
			}

			// If plugin is not started, start it.
			p.start()
			// Try making the call (in new routine)
			p.call(c.obj+"."+c.function, c.n, c.data, c.respCh)
		}
	}
}

func (m *Manager) Call(name string, n int, data interface{}) error {
	parts := strings.SplitN(name, ".", 2)
	if parts[0] == "" || parts[1] == "" {
		return errors.New("Invalid object name")
	}

	respCh := make(chan resp)
	m.calls <- call{obj: parts[0], function: parts[1], n: n, data: data, respCh: respCh}
	resp := <-respCh

	data = resp.data
	return resp.err
}

func (m *Manager) Stop() {
	m.events <- event{action: actionStop}
}

func (m *Manager) Debug(p *plugin) {
	m.events <- event{plugin: p}
}

func (m *Manager) Register(p *plugin) {
	m.events <- event{plugin: p, action: actionPluginRegister}
}
