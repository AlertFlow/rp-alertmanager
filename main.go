package main

import (
	"context"

	"github.com/AlertFlow/runner/pkg/models"
	"github.com/AlertFlow/runner/pkg/plugin"
	goplugin "github.com/hashicorp/go-plugin"
)

type AlertmanagerEndpointPlugin struct{}

func (p *AlertmanagerEndpointPlugin) Execute(ctx context.Context, req *plugin.Request) (*plugin.Response, error) {
	return &plugin.Response{
		Output:  "Processed: " + req.Input,
		Success: true,
	}, nil
}

func (p *AlertmanagerEndpointPlugin) StreamUpdates(req *plugin.Request, stream plugin.Plugin_StreamUpdatesServer) error {
	// Your streaming logic here
	updates := []string{"Starting", "Processing", "Completed"}
	for i, status := range updates {
		if err := stream.Send(&plugin.StatusUpdate{
			Status:   status,
			Progress: int32((i + 1) * 33),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (p *AlertmanagerEndpointPlugin) Details() *models.Plugin {
	return &models.Plugin{
		Name:    "Alertmanager",
		Type:    "payload_endpoint",
		Version: "1.0.11",
		Author:  "JustNZ",
		Payload: models.PayloadEndpoint{
			Name:     "Alertmanager",
			Type:     "alertmanager",
			Endpoint: "/alertmanager",
		},
	}
}

func main() {
	goplugin.Serve(&goplugin.ServeConfig{
		HandshakeConfig: plugin.Handshake,
		Plugins: map[string]goplugin.Plugin{
			"example_plugin": &plugin.GRPCPlugin{
				Impl: &AlertmanagerEndpointPlugin{},
			},
		},
		GRPCServer: goplugin.DefaultGRPCServer,
	})
}
