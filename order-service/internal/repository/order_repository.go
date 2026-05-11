package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"order-service/internal/domain"

	"github.com/redis/go-redis/v9"
)

type OrderRepository struct {
	db    *sql.DB
	redis *redis.Client
}

func NewOrderRepository(db *sql.DB, redisClient *redis.Client) *OrderRepository {
	return &OrderRepository{
		db:    db,
		redis: redisClient,
	}
}

func (r *OrderRepository) Create(order *domain.Order) error {
	query := `
		INSERT INTO orders (id, customer_id, item_name, amount, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(
		query,
		order.ID,
		order.CustomerID,
		order.ItemName,
		order.Amount,
		order.Status,
		order.CreatedAt,
	)

	return err
}

func (r *OrderRepository) GetByID(id string) (*domain.Order, error) {
	ctx := context.Background()

	cacheKey := fmt.Sprintf("order:%s", id)

	// 1. Check Redis first
	cachedOrder, err := r.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var order domain.Order

		err := json.Unmarshal([]byte(cachedOrder), &order)
		if err == nil {
			fmt.Println("Order loaded from Redis cache")
			return &order, nil
		}
	}

	// 2. Fallback to DB
	query := `
		SELECT id, customer_id, item_name, amount, status, created_at
		FROM orders
		WHERE id = $1
	`

	var order domain.Order

	err = r.db.QueryRow(query, id).Scan(
		&order.ID,
		&order.CustomerID,
		&order.ItemName,
		&order.Amount,
		&order.Status,
		&order.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// 3. Save to Redis
	orderJSON, err := json.Marshal(order)
	if err == nil {
		ttlStr := os.Getenv("CACHE_TTL_SECONDS")

		ttlSeconds := 300

		if ttlStr != "" {
			if parsed, parseErr := strconv.Atoi(ttlStr); parseErr == nil {
				ttlSeconds = parsed
			}
		}

		r.redis.Set(
			ctx,
			cacheKey,
			orderJSON,
			time.Duration(ttlSeconds)*time.Second,
		)
	}

	fmt.Println("Order loaded from PostgreSQL")

	return &order, nil
}

func (r *OrderRepository) UpdateStatus(id string, status string) error {
	query := `
		UPDATE orders
		SET status = $1
		WHERE id = $2
	`

	_, err := r.db.Exec(query, status, id)
	if err != nil {
		return err
	}

	// Cache invalidation
	ctx := context.Background()
	cacheKey := fmt.Sprintf("order:%s", id)

	r.redis.Del(ctx, cacheKey)

	fmt.Println("Redis cache invalidated for order:", id)

	return nil
}
