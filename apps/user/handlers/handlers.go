package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Handler struct {
	secretKey string
}

func NewHandler(secretKey string) *Handler {
	return &Handler{
		secretKey,
	}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	handler := router.Group("/user")
	{
		handler.POST("login", h.Login)
		handler.POST("register", h.Register)
	}
}

func (h *Handler) Login(c *gin.Context) {
	claims := jwt.MapClaims{
		"id":       1,
		"role":     "user",
		"house_id": 1,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(h.secretKey))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func (h *Handler) Register(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "not allowed"})
}
