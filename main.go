// filepath: /path/to/ping-plugin/main.go
package main

import (
	"net/rpc"

	"github.com/AlertFlow/runner/pkg/models"
	"github.com/AlertFlow/runner/pkg/plugins"

	"github.com/hashicorp/go-plugin"
)

// PingPlugin is an implementation of the Plugin interface
type AlertmanagerEndpointPlugin struct{}

func (p *AlertmanagerEndpointPlugin) Execute(args map[string]string) (string, error) {
	// Implement the ping logic here
	return "Pong", nil
}

func (p *AlertmanagerEndpointPlugin) Info() (models.Plugin, error) {
	return models.Plugin{
		Name:    "Alertmanager",
		Type:    "payload_endpoint",
		Version: "1.0.11",
		Author:  "JustNZ",
		Payload: models.PayloadEndpoint{
			Name:     "Alertmanager",
			Type:     "alertmanager",
			Endpoint: "/alertmanager",
		},
	}, nil
}

// PluginRPCServer is the RPC server for Plugin
type PluginRPCServer struct {
	Impl plugins.Plugin
}

func (s *PluginRPCServer) Execute(args map[string]string, resp *string) error {
	result, err := s.Impl.Execute(args)
	*resp = result
	return err
}

func (s *PluginRPCServer) Info(args interface{}, resp *models.Plugin) error {
	result, err := s.Impl.Info()
	*resp = result
	return err
}

// PluginServer is the implementation of plugin.Plugin interface
type PluginServer struct {
	Impl plugins.Plugin
}

func (p *PluginServer) Server(*plugin.MuxBroker) (interface{}, error) {
	return &PluginRPCServer{Impl: p.Impl}, nil
}

func (p *PluginServer) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &plugins.PluginRPC{Client: c}, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "PLUGIN_MAGIC_COOKIE",
			MagicCookieValue: "hello",
		},
		Plugins: map[string]plugin.Plugin{
			"plugin": &PluginServer{Impl: &AlertmanagerEndpointPlugin{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
