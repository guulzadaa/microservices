package repository

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type PaymentHTTPClient struct {
	baseURL string
	client  *http.Client
}

func NewPaymentHTTPClient(baseURL string, client *http.Client) *PaymentHTTPClient {
	return &PaymentHTTPClient{
		baseURL: baseURL,
		client:  client,
	}
}

type createPaymentRequest struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

type createPaymentResponse struct {
	ID            string `json:"id"`
	OrderID       string `json:"order_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
}

func (p *PaymentHTTPClient) CreatePayment(orderID string, amount int64) (string, string, error) {
	reqBody := createPaymentRequest{
		OrderID: orderID,
		Amount:  amount,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", err
	}

	resp, err := p.client.Post(
		p.baseURL+"/payments",
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", "", errors.New("payment service returned unexpected status")
	}

	var result createPaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}

	return result.TransactionID, result.Status, nil
}
