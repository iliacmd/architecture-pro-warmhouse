package handlers

import (
	"context"
	"device/db"
	"device/models"
	"device/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	IntegrationService *services.IntegrationService
	MetricService      *services.MetricService
	DB                 *db.DB
}

func NewHandler(ins *services.IntegrationService, ms *services.MetricService, db *db.DB) *Handler {
	return &Handler{
		IntegrationService: ins,
		MetricService:      ms,
		DB:                 db,
	}
}

// RegisterRoutes registers the sensor routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	devices := router.Group("/devices")
	{
		devices.POST(":id/commands", h.CreateCommand)
		devices.POST("", h.CreateDevice)
		devices.GET("", h.GetDevices)
	}
}

func (h *Handler) GetDevices(c *gin.Context) {
	devices, err := h.DB.GetDevices(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, devices)
}

func (h *Handler) CreateDevice(c *gin.Context) {
	var deviceCreate models.Device
	if err := c.ShouldBindJSON(&deviceCreate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device, err := h.DB.CreateDevice(context.Background(), deviceCreate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = h.MetricService.Publish(context.TODO(), models.TelemetryData{
		Time:            time.Time{},
		DeviceID:        device.ID,
		TelemetryTypeID: models.DeviceStatusTelemetryTypeID,
		StringValue:     "active",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, device)
}

func (h *Handler) CreateCommand(c *gin.Context) {
	var command models.Command

	if err := c.ShouldBindJSON(&command); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}

	payload := map[string]any{}

	if command.CommandPayload != nil {
		payload = command.CommandPayload
	}

	v, ok := payload["unit"].(string)
	if !ok || v == "" {
		payload["unit"] = "celsius"
	}

	reqCmd := services.CreateCommandRequest{
		DeviceID:       command.DeviceID,
		DeviceProtocol: "http",
		Type:           command.Type,
		CommandPayload: payload,
	}

	resp, err := h.IntegrationService.CreateCommand(reqCmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
