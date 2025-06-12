-- Создание таблицы заметок
CREATE TABLE notes (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    type VARCHAR(20) NOT NULL DEFAULT 'text',
    category VARCHAR(20) NOT NULL DEFAULT 'general',
    url TEXT,
    file_id VARCHAR(255),
    file_name VARCHAR(255),
    file_size BIGINT DEFAULT 0,
    tags TEXT,
    is_favorite BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    user_id BIGINT NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE
);

-- Создание индексов для оптимизации поиска
CREATE INDEX idx_notes_user_id ON notes(user_id);
CREATE INDEX idx_notes_type ON notes(type);
CREATE INDEX idx_notes_category ON notes(category);
CREATE INDEX idx_notes_is_favorite ON notes(is_favorite);
CREATE INDEX idx_notes_created_at ON notes(created_at);

-- Создание индекса для полнотекстового поиска
CREATE INDEX idx_notes_search ON notes USING gin(to_tsvector('russian', title || ' ' || coalesce(content, '') || ' ' || coalesce(tags, ''))); 