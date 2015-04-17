package sima

import (
	"log"
	"net/rpc"
)

type action int

const (
	actionPluginDebug action = iota
	actionPluginRun
	actionPluginRegister
	actionPluginStop
	actionStop
)

type event struct {
	plugin *Plugin
	action action
}

type Manager struct {
	events chan event
	plugins map[string]*Plugin
	callback func()
}

func NewManager(c func()) *Manager {
	m := &Manager{
		events: make(chan event),
		plugins: make(map[string]*Plugin),
		callback: c,
	}
	go m.run()
	return m
}

func (m *Manager) run() {
	for e := range m.events {
		var err error

		switch e.action {
		case actionPluginRegister:
			debug.Printf("Register plugin %s", e.plugin.name)
			err = m.registerPlugin(e.plugin)
		case actionPluginRun:
			e.plugin.status = pluginStatusRunning
			go e.plugin.run()
		case actionPluginStop:
			// TODO: Make rpc call to shut down.
			e.plugin.status = pluginStatusNone
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
			// TODO: Move this to an errors channel
			log.Print(err)
			err = nil
		}
	}
}

func (m *Manager) Stop() {
	m.events <- event{ action: actionStop }
}

func (m *Manager) Debug(p *Plugin) {
	m.events <- event{ plugin: p}
}

func (m *Manager) RegisterPlugin(p *Plugin) {
	m.events <- event{ plugin: p, action: actionPluginRegister}
}

func (m *Manager) registerPlugin(p *Plugin) error {
	m.plugins[p.name] = p

	client, err := rpc.DialHTTP("tcp", "127.0.0.1:8888")
	if err != nil {
		return err
	}

	var unused int
	var methods Methods

	if err := client.Call("Plugin.SimaPluginRegister", unused, &methods); err != nil {
		return err
	}

	p.methods = &methods

	// TODO: Save discovered methods.
	return nil
}
