package http

import (
	"database/sql"
	nethttp "net/http"
	"order-service/internal/usecase"

	grpcTransport "order-service/internal/transport/grpc"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	uc           *usecase.OrderUseCase
	streamServer *grpcTransport.OrderStreamServer
}

func NewOrderHandler(uc *usecase.OrderUseCase, streamServer *grpcTransport.OrderStreamServer) *OrderHandler {
	return &OrderHandler{
		uc:           uc,
		streamServer: streamServer,
	}
}

type createOrderRequest struct {
	CustomerID string `json:"customer_id"`
	ItemName   string `json:"item_name"`
	Amount     int64  `json:"amount"`
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req createOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(nethttp.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.uc.CreateOrder(req.CustomerID, req.ItemName, req.Amount)
	if err != nil {
		c.JSON(nethttp.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	h.streamServer.Notify(order.ID, order.Status)

	c.JSON(nethttp.StatusCreated, gin.H{
		"id":          order.ID,
		"customer_id": order.CustomerID,
		"item_name":   order.ItemName,
		"amount":      order.Amount,
		"status":      order.Status,
		"created_at":  order.CreatedAt,
	})
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")

	order, err := h.uc.GetOrderByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(nethttp.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(nethttp.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(nethttp.StatusOK, gin.H{
		"id":          order.ID,
		"customer_id": order.CustomerID,
		"item_name":   order.ItemName,
		"amount":      order.Amount,
		"status":      order.Status,
		"created_at":  order.CreatedAt,
	})
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	id := c.Param("id")

	err := h.uc.CancelOrder(id)
	if err != nil {
		c.JSON(nethttp.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.streamServer.Notify(id, "Cancelled")

	c.JSON(nethttp.StatusOK, gin.H{"message": "order cancelled"})
}
