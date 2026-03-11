-- создадим техническую таблицу, в которой будем хранить данные о текущей версии
CREATE TABLE IF NOT EXISTS go_migrations (
    id integer PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    version integer NOT NULL,
    created_at timestamp DEFAULT now() NOT NULL
);

-- скрипты создания сущностей
CREATE TABLE IF NOT EXISTS widgets (
    id integer PRIMARY KEY,
    name text NOT NULL,
    weight numeric NOT NULL,
    -- технические колонки для отслеживания состояния сущности в "стандартном" виде, с которым обычно умеют работать около-ORM фреймворки
    created_at timestamp DEFAULT now() NOT NULL,
    updated_at timestamp DEFAULT now() NOT NULL,
    deleted_at timestamp
);

INSERT INTO widgets (id, name, weight) VALUES
    (1, 'cup', 0.1) ON CONFLICT (id) DO UPDATE SET name = excluded.name, weight = excluded.weight;

-- в случае исполнения скриптов вручную мы должны прописать версию сами
INSERT INTO go_migrations (version) VALUES
    (1);