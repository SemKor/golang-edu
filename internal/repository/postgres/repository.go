package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"task5/internal/domain/model"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(ctx context.Context, user model.User) (model.User, error) {
	query := `
		INSERT INTO users (
			username,
			email,
			password_hash,
			first_name,
			last_name,
			is_premium,
			premium_expires_at,
			created_at,
			updated_at,
			deleted_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW(), NULL)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.IsPremium,
		user.PremiumExpiresAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (model.User, error) {
	query := `
		SELECT
			id,
			username,
			email,
			password_hash,
			first_name,
			last_name,
			is_premium,
			premium_expires_at,
			created_at,
			updated_at,
			deleted_at
		FROM users
		WHERE username = $1 AND deleted_at IS NULL
	`

	var user model.User
	var premiumExpiresAt sql.NullTime
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.IsPremium,
		&premiumExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		return model.User{}, err
	}

	if premiumExpiresAt.Valid {
		user.PremiumExpiresAt = &premiumExpiresAt.Time
	}
	if deletedAt.Valid {
		user.DeletedAt = &deletedAt.Time
	}

	return user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (model.User, error) {
	query := `
		SELECT
			id,
			username,
			email,
			password_hash,
			first_name,
			last_name,
			is_premium,
			premium_expires_at,
			created_at,
			updated_at,
			deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var user model.User
	var premiumExpiresAt sql.NullTime
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.IsPremium,
		&premiumExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		return model.User{}, err
	}

	if premiumExpiresAt.Valid {
		user.PremiumExpiresAt = &premiumExpiresAt.Time
	}
	if deletedAt.Valid {
		user.DeletedAt = &deletedAt.Time
	}

	return user, nil
}

func (r *Repository) SaveToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO tokens (
			user_id,
			token,
			expires_at,
			created_at,
			updated_at,
			deleted_at
		)
		VALUES ($1, $2, $3, NOW(), NOW(), NULL)
	`

	_, err := r.db.ExecContext(ctx, query, userID, token, expiresAt)
	if err != nil {
		return fmt.Errorf("cannot save token: %w", err)
	}

	return nil
}

func (r *Repository) GetUserRoles(ctx context.Context, userID int64) ([]string, error) {
	query := `
		SELECT r.name
		FROM roles r
		INNER JOIN users_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *Repository) GetUserPermissions(ctx context.Context, userID int64) ([]string, error) {
	query := `
		SELECT DISTINCT p.name
		FROM permissions p
		INNER JOIN roles_permissions rp ON rp.permission_id = p.id
		INNER JOIN users_roles ur ON ur.role_id = rp.role_id
		WHERE ur.user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var permission string
		if err := rows.Scan(&permission); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

func (r *Repository) GetProducts(ctx context.Context) ([]model.Product, error) {
	query := `
		SELECT
			id,
			name,
			description,
			price,
			category_id,
			is_active,
			created_at,
			updated_at,
			deleted_at
		FROM products
		WHERE deleted_at IS NULL AND is_active = TRUE
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var product model.Product
		var deletedAt sql.NullTime

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.CategoryID,
			&product.IsActive,
			&product.CreatedAt,
			&product.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, err
		}

		if deletedAt.Valid {
			product.DeletedAt = &deletedAt.Time
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func (r *Repository) GetProductByID(ctx context.Context, id int64) (model.Product, error) {
	query := `
		SELECT
			id,
			name,
			description,
			price,
			category_id,
			is_active,
			created_at,
			updated_at,
			deleted_at
		FROM products
		WHERE id = $1 AND deleted_at IS NULL AND is_active = TRUE
	`

	var product model.Product
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.CategoryID,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		return model.Product{}, err
	}

	if deletedAt.Valid {
		product.DeletedAt = &deletedAt.Time
	}

	return product, nil
}

func (r *Repository) GetCart(ctx context.Context, userID int64) (model.Cart, error) {
	query := `
		SELECT id
		FROM carts
		WHERE user_id = $1 AND deleted_at IS NULL
	`

	var cart model.Cart
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&cart.ID)
	if err != nil {
		return model.Cart{}, err
	}

	cart.UserID = userID

	// получаем товары
	itemsQuery := `
		SELECT
			ci.id,
			ci.product_id,
			ci.quantity
		FROM cart_items ci
		WHERE ci.cart_id = $1 AND ci.deleted_at IS NULL
	`

	rows, err := r.db.QueryContext(ctx, itemsQuery, cart.ID)
	if err != nil {
		return model.Cart{}, err
	}
	defer rows.Close()

	var items []model.CartItem
	for rows.Next() {
		var item model.CartItem

		err := rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.Quantity,
		)
		if err != nil {
			return model.Cart{}, err
		}

		item.CartID = cart.ID
		items = append(items, item)
	}

	cart.Items = items
	return cart, nil
}

