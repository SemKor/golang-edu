-- скрипты отката к предыдущей версии
ALTER TABLE IF EXISTS widgets
    DROP COLUMN IF EXISTS quantity;

-- в случае исполнения скриптов вручную мы должны прописать версию сами
INSERT INTO go_migrations (version) VALUES
    (1);