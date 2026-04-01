package http

import (
	"errors"
	"net/http"

	"payment-service/internal/domain"
	"payment-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

// PaymentHandler handles HTTP requests for payments.
// Handlers are thin: they parse requests, call use cases, and return responses.
type PaymentHandler struct {
	useCase *usecase.PaymentUseCase
}

// NewPaymentHandler creates a new PaymentHandler.
func NewPaymentHandler(useCase *usecase.PaymentUseCase) *PaymentHandler {
	return &PaymentHandler{useCase: useCase}
}

// CreatePaymentRequest represents the JSON payload for authorizing a payment.
type CreatePaymentRequest struct {
	OrderID string `json:"order_id" binding:"required"`
	Amount  int64  `json:"amount" binding:"required"`
}

// PaymentResponse represents the JSON response for a payment.
type PaymentResponse struct {
	ID            string `json:"id"`
	OrderID       string `json:"order_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
}

// toPaymentResponse converts a domain Payment to a PaymentResponse.
func toPaymentResponse(payment *domain.Payment) PaymentResponse {
	return PaymentResponse{
		ID:            payment.ID,
		OrderID:       payment.OrderID,
		TransactionID: payment.TransactionID,
		Amount:        payment.Amount,
		Status:        payment.Status,
	}
}

// CreatePayment handles POST /payments
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload: " + err.Error()})
		return
	}

	payment, err := h.useCase.AuthorizePayment(c.Request.Context(), req.OrderID, req.Amount)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAmount) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process payment"})
		return
	}

	c.JSON(http.StatusCreated, toPaymentResponse(payment))
}

// GetPayment handles GET /payments/:order_id
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	orderID := c.Param("order_id")

	payment, err := h.useCase.GetPaymentByOrderID(c.Request.Context(), orderID)
	if err != nil {
		if errors.Is(err, domain.ErrPaymentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get payment"})
		return
	}

	c.JSON(http.StatusOK, toPaymentResponse(payment))
}

// RegisterRoutes registers the payment routes with the Gin engine.
func (h *PaymentHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "Payment Service",
			"version": "1.0.0",
			"endpoints": []gin.H{
				{"method": "POST", "path": "/payments", "description": "Authorize a payment"},
				{"method": "GET", "path": "/payments/:order_id", "description": "Get payment by order ID"},
			},
		})
	})
	r.POST("/payments", h.CreatePayment)
	r.GET("/payments/:order_id", h.GetPayment)
}
