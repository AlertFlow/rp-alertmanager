// filepath: /path/to/ping-plugin/main.go
package main

import (
	"encoding/json"
	"net/rpc"

	"github.com/AlertFlow/runner/pkg/alerts"
	"github.com/AlertFlow/runner/pkg/plugins"

	"github.com/v1Flows/alertFlow/services/backend/pkg/models"

	"github.com/hashicorp/go-plugin"
	"github.com/tidwall/gjson"
)

type Payload struct {
	Receiver string     `json:"receiver"`
	Status   string     `json:"status"`
	Origin   string     `json:"externalURL"`
	Alerts   []struct{} `json:"alerts"`
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

	payloadString := string(incPayload)

	payload := Payload{}
	json.Unmarshal(incPayload, &payload)

	alertData := models.Alerts{
		Payload:  incPayload,
		FlowID:   payload.Receiver,
		RunnerID: request.Config.Alertflow.RunnerID,
		Plugin:   "Alertmanager",
		Status:   payload.Status,
		Origin:   payload.Origin,
	}

	// search for alertname in payload
	if gjson.Get(payloadString, "commonLabels.alertname").Exists() {
		alertData.Name = gjson.Get(payloadString, "commonLabels.alertname").String()
	} else if gjson.Get(payloadString, "groupLabels.alertname").Exists() {
		alertData.Name = gjson.Get(payloadString, "groupLabels.alertname").String()
	} else {
		alertData.Name = "Unknown"
	}

	// check if we have more than one Alert
	if gjson.Get(payloadString, "alerts").Exists() {
		alerts := gjson.Get(payloadString, "alerts").Array()
		for _, alert := range alerts {
			if alert.Get("labels.alertname").Exists() {
				alertData.GroupedAlerts = append(alertData.GroupedAlerts, models.GroupedAlert{
					Name:        alert.Get("labels.alertname").String(),
					Description: alert.Get("annotations.description").String(),
					Status:      alert.Get("status").String(),
					Affected:    []string{alert.Get("labels.instance").String()},
				})
			}
		}
	} else {
		alertData.GroupedAlerts = append(alertData.GroupedAlerts, models.GroupedAlert{
			Name:        alertData.Name,
			Description: gjson.Get(payloadString, "commonAnnotations.description").String(),
			Status:      alertData.Status,
			Affected:    alertData.Affected,
		})
	}

	// get intance from payload
	if gjson.Get(payloadString, "commonLabels.instance").Exists() {
		alertData.Affected = []string{gjson.Get(payloadString, "commonLabels.instance").String()}
	} else {
		alertData.Affected = []string{"Unknown"}
	}

	// check if alert is resolved
	if payload.Status == "resolved" {
		alerts.UpdateAlert(request.Config, alertData)
	} else {
		alerts.SendAlert(request.Config, alertData)
	}

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
			Icon:     "vscode-icons:file-type-prometheus",
			Color:    "#e6522c",
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
