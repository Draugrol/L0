package repository

import (
	"context"
	"database/sql"
	"fmt"

	"order-service/internal/models"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type OrderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func NewPostgresDB(host, port, user, password, dbname string) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}

func (r *OrderRepository) SaveOrder(ctx context.Context, order *models.Order) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert order
	orderQuery := `
		INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature,
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (order_uid) DO NOTHING
	`
	_, err = tx.ExecContext(ctx, orderQuery,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerID, order.DeliveryService,
		order.Shardkey, order.SmID, order.DateCreated, order.OofShard,
	)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	// Insert delivery
	deliveryQuery := `
		INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (order_uid) DO NOTHING
	`
	_, err = tx.ExecContext(ctx, deliveryQuery,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone,
		order.Delivery.Zip, order.Delivery.City, order.Delivery.Address,
		order.Delivery.Region, order.Delivery.Email,
	)
	if err != nil {
		return fmt.Errorf("failed to insert delivery: %w", err)
	}

	// Insert payment
	paymentQuery := `
		INSERT INTO payment (order_uid, transaction, request_id, currency, provider,
			amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (order_uid) DO NOTHING
	`
	_, err = tx.ExecContext(ctx, paymentQuery,
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID,
		order.Payment.Currency, order.Payment.Provider, order.Payment.Amount,
		order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee,
	)
	if err != nil {
		return fmt.Errorf("failed to insert payment: %w", err)
	}

	// Insert items
	itemQuery := `
		INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name,
			sale, size, total_price, nm_id, brand, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	for _, item := range order.Items {
		_, err = tx.ExecContext(ctx, itemQuery,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price,
			item.Rid, item.Name, item.Sale, item.Size, item.TotalPrice,
			item.NmID, item.Brand, item.Status,
		)
		if err != nil {
			return fmt.Errorf("failed to insert item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *OrderRepository) GetOrder(ctx context.Context, orderUID string) (*models.Order, error) {
	var order models.Order

	// Get order
	orderQuery := `SELECT * FROM orders WHERE order_uid = $1`
	err := r.db.GetContext(ctx, &order, orderQuery, orderUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Get delivery
	deliveryQuery := `SELECT * FROM delivery WHERE order_uid = $1`
	err = r.db.GetContext(ctx, &order.Delivery, deliveryQuery, orderUID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get delivery: %w", err)
	}

	// Get payment
	paymentQuery := `SELECT * FROM payment WHERE order_uid = $1`
	err = r.db.GetContext(ctx, &order.Payment, paymentQuery, orderUID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	// Get items
	itemsQuery := `SELECT * FROM items WHERE order_uid = $1`
	err = r.db.SelectContext(ctx, &order.Items, itemsQuery, orderUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}

	return &order, nil
}

func (r *OrderRepository) GetAllOrders(ctx context.Context) ([]models.Order, error) {
	var orderUIDs []string
	query := `SELECT order_uid FROM orders ORDER BY date_created DESC`
	err := r.db.SelectContext(ctx, &orderUIDs, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get order UIDs: %w", err)
	}

	orders := make([]models.Order, 0, len(orderUIDs))
	for _, uid := range orderUIDs {
		order, err := r.GetOrder(ctx, uid)
		if err != nil {
			return nil, err
		}
		if order != nil {
			orders = append(orders, *order)
		}
	}

	return orders, nil
}
