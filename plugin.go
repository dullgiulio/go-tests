package sima

type pluginStatus int

const (
	pluginStatusNone pluginStatus = iota
	pluginStatusRunning
)

type Plugin struct {
	name string
	methods *Methods
	status pluginStatus
	// transport -> interface
}

func NewPlugin(name string) *Plugin {
	return &Plugin{ name: name }
}

func (p *Plugin) run() {
	debug.Printf("Plugin running")
}

func (p *Plugin) String() string {
	return p.name + " " + p.methods.String()
}
