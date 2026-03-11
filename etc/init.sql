CREATE TABLE IF NOT EXISTS users (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE users IS 'Таблица с информацией по пользователям';

CREATE TABLE IF NOT EXISTS tokens (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id bigint REFERENCES users(id) NOT NULL,
    token bytea NOT NULL,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE tokens IS 'Таблица с информацией по выданным токенам по результату авторизации';


CREATE TABLE IF NOT EXISTS roles (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE roles IS 'Таблица с информацией по ролям пользователей системы';

CREATE TABLE IF NOT EXISTS permissions (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE permissions IS 'Таблица с правами пользователей';
COMMENT ON COLUMN permissions.name IS 'Часть маршрута обращения к endpoint-у сервера, на которую выдается право доступа';

CREATE TABLE IF NOT EXISTS users_roles (
    user_id bigint REFERENCES users(id) NOT NULL,
    role_id bigint REFERENCES roles(id) NOT NULL,
    PRIMARY KEY (user_id, role_id)
);

COMMENT ON TABLE users_roles IS 'Таблица СВЯЗЕЙ пользователей и ролей';

CREATE TABLE IF NOT EXISTS roles_permissions (
    role_id bigint REFERENCES roles(id) NOT NULL,
    permission_id bigint REFERENCES permissions(id) NOT NULL,
    PRIMARY KEY (role_id, permission_id)
);

COMMENT ON TABLE roles_permissions IS 'Таблица СВЯЗЕЙ ролей и прав пользователей';

CREATE TABLE IF NOT EXISTS products (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE products IS 'Таблица с информацией по товарам';

CREATE TABLE IF NOT EXISTS product_categories (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE product_categories IS 'Таблица с информацией по категориям товаров';

CREATE TABLE IF NOT EXISTS product_discounts (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE users IS 'Таблица с информацией по скидкам на товары/категории товаров';

CREATE TABLE IF NOT EXISTS carts (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE users IS 'Таблица с информацией по корзине пользователя';

CREATE TABLE IF NOT EXISTS orders (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    address text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE orders IS 'Таблица с информацией по заказам пользователя';


