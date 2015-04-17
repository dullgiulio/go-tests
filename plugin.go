package sima

type pluginStatus int

const (
	pluginStatusNone pluginStatus = iota
	pluginStatusRunning
)

type Plugin struct {
	name   string
	objs   *Objects
	status pluginStatus
	client Client
}

func NewPlugin(name string, c Client) *Plugin {
	return &Plugin{name: name, client: c}
}

func (p *Plugin) String() string {
	return p.name + " " + p.objs.String()
}

func (p *Plugin) Stop() error {
	// TODO: Make sure it's actually dead
	return p.client.Call("SimaRpc.Exit", 1, SimaNil)
}
