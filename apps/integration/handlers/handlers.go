package handlers

import (
	"context"
	"integration/models"
	"integration/services"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	TemperatureService *services.TemperatureService
	MetricService      *services.MetricService
}

func NewHandler(ts *services.TemperatureService, ms *services.MetricService) *Handler {
	return &Handler{
		TemperatureService: ts,
		MetricService:      ms,
	}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	handler := router.Group("/commands")
	{
		handler.POST("", h.CreateCommand)
	}
}

func (h *Handler) CreateCommand(c *gin.Context) {
	var command models.Command

	if err := c.ShouldBindJSON(&command); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	switch command.Type {
	case "get_temperature":
		resp, err := h.TemperatureService.GetTemperatureByID(strconv.Itoa(command.DeviceID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)

		err = h.MetricService.Publish(context.TODO(), models.TelemetryData{
			DeviceID:        command.DeviceID,
			TelemetryTypeID: models.TemperatureTelemetryTypeID,
			NumericValue:    resp.Value,
		})
		if err != nil {
			log.Printf("Error publishing metrics: %v", err)
		}

		log.Print("Get temperature telemetry published")
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unknown command"})
	}

	return
}
