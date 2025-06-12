-- Удаление индексов
DROP INDEX IF EXISTS idx_tasks_priority;
DROP INDEX IF EXISTS idx_tasks_created_at;
DROP INDEX IF EXISTS idx_tasks_notify_at;
DROP INDEX IF EXISTS idx_tasks_status;
DROP INDEX IF EXISTS idx_tasks_user_id;
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_telegram_id;
DROP INDEX IF EXISTS idx_users_telegram_id;

-- Удаление таблиц в правильном порядке (с учетом внешних ключей)
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users; 