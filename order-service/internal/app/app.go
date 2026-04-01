package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"order-service/internal/repository"
	transporthttp "order-service/internal/transport/http"
	"order-service/internal/usecase"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// Config holds configuration for the application.
type Config struct {
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	ServerPort      string
	PaymentBaseURL  string
	PaymentTimeout  time.Duration
}

// LoadConfig reads configuration from environment variables with defaults.
func LoadConfig() *Config {
	return &Config{
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5440"),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", "postgres"),
		DBName:         getEnv("DB_NAME", "order_db"),
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		PaymentBaseURL: getEnv("PAYMENT_BASE_URL", "http://localhost:8081"),
		PaymentTimeout: 2 * time.Second, // Max 2 seconds timeout as required
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// App holds the application components.
type App struct {
	Config *Config
	DB     *sql.DB
	Router *gin.Engine
}

// NewApp creates and wires the application (Composition Root logic).
func NewApp(cfg *Config) (*App, error) {
	// 1. Connect to database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	log.Println("Connected to Order database")

	// 2. Create repository (implements domain.OrderRepository)
	orderRepo := repository.NewPostgresOrderRepository(db)

	// 3. Create HTTP client with timeout for Payment Service
	httpClient := &http.Client{
		Timeout: cfg.PaymentTimeout,
	}

	// 4. Create payment client (implements domain.PaymentClient)
	paymentClient := transporthttp.NewPaymentHTTPClient(httpClient, cfg.PaymentBaseURL)

	// 5. Create use case with injected dependencies
	orderUseCase := usecase.NewOrderUseCase(orderRepo, paymentClient)

	// 6. Create HTTP handler
	handler := transporthttp.NewOrderHandler(orderUseCase)

	// 7. Set up Gin router
	router := gin.Default()
	handler.RegisterRoutes(router)

	return &App{
		Config: cfg,
		DB:     db,
		Router: router,
	}, nil
}

// Run starts the HTTP server.
func (a *App) Run() error {
	addr := ":" + a.Config.ServerPort
	log.Printf("Order Service starting on %s", addr)
	return a.Router.Run(addr)
}

// Close cleans up resources.
func (a *App) Close() {
	if a.DB != nil {
		a.DB.Close()
	}
}
