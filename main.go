// filepath: /path/to/ping-plugin/main.go
package main

import (
	"net/rpc"

	"github.com/AlertFlow/runner/pkg/plugins"

	"github.com/v1Flows/alertFlow/services/backend/pkg/models"

	"github.com/hashicorp/go-plugin"
)

// PingPlugin is an implementation of the Plugin interface
type AlertmanagerEndpointPlugin struct{}

func (p *AlertmanagerEndpointPlugin) Execute(request plugins.ExecuteRequest) (plugins.ExecuteResponse, error) {
	// Implement the ping logic here
	return plugins.ExecuteResponse{
		Success: true,
		Error:   "",
	}, nil
}

func (p *AlertmanagerEndpointPlugin) Info() (models.Plugins, error) {
	return models.Plugins{
		Name:    "Alertmanager",
		Type:    "endpoint",
		Version: "1.0.11",
		Author:  "JustNZ",
		Endpoints: models.PayloadEndpoints{
			ID:       "alertmanager",
			Name:     "Alertmanager",
			Endpoint: "/alertmanager",
		},
	}, nil
}

// PluginRPCServer is the RPC server for Plugin
type PluginRPCServer struct {
	Impl plugins.Plugin
}

func (s *PluginRPCServer) Execute(request plugins.ExecuteRequest, resp *plugins.ExecuteResponse) error {
	result, err := s.Impl.Execute(request)
	*resp = result
	return err
}

func (s *PluginRPCServer) Info(args interface{}, resp *models.Plugins) error {
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
