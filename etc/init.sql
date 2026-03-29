CREATE TABLE IF NOT EXISTS users (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    username text NOT NULL UNIQUE CHECK (length(username) > 0),
    email text NOT NULL UNIQUE CHECK (position('@' in email) > 1),
    password_hash text NOT NULL CHECK (length(password_hash) > 0),

    first_name text NOT NULL CHECK (length(first_name) > 0),
    last_name text NOT NULL CHECK (length(last_name) > 0),

    is_premium boolean NOT NULL DEFAULT false,
    premium_expires_at timestamptz NULL,

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL
);

COMMENT ON TABLE users IS 'Таблица с информацией по пользователям';

CREATE TABLE IF NOT EXISTS tokens (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token text NOT NULL CHECK (length(token) > 0),
    expires_at timestamptz NOT NULL,

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL
);

COMMENT ON TABLE tokens IS 'Таблица с информацией по выданным токенам по результату авторизации';


CREATE TABLE IF NOT EXISTS roles (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    name text NOT NULL UNIQUE CHECK (length(name) > 0),

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL
);

COMMENT ON TABLE roles IS 'Таблица с информацией по ролям пользователей системы';

CREATE TABLE IF NOT EXISTS permissions (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    name text NOT NULL UNIQUE CHECK (length(name) > 0),

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL
);

COMMENT ON TABLE permissions IS 'Таблица с правами пользователей';
COMMENT ON COLUMN permissions.name IS 'Часть маршрута обращения к endpoint-у сервера, на которую выдается право доступа';

CREATE TABLE IF NOT EXISTS users_roles (

    user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id bigint NOT NULL REFERENCES roles(id)  ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

COMMENT ON TABLE users_roles IS 'Таблица СВЯЗЕЙ пользователей и ролей';

CREATE TABLE IF NOT EXISTS roles_permissions (

    role_id bigint NOT NULL REFERENCES roles(id)  ON DELETE CASCADE,
    permission_id bigint NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

COMMENT ON TABLE roles_permissions IS 'Таблица СВЯЗЕЙ ролей и прав пользователей';

CREATE TABLE IF NOT EXISTS product_categories (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    name text NOT NULL UNIQUE CHECK (length(name) > 0),

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL
);

COMMENT ON TABLE product_categories IS 'Таблица с информацией по категориям товаров';

CREATE TABLE IF NOT EXISTS products (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    name text NOT NULL CHECK (length(name) > 0),
    description text NOT NULL DEFAULT '',
    price numeric(12,2) NOT NULL CHECK (price >= 0),

    category_id bigint NOT NULL REFERENCES product_categories(id),

    is_active boolean NOT NULL DEFAULT true,

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL
);

COMMENT ON TABLE products IS 'Таблица с информацией по товарам';


CREATE TABLE IF NOT EXISTS product_discounts (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    product_id bigint NULL REFERENCES products(id) ON DELETE CASCADE,
    category_id bigint NULL REFERENCES product_categories(id) ON DELETE CASCADE,

    discount_percent numeric(5,2) NOT NULL CHECK (discount_percent >= 0 AND discount_percent <= 100),

    for_premium_only boolean NOT NULL DEFAULT false,

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL,

    CHECK (
        (product_id IS NOT NULL AND category_id IS NULL)
        OR
        (product_id IS NULL AND category_id IS NOT NULL)
    )
);

COMMENT ON TABLE product_discounts IS 'Таблица с информацией по скидкам на товары и категории товаров';


CREATE TABLE IF NOT EXISTS carts (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    user_id bigint NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL
);

COMMENT ON TABLE carts IS 'Таблица с информацией по корзине пользователя';

CREATE TABLE IF NOT EXISTS orders (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    status text NOT NULL DEFAULT 'new',
    total_price numeric(12,2) NOT NULL DEFAULT 0 CHECK (total_price >= 0),
    address text NOT NULL CHECK (length(address) > 0),

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL,

    CHECK (status IN ('new', 'paid', 'cancelled'))
);

COMMENT ON TABLE orders IS 'Таблица с информацией по заказам пользователя';

CREATE TABLE IF NOT EXISTS cart_items (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    cart_id bigint NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    product_id bigint NOT NULL REFERENCES products(id),
    quantity integer NOT NULL CHECK (quantity > 0),

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL,

    UNIQUE (cart_id, product_id)
);

COMMENT ON TABLE cart_items IS 'Товары в корзине пользователя';

CREATE TABLE IF NOT EXISTS order_items (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    order_id bigint NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id bigint NOT NULL REFERENCES products(id),
    quantity integer NOT NULL CHECK (quantity > 0),

    price numeric(12,2) NOT NULL CHECK (price >= 0),
    discount_percent numeric(5,2) NOT NULL DEFAULT 0 CHECK (discount_percent >= 0 AND discount_percent <= 100),

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NULL
);

COMMENT ON TABLE order_items IS 'Товары, входящие в заказ пользователя';

INSERT INTO roles (name)
VALUES
    ('admin'),
    ('user')
ON CONFLICT (name) DO NOTHING;


INSERT INTO permissions (name)
VALUES
    ('GET/products'),
    ('GET/product'),
    ('POST/products'),
    ('PUT/products'),
    ('DELETE/product'),

    ('GET/user'),

    ('PUT/cart'),
    ('GET/cart'),

    ('POST/order'),
    ('GET/order'),
    ('DELETE/order'),
    ('GET/orders'),

    ('POST/pay')
ON CONFLICT (name) DO NOTHING;

INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'user'
AND p.name IN (
    'GET/products',
    'GET/product',

    'GET/user',

    'PUT/cart',
    'GET/cart',

    'POST/order',
    'GET/order',
    'DELETE/order',
    'GET/orders',

    'POST/pay'
)
ON CONFLICT DO NOTHING;

INSERT INTO roles_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin'
AND p.name IN (
    'GET/products',
    'GET/product',
    'POST/products',
    'PUT/products',
    'DELETE/product',

    'GET/user',

    'PUT/cart',
    'GET/cart',

    'POST/order',
    'GET/order',
    'DELETE/order',
    'GET/orders',

    'POST/pay'
)
ON CONFLICT DO NOTHING;


