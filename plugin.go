// FILE: alertmanager/alertmanager.go
package main

import (
	"encoding/json"
	"io"
	"net/http"

	"gitlab.justlab.xyz/alertflow-public/runner/config"
	"gitlab.justlab.xyz/alertflow-public/runner/pkg/models"
	"gitlab.justlab.xyz/alertflow-public/runner/pkg/payloads"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Receiver struct {
	Receiver string `json:"receiver"`
}

type Alertmanager struct{}

func (p *Alertmanager) Init() models.Plugin {
	return models.Plugin{
		Name:    "Alertmanager",
		Type:    "payload_endpoint",
		Version: "1.0.4",
		Creator: "JustNZ",
	}
}

func (p *Alertmanager) Details() models.PluginDetails {
	return models.PluginDetails{
		Payload: models.PayloadInjector{
			Name:     "Alertmanager",
			Type:     "alertmanager",
			Endpoint: "/alertmanager",
		},
	}
}

func (h *Alertmanager) Handle(context *gin.Context) {
	log.Info("Received Alertmanager Payload")
	incPayload, err := io.ReadAll(context.Request.Body)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read request body",
		})
		return
	}

	receiver := Receiver{}
	json.Unmarshal(incPayload, &receiver)

	payloadData := models.Payload{
		Payload:  incPayload,
		FlowID:   receiver.Receiver,
		RunnerID: config.Config.RunnerID,
		Endpoint: "alertmanager",
	}

	payloads.SendPayload(payloadData)
}

var Plugin Alertmanager
