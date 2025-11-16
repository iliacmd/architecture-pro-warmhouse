package main

import (
	"context"
	"errors"
	"fmt"
	"integration/services"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rabbitmq/amqp091-go"

	"integration/handlers"
)

func main() {
	queueName := getEnv("RABBITMQ_METRICS_QUEUE", "device_metrics")

	ch, closeFn, err := newRabbit(
		getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672"),
		queueName,
	)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ channel: %v", err)
	}
	defer func() {
		_ = closeFn()
	}()

	// Initialize temperature service
	metricService := services.NewMetricService(ch, queueName)
	log.Printf("Metric service initialized\n")

	// Initialize temperature service
	temperatureAPIURL := getEnv("TEMPERATURE_API_URL", "http://temperature-api:8080")
	temperatureService := services.NewTemperatureService(temperatureAPIURL)
	log.Printf("Temperature service initialized with API URL: %s\n", temperatureAPIURL)

	// Initialize router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// API routes
	apiRoutes := router.Group("/api/v1")

	// Register sensor routes
	handler := handlers.NewHandler(temperatureService, metricService)
	handler.RegisterRoutes(apiRoutes)

	// Start server
	srv := &http.Server{
		Addr:    getEnv("PORT", ":8080"),
		Handler: router,
	}

	// Start the server in a goroutine
	go func() {
		log.Printf("Server starting on %s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v\n", err)
	}

	log.Println("Server exited properly")
}

func newRabbit(uri, queueName string) (*amqp091.Channel, func() error, error) {
	conn, err := amqp091.Dial(uri)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, conn.Close, fmt.Errorf("failed to open a channel: %v", err)
	}

	closeFunc := func() error {
		if err = ch.Close(); err != nil {
			return err
		}

		if err = conn.Close(); err != nil {
			return err
		}

		return nil
	}

	// --- Создаем очередь ---
	_, err = ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return nil, closeFunc, fmt.Errorf("failed to declare a queue: %v", err)
	}

	log.Println("Queue initialized:", queueName)

	return ch, closeFunc, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
