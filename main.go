// filepath: /path/to/ping-plugin/main.go
package main

import (
	"encoding/json"
	"net/rpc"

	"github.com/AlertFlow/runner/pkg/alerts"
	"github.com/AlertFlow/runner/pkg/plugins"

	"github.com/v1Flows/alertFlow/services/backend/pkg/models"

	"github.com/hashicorp/go-plugin"
)

type Receiver struct {
	Receiver string `json:"receiver"`
}

// AlertmanagerEndpointPlugin is an implementation of the Plugin interface
type AlertmanagerEndpointPlugin struct{}

func (p *AlertmanagerEndpointPlugin) ExecuteTask(request plugins.ExecuteTaskRequest) (plugins.Response, error) {
	return plugins.Response{
		Success: false,
	}, nil
}

func (p *AlertmanagerEndpointPlugin) HandlePayload(request plugins.PayloadHandlerRequest) (plugins.Response, error) {
	incPayload := request.Body

	receiver := Receiver{}
	json.Unmarshal(incPayload, &receiver)

	alertData := models.Alerts{
		Payload:  incPayload,
		FlowID:   receiver.Receiver,
		RunnerID: request.Config.Alertflow.RunnerID,
		Endpoint: "alertmanager",
	}

	alerts.SendAlert(request.Config, alertData)

	return plugins.Response{
		Success: true,
	}, nil
}

func (p *AlertmanagerEndpointPlugin) Info() (models.Plugins, error) {
	return models.Plugins{
		Name:    "Alertmanager",
		Type:    "endpoint",
		Version: "1.1.0",
		Author:  "JustNZ",
		Endpoints: models.AlertEndpoints{
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

func (s *PluginRPCServer) ExecuteTask(request plugins.ExecuteTaskRequest, resp *plugins.Response) error {
	result, err := s.Impl.ExecuteTask(request)
	*resp = result
	return err
}

func (s *PluginRPCServer) HandlePayload(request plugins.PayloadHandlerRequest, resp *plugins.Response) error {
	result, err := s.Impl.HandlePayload(request)
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
