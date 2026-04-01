package http

import (
	"errors"
	"net/http"

	"order-service/internal/domain"
	"order-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

// OrderHandler handles HTTP requests for orders.
// Handlers are thin: they parse requests, call use cases, and return responses.
type OrderHandler struct {
	useCase *usecase.OrderUseCase
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler(useCase *usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{useCase: useCase}
}

// CreateOrderRequest represents the JSON payload for creating an order.
type CreateOrderRequest struct {
	CustomerID string `json:"customer_id" binding:"required"`
	ItemName   string `json:"item_name" binding:"required"`
	Amount     int64  `json:"amount" binding:"required"`
}

// OrderResponse represents the JSON response for an order.
type OrderResponse struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	ItemName   string `json:"item_name"`
	Amount     int64  `json:"amount"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}

// toOrderResponse converts a domain Order to an OrderResponse.
func toOrderResponse(order *domain.Order) OrderResponse {
	return OrderResponse{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		ItemName:   order.ItemName,
		Amount:     order.Amount,
		Status:     order.Status,
		CreatedAt:  order.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// CreateOrder handles POST /orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload: " + err.Error()})
		return
	}

	order, err := h.useCase.CreateOrder(c.Request.Context(), req.CustomerID, req.ItemName, req.Amount)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAmount) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, domain.ErrPaymentUnavailable) {
			// Return the order (marked as Failed) with 503 status
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "payment service unavailable",
				"order": toOrderResponse(order),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order"})
		return
	}

	status := http.StatusCreated
	if order.Status == domain.StatusFailed {
		status = http.StatusPaymentRequired
	}
	c.JSON(status, toOrderResponse(order))
}

// GetOrder handles GET /orders/:id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")

	order, err := h.useCase.GetOrder(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get order"})
		return
	}

	c.JSON(http.StatusOK, toOrderResponse(order))
}

// CancelOrder handles PATCH /orders/:id/cancel
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	id := c.Param("id")

	order, err := h.useCase.CancelOrder(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		if errors.Is(err, domain.ErrCannotCancelPaid) || errors.Is(err, domain.ErrCannotCancelOrder) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel order"})
		return
	}

	c.JSON(http.StatusOK, toOrderResponse(order))
}

// RegisterRoutes registers the order routes with the Gin engine.
func (h *OrderHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "Order Service",
			"version": "1.0.0",
			"endpoints": []gin.H{
				{"method": "POST", "path": "/orders", "description": "Create a new order"},
				{"method": "GET", "path": "/orders/:id", "description": "Get order by ID"},
				{"method": "PATCH", "path": "/orders/:id/cancel", "description": "Cancel a pending order"},
			},
		})
	})
	r.POST("/orders", h.CreateOrder)
	r.GET("/orders/:id", h.GetOrder)
	r.PATCH("/orders/:id/cancel", h.CancelOrder)
}
