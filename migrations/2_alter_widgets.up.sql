-- скрипты создания сущностей
ALTER TABLE IF EXISTS widgets
 ADD COLUMN quantity integer DEFAULT 0 NOT NULL;

-- в случае исполнения скриптов вручную мы должны прописать версию сами
INSERT INTO go_migrations (version) VALUES
    (2);
