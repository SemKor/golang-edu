package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
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

func (r *Repository) GetProducts(ctx context.Context, filter model.ProductFilter) ([]model.Product, int, error) {
	baseQuery := `
		FROM products
		WHERE deleted_at IS NULL AND is_active = TRUE
	`

	args := []interface{}{}
	argIndex := 1

	if filter.Name != "" {
		baseQuery += fmt.Sprintf(" AND LOWER(name) LIKE LOWER($%d)", argIndex)
		args = append(args, "%"+filter.Name+"%")
		argIndex++
	}

	if len(filter.Category) > 0 {
		baseQuery += fmt.Sprintf(" AND category_id = ANY($%d)", argIndex)
		args = append(args, pq.Array(filter.Category))
		argIndex++
	}

	if filter.MinPrice > 0 {
		baseQuery += fmt.Sprintf(" AND price >= $%d", argIndex)
		args = append(args, filter.MinPrice)
		argIndex++
	}

	if filter.MaxPrice > 0 {
		baseQuery += fmt.Sprintf(" AND price <= $%d", argIndex)
		args = append(args, filter.MaxPrice)
		argIndex++
	}

	countQuery := "SELECT COUNT(*) " + baseQuery

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, name, description, price, category_id, is_active, created_at, updated_at, deleted_at
	` + baseQuery + `
		ORDER BY id
	`

	if filter.Count > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Count)
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		var deletedAt sql.NullTime

		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.Price,
			&p.CategoryID,
			&p.IsActive,
			&p.CreatedAt,
			&p.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if deletedAt.Valid {
			p.DeletedAt = &deletedAt.Time
		}

		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return products, total, nil
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
		if err == sql.ErrNoRows {
			return model.Cart{
				UserID: userID,
				Items:  []model.CartItem{},
			}, nil
		}
		return model.Cart{}, err
	}

	cart.UserID = userID

	var isPremium bool
	err = r.db.QueryRowContext(ctx,
		`SELECT is_premium FROM users WHERE id = $1 AND deleted_at IS NULL`,
		userID,
	).Scan(&isPremium)
	if err != nil {
		return model.Cart{}, err
	}

	itemsQuery := `
		SELECT
			ci.id,
			ci.product_id,
			ci.quantity,
			p.id,
			p.name,
			p.description,
			p.price,
			p.category_id,
			p.is_active,
			p.created_at,
			p.updated_at,
			p.deleted_at
		FROM cart_items ci
		INNER JOIN products p ON p.id = ci.product_id
		WHERE ci.cart_id = $1
		  AND ci.deleted_at IS NULL
		  AND p.deleted_at IS NULL
		  AND p.is_active = TRUE
		ORDER BY ci.id
	`

	rows, err := r.db.QueryContext(ctx, itemsQuery, cart.ID)
	if err != nil {
		return model.Cart{}, err
	}
	defer rows.Close()

	var items []model.CartItem
	for rows.Next() {
		var item model.CartItem
		var product model.Product
		var productDeletedAt sql.NullTime

		err := rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.Quantity,
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.CategoryID,
			&product.IsActive,
			&product.CreatedAt,
			&product.UpdatedAt,
			&productDeletedAt,
		)
		if err != nil {
			return model.Cart{}, err
		}

		if productDeletedAt.Valid {
			product.DeletedAt = &productDeletedAt.Time
		}

		item.CartID = cart.ID

		discount, err := r.GetProductDiscount(ctx, item.ProductID, isPremium)
		if err != nil {
			return model.Cart{}, err
		}

		item.DiscountPercent = discount
		product.Price = product.Price * (1 - discount/100)
		item.Product = &product

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return model.Cart{}, err
	}

	cart.Items = items
	return cart, nil
}

func (r *Repository) CreateOrder(ctx context.Context, userID int64, address string) (model.Order, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Order{}, err
	}
	defer tx.Rollback()

	var cartID int64
	err = tx.QueryRowContext(ctx,
		`SELECT id FROM carts WHERE user_id = $1 AND deleted_at IS NULL`,
		userID,
	).Scan(&cartID)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Order{}, fmt.Errorf("cart not found")
		}
		return model.Order{}, err
	}

	var isPremium bool
	err = tx.QueryRowContext(ctx,
		`SELECT is_premium FROM users WHERE id = $1 AND deleted_at IS NULL`,
		userID,
	).Scan(&isPremium)
	if err != nil {
		return model.Order{}, err
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT ci.product_id, ci.quantity, p.price
		FROM cart_items ci
		INNER JOIN products p ON p.id = ci.product_id
		WHERE ci.cart_id = $1
		  AND ci.deleted_at IS NULL
		  AND p.deleted_at IS NULL
		  AND p.is_active = TRUE
	`, cartID)
	if err != nil {
		return model.Order{}, err
	}
	defer rows.Close()

	type orderProduct struct {
		ProductID int64
		Quantity  int
		Price     float64
		Discount  float64
	}

	var products []orderProduct
	var totalPrice float64

	for rows.Next() {
		var item orderProduct
		err := rows.Scan(&item.ProductID, &item.Quantity, &item.Price)
		if err != nil {
			return model.Order{}, err
		}

		discount, err := r.GetProductDiscount(ctx, item.ProductID, isPremium)
		if err != nil {
			return model.Order{}, err
		}

		item.Discount = discount
		products = append(products, item)

		finalPrice := item.Price * (1 - discount/100)
		totalPrice += finalPrice * float64(item.Quantity)
	}

	if err := rows.Err(); err != nil {
		return model.Order{}, err
	}

	if len(products) == 0 {
		return model.Order{}, fmt.Errorf("cart is empty")
	}

	var order model.Order
	err = tx.QueryRowContext(ctx,
		`INSERT INTO orders (user_id, status, total_price, address, created_at, updated_at)
		 VALUES ($1, 'new', $2, $3, NOW(), NOW())
		 RETURNING id, user_id, status, total_price, address, created_at, updated_at`,
		userID, totalPrice, address,
	).Scan(
		&order.ID,
		&order.UserID,
		&order.Status,
		&order.TotalPrice,
		&order.Address,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		return model.Order{}, err
	}

	for _, item := range products {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO order_items (
				order_id,
				product_id,
				quantity,
				price,
				discount_percent,
				created_at,
				updated_at
			)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`,
			order.ID,
			item.ProductID,
			item.Quantity,
			item.Price,
			item.Discount,
		)
		if err != nil {
			return model.Order{}, err
		}
	}

	_, err = tx.ExecContext(ctx,
		`DELETE FROM cart_items WHERE cart_id = $1`,
		cartID,
	)
	if err != nil {
		return model.Order{}, err
	}

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

func (r *Repository) AssignRoleToUser(ctx context.Context, userID int64, roleName string) error {
	var roleID int64

	err := r.db.QueryRowContext(ctx,
		`SELECT id FROM roles WHERE name = $1`,
		roleName,
	).Scan(&roleID)
	if err != nil {
		return fmt.Errorf("cannot find role %s: %w", roleName, err)
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO users_roles (user_id, role_id)
		 VALUES ($1, $2)
		 ON CONFLICT (user_id, role_id) DO NOTHING`,
		userID, roleID,
	)
	if err != nil {
		return fmt.Errorf("cannot assign role to user: %w", err)
	}

	return nil
}

