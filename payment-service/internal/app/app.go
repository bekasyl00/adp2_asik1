package app

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"payment-service/internal/repository"
	transporthttp "payment-service/internal/transport/http"
	"payment-service/internal/usecase"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// Config holds configuration for the Payment Service application.
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string
}

// LoadConfig reads configuration from environment variables with defaults.
func LoadConfig() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5441"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "payment_db"),
		ServerPort: getEnv("SERVER_PORT", "8081"),
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
	log.Println("Connected to Payment database")

	// 2. Create repository (implements domain.PaymentRepository)
	paymentRepo := repository.NewPostgresPaymentRepository(db)

	// 3. Create use case with injected dependencies
	paymentUseCase := usecase.NewPaymentUseCase(paymentRepo)

	// 4. Create HTTP handler
	handler := transporthttp.NewPaymentHandler(paymentUseCase)

	// 5. Set up Gin router
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
	log.Printf("Payment Service starting on %s", addr)
	return a.Router.Run(addr)
}

// Close cleans up resources.
func (a *App) Close() {
	if a.DB != nil {
		a.DB.Close()
	}
}
