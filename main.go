// FILE: alertmanager/alertmanager.go
package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/AlertFlow/runner/pkg/models"
	"github.com/AlertFlow/runner/pkg/payloads"
	"github.com/AlertFlow/runner/pkg/protocol"
)

func main() {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		os.Exit(0)
	}()

	// Process requests
	for {
		var req protocol.Request
		if err := decoder.Decode(&req); err != nil {
			os.Exit(1)
		}

		// Handle the request
		resp := handle(req)

		if err := encoder.Encode(resp); err != nil {
			os.Exit(1)
		}
	}
}

type Receiver struct {
	Receiver string `json:"receiver"`
}

func Details() models.Plugin {
	plugin := models.Plugin{
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

	return plugin
}

func payload(body json.RawMessage) (success bool, err error) {
	// receiver := Receiver{}
	// json.Unmarshal(body, &receiver)

	payloadData := models.Payload{
		Payload:  body,
		FlowID:   "test",
		RunnerID: "test2",
		Endpoint: "alertmanager",
	}

	payloads.SendPayload(payloadData)

	return true, nil
}

func handle(req protocol.Request) protocol.Response {
	switch req.Action {
	case "details":
		return protocol.Response{
			Success: true,
			Plugin:  Details(),
		}

	case "payload":
		success, err := payload(req.Data["body"].(json.RawMessage))
		if err != nil {
			return protocol.Response{
				Success: false,
				Error:   err.Error(),
			}
		}

		return protocol.Response{
			Success: success,
			Data:    nil,
		}

	default:
		return protocol.Response{
			Success: false,
			Error:   "unknown action",
		}
	}
}
