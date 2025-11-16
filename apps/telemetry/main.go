package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rabbitmq/amqp091-go"

	"telemetry/db"
	"telemetry/http/httpHandlers"
	"telemetry/rabbit/handlers"
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

	// Set up database connection
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/telemetry")
	database, err := db.New(dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer database.Close()

	log.Println("Connected to database successfully")

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
	handler := httpHandlers.NewHandler()
	handler.RegisterRoutes(apiRoutes)

	// Start server
	srv := &http.Server{
		Addr:    getEnv("PORT", ":8080"),
		Handler: router,
	}

	go func() {
		startConsumer(ch, queueName, &handlers.DeviceMetricHandler{DB: database})
	}()

	// Start the server in a goroutine
	go func() {
		log.Printf("Server starting on %s\n", srv.Addr)
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

func startConsumer(ch *amqp091.Channel, queueName string, handler *handlers.DeviceMetricHandler) {
	msgs, err := ch.Consume(
		queueName,
		"",
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		log.Fatalf("failed to register a consumer: %v", err)
	}

	log.Println("Consumer started...")

	for m := range msgs {
		if err = handler.Handle(m); err != nil {
			log.Printf("Error handling message: %v\n", err)
		}
	}
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
