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
	args          interface{}
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
				err = e.plugin.register()
				/*
					                for _, obj := range e.plugin.objs {
										if p, ok := m.plugins[obj]; ok {
											log.Print("Object ", obj, " already registered in ", p.String())
										}
										m.plugins[obj] = e.plugin
									}
				*/
				m.plugins["Plugin"] = e.plugin
			case actionPluginUnregister:
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
			}
		case c := <-m.calls:
			p, ok := m.plugins[c.obj]
			if !ok {
				log.Print("Object ", c.obj, " not found")
				continue
			}

			p.call(c)
		}
	}
}

func (m *Manager) Call(name string, args interface{}) (interface{}, error) {
	parts := strings.SplitN(name, ".", 2)
	if parts[0] == "" || parts[1] == "" {
		return nil, errors.New("Invalid object name")
	}

	respCh := make(chan resp)
	m.calls <- call{obj: parts[0], function: parts[1], args: args, respCh: respCh}
	result := <-respCh

	return result.data, result.err
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