func (r *Repository) GetProductDiscount(ctx context.Context, productID int64, isPremium bool) (float64, error) {
	query := `
		SELECT COALESCE(MAX(pd.discount_percent), 0)
		FROM products p
		LEFT JOIN product_discounts pd
			ON (
				(pd.product_id = p.id AND pd.category_id IS NULL)
				OR
				(pd.category_id = p.category_id AND pd.product_id IS NULL)
			)
		WHERE p.id = $1
		  AND p.deleted_at IS NULL
		  AND (
			  pd.id IS NULL
			  OR pd.deleted_at IS NULL
		  )
		  AND (
			  pd.id IS NULL
			  OR pd.for_premium_only = FALSE
			  OR $2 = TRUE
		  )
	`

	var discount float64
	err := r.db.QueryRowContext(ctx, query, productID, isPremium).Scan(&discount)
	if err != nil {
		return 0, err
	}

	return discount, nil
}

func (r *Repository) CancelOrder(ctx context.Context, orderID int64, userID int64) error {
	var status string

	err := r.db.QueryRowContext(ctx,
		`SELECT status
		 FROM orders
		 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		orderID, userID,
	).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("order not found")
		}
		return err
	}

	if status == "paid" {
		return fmt.Errorf("paid order cannot be cancelled")
	}

	if status == "cancelled" {
		return nil
	}

	res, err := r.db.ExecContext(ctx,
		`UPDATE orders
		 SET status = 'cancelled',
		     updated_at = NOW()
		 WHERE id = $1
		   AND user_id = $2
		   AND deleted_at IS NULL`,
		orderID, userID,
	)
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

func (r *Repository) GetOrderByIDForUser(ctx context.Context, orderID int64, userID int64) (model.Order, error) {
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
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	var order model.Order
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, orderID, userID).Scan(
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

func (r *Repository) ActivatePremium(ctx context.Context, userID int64, expiresAt time.Time) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE users
		 SET is_premium = TRUE,
		     premium_expires_at = $2,
		     updated_at = NOW()
		 WHERE id = $1
		   AND deleted_at IS NULL`,
		userID, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("cannot activate premium: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *Repository) CreateProducts(ctx context.Context, products []model.ProductCreateInput) ([]model.Product, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	created := make([]model.Product, 0, len(products))

	for _, input := range products {
		var product model.Product
		var deletedAt sql.NullTime

		err := tx.QueryRowContext(ctx,
			`INSERT INTO products (
				name,
				description,
				price,
				category_id,
				is_active,
				created_at,
				updated_at,
				deleted_at
			)
			VALUES ($1, $2, $3, $4, TRUE, NOW(), NOW(), NULL)
			RETURNING id, name, description, price, category_id, is_active, created_at, updated_at, deleted_at`,
			input.Name,
			input.Description,
			input.Price,
			input.CategoryID,
		).Scan(
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

		if input.PremiumDiscount > 0 {
			_, err = tx.ExecContext(ctx,
				`INSERT INTO product_discounts (
					product_id,
					category_id,
					discount_percent,
					for_premium_only,
					created_at,
					updated_at,
					deleted_at
				)
				VALUES ($1, NULL, $2, TRUE, NOW(), NOW(), NULL)`,
				product.ID,
				input.PremiumDiscount,
			)
			if err != nil {
				return nil, err
			}
		}

		if input.CategoryDiscount > 0 {
			_, err = tx.ExecContext(ctx,
				`INSERT INTO product_discounts (
					product_id,
					category_id,
					discount_percent,
					for_premium_only,
					created_at,
					updated_at,
					deleted_at
				)
				VALUES (NULL, $1, $2, FALSE, NOW(), NOW(), NULL)`,
				product.CategoryID,
				input.CategoryDiscount,
			)
			if err != nil {
				return nil, err
			}
		}

		created = append(created, product)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return created, nil
}

func (r *Repository) UpdateProducts(ctx context.Context, products []model.ProductUpdateInput) ([]model.Product, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	updated := make([]model.Product, 0, len(products))

	for _, input := range products {
		var product model.Product
		var deletedAt sql.NullTime

		err := tx.QueryRowContext(ctx,
			`UPDATE products
			 SET name = $2,
			     description = $3,
			     price = $4,
			     category_id = $5,
			     updated_at = NOW()
			 WHERE id = $1
			   AND deleted_at IS NULL
			 RETURNING id, name, description, price, category_id, is_active, created_at, updated_at, deleted_at`,
			input.ID,
			input.Name,
			input.Description,
			input.Price,
			input.CategoryID,
		).Scan(
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

		// Удаляем только скидки конкретного товара (НЕ категории!)
		_, err = tx.ExecContext(ctx,
			`UPDATE product_discounts
			 SET deleted_at = NOW(),
			     updated_at = NOW()
			 WHERE product_id = $1`,
			product.ID,
		)
		if err != nil {
			return nil, err
		}

		// Добавляем premium скидку (если есть)
		if input.PremiumDiscount > 0 {
			_, err = tx.ExecContext(ctx,
				`INSERT INTO product_discounts (
					product_id,
					category_id,
					discount_percent,
					for_premium_only,
					created_at,
					updated_at,
					deleted_at
				)
				VALUES ($1, NULL, $2, TRUE, NOW(), NOW(), NULL)`,
				product.ID,
				input.PremiumDiscount,
			)
			if err != nil {
				return nil, err
			}
		}

		updated = append(updated, product)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return updated, nil
}

func (r *Repository) DeleteProduct(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE products
		 SET deleted_at = NOW(),
		     updated_at = NOW()
		 WHERE id = $1 AND deleted_at IS NULL`,
		id,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

func (r *Repository) ReplaceCart(ctx context.Context, userID int64, items []model.CartUpdateItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var cartID int64
	err = tx.QueryRowContext(ctx,
		`SELECT id FROM carts WHERE user_id = $1 AND deleted_at IS NULL`,
		userID,
	).Scan(&cartID)
	if err != nil {
		if err == sql.ErrNoRows {
			err = tx.QueryRowContext(ctx,
				`INSERT INTO carts (user_id, created_at, updated_at, deleted_at)
				 VALUES ($1, NOW(), NOW(), NULL)
				 RETURNING id`,
				userID,
			).Scan(&cartID)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	_, err = tx.ExecContext(ctx,
		`DELETE FROM cart_items WHERE cart_id = $1`,
		cartID,
	)
	if err != nil {
		return err
	}

	for _, item := range items {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO cart_items (
				cart_id,
				product_id,
				quantity,
				created_at,
				updated_at,
				deleted_at
			)
			VALUES ($1, $2, $3, NOW(), NOW(), NULL)`,
			cartID,
			item.ProductID,
			item.Quantity,
		)
		if err != nil {
			return err
		}
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE carts
		 SET updated_at = NOW()
		 WHERE id = $1`,
		cartID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
