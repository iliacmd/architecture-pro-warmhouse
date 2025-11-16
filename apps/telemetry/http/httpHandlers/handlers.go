package httpHandlers

import (
	"github.com/gin-gonic/gin"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

// RegisterRoutes registers the sensor routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	_ = router.Group("/handler")
}