func (r *Repository) AddToCart(ctx context.Context, userID, productID int64, quantity int) error {
	var cartID int64

	err := r.db.QueryRowContext(ctx,
		`SELECT id FROM carts WHERE user_id = $1 AND deleted_at IS NULL`,
		userID,
	).Scan(&cartID)

	if err != nil {
		// создаем корзину
		err = r.db.QueryRowContext(ctx,
			`INSERT INTO carts (user_id, created_at, updated_at)
			 VALUES ($1, NOW(), NOW())
			 RETURNING id`,
			userID,
		).Scan(&cartID)
		if err != nil {
			return err
		}
	}

	// вставка или обновление
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO cart_items (cart_id, product_id, quantity, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (cart_id, product_id)
		DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity,
			updated_at = NOW()
	`, cartID, productID, quantity)

	return err
}

func (r *Repository) CreateOrder(ctx context.Context, userID int64, address string) (model.Order, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Order{}, err
	}
	defer tx.Rollback()

	var cartID int64
	err = tx.QueryRowContext(ctx,
		`SELECT id FROM carts WHERE user_id = $1`,
		userID,
	).Scan(&cartID)
	if err != nil {
		return model.Order{}, fmt.Errorf("cart not found")
	}

	// создаем заказ
	var order model.Order
	err = tx.QueryRowContext(ctx,
		`INSERT INTO orders (user_id, address, created_at, updated_at)
		 VALUES ($1, $2, NOW(), NOW())
		 RETURNING id`,
		userID, address,
	).Scan(&order.ID)
	if err != nil {
		return model.Order{}, err
	}

	// переносим товары из корзины
	rows, err := tx.QueryContext(ctx,
		`SELECT product_id, quantity FROM cart_items WHERE cart_id = $1`,
		cartID,
	)
	if err != nil {
		return model.Order{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var productID int64
		var quantity int

		err := rows.Scan(&productID, &quantity)
		if err != nil {
			return model.Order{}, err
		}

		_, err = tx.ExecContext(ctx,
			`INSERT INTO order_items (order_id, product_id, quantity, price, created_at, updated_at)
			 VALUES ($1, $2, $3, 0, NOW(), NOW())`,
			order.ID, productID, quantity,
		)
		if err != nil {
			return model.Order{}, err
		}
	}

	// очищаем корзину
	_, _ = tx.ExecContext(ctx,
		`DELETE FROM cart_items WHERE cart_id = $1`,
		cartID,
	)

	err = tx.Commit()
	if err != nil {
		return model.Order{}, err
	}

	return order, nil
}

func (r *Repository) GetOrderByID(ctx context.Context, id int64) (model.Order, error) {
	query := `
		SELECT
			id,
			user_id,
			status,
			total_price,
			address,
			created_at,
			updated_at,
			deleted_at
		FROM orders
		WHERE id = $1 AND deleted_at IS NULL
	`

	var order model.Order
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.TotalPrice,
		&order.Address,
		&order.CreatedAt,
		&order.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		return model.Order{}, err
	}

	if deletedAt.Valid {
		order.DeletedAt = &deletedAt.Time
	}

	itemsQuery := `
		SELECT
			id,
			order_id,
			product_id,
			quantity,
			price,
			discount_percent,
			created_at,
			updated_at,
			deleted_at
		FROM order_items
		WHERE order_id = $1 AND deleted_at IS NULL
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, itemsQuery, order.ID)
	if err != nil {
		return model.Order{}, err
	}
	defer rows.Close()

	var items []model.OrderItem
	for rows.Next() {
		var item model.OrderItem
		var itemDeletedAt sql.NullTime

		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.Price,
			&item.DiscountPercent,
			&item.CreatedAt,
			&item.UpdatedAt,
			&itemDeletedAt,
		)
		if err != nil {
			return model.Order{}, err
		}

		if itemDeletedAt.Valid {
			item.DeletedAt = &itemDeletedAt.Time
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return model.Order{}, err
	}

	order.Items = items
	return order, nil
}

func (r *Repository) GetOrdersByUser(ctx context.Context, userID int64) ([]model.Order, error) {
	query := `
		SELECT
			id,
			user_id,
			status,
			total_price,
			address,
			created_at,
			updated_at,
			deleted_at
		FROM orders
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY id DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var order model.Order
		var deletedAt sql.NullTime

		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Status,
			&order.TotalPrice,
			&order.Address,
			&order.CreatedAt,
			&order.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, err
		}

		if deletedAt.Valid {
			order.DeletedAt = &deletedAt.Time
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *Repository) PayOrder(ctx context.Context, orderID int64) error {
	query := `
		UPDATE orders
		SET status = 'paid',
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	res, err := r.db.ExecContext(ctx, query, orderID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}
