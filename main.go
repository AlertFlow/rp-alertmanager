// filepath: /path/to/ping-plugin/main.go
package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/rpc"

	"github.com/AlertFlow/runner/pkg/payloads"
	"github.com/AlertFlow/runner/pkg/plugins"

	"github.com/v1Flows/alertFlow/services/backend/pkg/models"

	"github.com/gin-gonic/gin"
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
		Error:   "Not implemented",
	}, nil
}

func (p *AlertmanagerEndpointPlugin) HandlePayload(request plugins.PayloadHandlerRequest) (plugins.Response, error) {
	context := request.Context

	incPayload, err := io.ReadAll(context.Request.Body)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read request body",
		})
		return plugins.Response{
			Success: false,
			Error:   "Failed to read request body",
		}, nil
	}

	receiver := Receiver{}
	json.Unmarshal(incPayload, &receiver)

	payloadData := models.Payloads{
		Payload:  incPayload,
		FlowID:   receiver.Receiver,
		RunnerID: request.Config.Alertflow.RunnerID,
		Endpoint: "alertmanager",
	}

	payloads.SendPayload(request.Config, payloadData)

	return plugins.Response{
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

func (s *PluginRPCServer) ExecuteTask(request plugins.ExecuteTaskRequest, resp *plugins.Response) error {
	result, err := s.Impl.ExecuteTask(request)
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
