package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"todolist/internal/domain"
)

// NoteRepositoryImpl реализует интерфейс NoteRepository без Squirrel
type NoteRepositoryImpl struct {
	db *Database
}

// NewNoteRepository создает новый экземпляр NoteRepositoryImpl
func NewNoteRepository(db *Database) domain.NoteRepository {
	return &NoteRepositoryImpl{
		db: db,
	}
}

// Create создает новую заметку
func (r *NoteRepositoryImpl) Create(ctx context.Context, note *domain.Note) error {
	query := `
		INSERT INTO notes (title, content, type, category, url, file_id, file_name, file_size, tags, is_favorite, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at`

	err := r.db.DB.QueryRowContext(ctx, query,
		note.Title, note.Content, note.Type, note.Category, note.URL,
		note.FileID, note.FileName, note.FileSize, note.Tags,
		note.IsFavorite, note.UserID).Scan(
		&note.ID, &note.CreatedAt, &note.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}

	return nil
}

// GetByID получает заметку по ID
func (r *NoteRepositoryImpl) GetByID(ctx context.Context, id int) (*domain.Note, error) {
	query := `
		SELECT id, title, content, type, category, url, file_id, file_name, file_size, 
		       tags, is_favorite, created_at, updated_at, user_id
		FROM notes WHERE id = $1`

	note := &domain.Note{}
	err := r.db.DB.QueryRowContext(ctx, query, id).Scan(
		&note.ID, &note.Title, &note.Content, &note.Type, &note.Category,
		&note.URL, &note.FileID, &note.FileName, &note.FileSize, &note.Tags,
		&note.IsFavorite, &note.CreatedAt, &note.UpdatedAt, &note.UserID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("note not found")
		}
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	return note, nil
}

// GetByUserID получает все заметки пользователя
func (r *NoteRepositoryImpl) GetByUserID(ctx context.Context, userID int64) ([]*domain.Note, error) {
	query := `
		SELECT id, title, content, type, category, url, file_id, file_name, file_size,
		       tags, is_favorite, created_at, updated_at, user_id
		FROM notes WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notes: %w", err)
	}
	defer rows.Close()

	return r.scanNotes(rows)
}

// GetByCategory получает заметки пользователя по категории
func (r *NoteRepositoryImpl) GetByCategory(ctx context.Context, userID int64, category domain.NoteCategory) ([]*domain.Note, error) {
	query := `
		SELECT id, title, content, type, category, url, file_id, file_name, file_size,
		       tags, is_favorite, created_at, updated_at, user_id
		FROM notes WHERE user_id = $1 AND category = $2 ORDER BY created_at DESC`

	rows, err := r.db.DB.QueryContext(ctx, query, userID, category)
	if err != nil {
		return nil, fmt.Errorf("failed to get notes by category: %w", err)
	}
	defer rows.Close()

	return r.scanNotes(rows)
}

// GetByType получает заметки пользователя по типу
func (r *NoteRepositoryImpl) GetByType(ctx context.Context, userID int64, noteType domain.NoteType) ([]*domain.Note, error) {
	query := `
		SELECT id, title, content, type, category, url, file_id, file_name, file_size,
		       tags, is_favorite, created_at, updated_at, user_id
		FROM notes WHERE user_id = $1 AND type = $2 ORDER BY created_at DESC`

	rows, err := r.db.DB.QueryContext(ctx, query, userID, noteType)
	if err != nil {
		return nil, fmt.Errorf("failed to get notes by type: %w", err)
	}
	defer rows.Close()

	return r.scanNotes(rows)
}

// GetFavorites получает избранные заметки пользователя
func (r *NoteRepositoryImpl) GetFavorites(ctx context.Context, userID int64) ([]*domain.Note, error) {
	query := `
		SELECT id, title, content, type, category, url, file_id, file_name, file_size,
		       tags, is_favorite, created_at, updated_at, user_id
		FROM notes WHERE user_id = $1 AND is_favorite = true ORDER BY created_at DESC`

	rows, err := r.db.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get favorite notes: %w", err)
	}
	defer rows.Close()

	return r.scanNotes(rows)
}

// Search выполняет поиск заметок по запросу
func (r *NoteRepositoryImpl) Search(ctx context.Context, userID int64, query string) ([]*domain.Note, error) {
	searchQuery := "%" + strings.ToLower(query) + "%"

	sqlQuery := `
		SELECT id, title, content, type, category, url, file_id, file_name, file_size,
		       tags, is_favorite, created_at, updated_at, user_id
		FROM notes 
		WHERE user_id = $1 AND (
			LOWER(title) LIKE $2 OR 
			LOWER(content) LIKE $2 OR 
			LOWER(tags) LIKE $2
		) ORDER BY created_at DESC`

	rows, err := r.db.DB.QueryContext(ctx, sqlQuery, userID, searchQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to search notes: %w", err)
	}
	defer rows.Close()

	return r.scanNotes(rows)
}

// Update обновляет заметку
func (r *NoteRepositoryImpl) Update(ctx context.Context, note *domain.Note) error {
	note.UpdatedAt = time.Now()

	query := `
		UPDATE notes SET 
			title = $1, content = $2, type = $3, category = $4, url = $5,
			file_id = $6, file_name = $7, file_size = $8, tags = $9,
			is_favorite = $10, updated_at = $11
		WHERE id = $12`

	_, err := r.db.DB.ExecContext(ctx, query,
		note.Title, note.Content, note.Type, note.Category, note.URL,
		note.FileID, note.FileName, note.FileSize, note.Tags,
		note.IsFavorite, note.UpdatedAt, note.ID)

	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	return nil
}

// Delete удаляет заметку
func (r *NoteRepositoryImpl) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM notes WHERE id = $1`

	_, err := r.db.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	return nil
}

// scanNotes сканирует строки и возвращает массив заметок
func (r *NoteRepositoryImpl) scanNotes(rows *sql.Rows) ([]*domain.Note, error) {
	var notes []*domain.Note

	for rows.Next() {
		note := &domain.Note{}
		err := rows.Scan(
			&note.ID, &note.Title, &note.Content, &note.Type, &note.Category,
			&note.URL, &note.FileID, &note.FileName, &note.FileSize, &note.Tags,
			&note.IsFavorite, &note.CreatedAt, &note.UpdatedAt, &note.UserID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}
		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return notes, nil
}
