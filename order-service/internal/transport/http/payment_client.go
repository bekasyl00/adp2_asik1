package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"order-service/internal/domain"
)

// PaymentHTTPClient implements domain.PaymentClient using REST.
// It calls the Payment Service over HTTP with proper timeouts.
type PaymentHTTPClient struct {
	client  *http.Client
	baseURL string
}

// NewPaymentHTTPClient creates a new PaymentHTTPClient.
// The http.Client should be configured with a timeout at the composition root.
func NewPaymentHTTPClient(client *http.Client, baseURL string) *PaymentHTTPClient {
	return &PaymentHTTPClient{
		client:  client,
		baseURL: baseURL,
	}
}

// paymentRequest represents the JSON payload sent to the Payment Service.
type paymentRequest struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

// paymentAPIResponse represents the JSON response from the Payment Service.
type paymentAPIResponse struct {
	ID            string `json:"id"`
	OrderID       string `json:"order_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
}

// AuthorizePayment calls the Payment Service to authorize a payment.
func (c *PaymentHTTPClient) AuthorizePayment(ctx context.Context, orderID string, amount int64) (*domain.PaymentResponse, error) {
	reqBody := paymentRequest{
		OrderID: orderID,
		Amount:  amount,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payment request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/payments", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create payment request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("payment service call failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read payment response: %w", err)
	}

	var paymentResp paymentAPIResponse
	if err := json.Unmarshal(body, &paymentResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payment response: %w", err)
	}

	return &domain.PaymentResponse{
		OrderID:       paymentResp.OrderID,
		TransactionID: paymentResp.TransactionID,
		Amount:        paymentResp.Amount,
		Status:        paymentResp.Status,
	}, nil
}
