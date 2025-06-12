package usecase

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"todolist/internal/domain"
)

// NoteService предоставляет бизнес-логику для работы с заметками
type NoteService struct {
	noteRepo domain.NoteRepository
}

// NewNoteService создает новый экземпляр NoteService
func NewNoteService(noteRepo domain.NoteRepository) *NoteService {
	return &NoteService{
		noteRepo: noteRepo,
	}
}

// CreateNote создает новую заметку
func (s *NoteService) CreateNote(ctx context.Context, userID int64, title, content, category, tags string) (*domain.Note, error) {
	note := &domain.Note{
		Title:    title,
		Content:  content,
		Category: domain.NoteCategory(category),
		Tags:     tags,
		UserID:   userID,
	}

	// Определяем тип заметки
	note.Type = s.determineNoteType(content)

	// Если это ссылка, извлекаем URL
	if note.Type == domain.NoteTypeLink {
		note.URL = s.extractURL(content)
	}

	// Устанавливаем категорию по умолчанию
	if note.Category == "" {
		note.Category = domain.NoteCategoryGeneral
	}

	err := s.noteRepo.Create(ctx, note)
	if err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	return note, nil
}

// CreateNoteFromFile создает заметку из файла
func (s *NoteService) CreateNoteFromFile(ctx context.Context, userID int64, title, fileID, fileName string, fileSize int64, noteType domain.NoteType, category, tags string) (*domain.Note, error) {
	note := &domain.Note{
		Title:    title,
		Type:     noteType,
		Category: domain.NoteCategory(category),
		FileID:   fileID,
		FileName: fileName,
		FileSize: fileSize,
		Tags:     tags,
		UserID:   userID,
	}

	// Устанавливаем категорию по умолчанию
	if note.Category == "" {
		note.Category = domain.NoteCategoryGeneral
	}

	err := s.noteRepo.Create(ctx, note)
	if err != nil {
		return nil, fmt.Errorf("failed to create note from file: %w", err)
	}

	return note, nil
}

// GetNote получает заметку по ID
func (s *NoteService) GetNote(ctx context.Context, id int) (*domain.Note, error) {
	note, err := s.noteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	return note, nil
}

// GetUserNotes получает все заметки пользователя
func (s *NoteService) GetUserNotes(ctx context.Context, userID int64) ([]*domain.Note, error) {
	notes, err := s.noteRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user notes: %w", err)
	}

	return notes, nil
}

// GetNotesByCategory получает заметки по категории
func (s *NoteService) GetNotesByCategory(ctx context.Context, userID int64, category domain.NoteCategory) ([]*domain.Note, error) {
	notes, err := s.noteRepo.GetByCategory(ctx, userID, category)
	if err != nil {
		return nil, fmt.Errorf("failed to get notes by category: %w", err)
	}

	return notes, nil
}

// GetNotesByType получает заметки по типу
func (s *NoteService) GetNotesByType(ctx context.Context, userID int64, noteType domain.NoteType) ([]*domain.Note, error) {
	notes, err := s.noteRepo.GetByType(ctx, userID, noteType)
	if err != nil {
		return nil, fmt.Errorf("failed to get notes by type: %w", err)
	}

	return notes, nil
}

// GetFavoriteNotes получает избранные заметки
func (s *NoteService) GetFavoriteNotes(ctx context.Context, userID int64) ([]*domain.Note, error) {
	notes, err := s.noteRepo.GetFavorites(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get favorite notes: %w", err)
	}

	return notes, nil
}

// SearchNotes выполняет поиск заметок
func (s *NoteService) SearchNotes(ctx context.Context, userID int64, query string) ([]*domain.Note, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	notes, err := s.noteRepo.Search(ctx, userID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search notes: %w", err)
	}

	return notes, nil
}

// UpdateNote обновляет заметку
func (s *NoteService) UpdateNote(ctx context.Context, note *domain.Note) error {
	// Обновляем время изменения
	note.UpdatedAt = time.Now()

	// Переопределяем тип заметки при обновлении содержимого
	if note.Content != "" {
		note.Type = s.determineNoteType(note.Content)
		if note.Type == domain.NoteTypeLink {
			note.URL = s.extractURL(note.Content)
		}
	}

	err := s.noteRepo.Update(ctx, note)
	if err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}

	return nil
}

// ToggleFavorite переключает статус избранного для заметки
func (s *NoteService) ToggleFavorite(ctx context.Context, noteID int) (*domain.Note, error) {
	note, err := s.noteRepo.GetByID(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	note.ToggleFavorite()

	err = s.noteRepo.Update(ctx, note)
	if err != nil {
		return nil, fmt.Errorf("failed to toggle favorite: %w", err)
	}

	return note, nil
}

// DeleteNote удаляет заметку
func (s *NoteService) DeleteNote(ctx context.Context, id int) error {
	err := s.noteRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	return nil
}

// determineNoteType определяет тип заметки по содержимому
func (s *NoteService) determineNoteType(content string) domain.NoteType {
	// Проверяем, является ли содержимое ссылкой
	if s.isURL(content) {
		return domain.NoteTypeLink
	}

	return domain.NoteTypeText
}

// isURL проверяет, является ли строка URL
func (s *NoteService) isURL(str string) bool {
	// Простая проверка на URL
	urlRegex := regexp.MustCompile(`^https?://`)
	if urlRegex.MatchString(str) {
		return true
	}

	// Проверяем через стандартную библиотеку
	_, err := url.ParseRequestURI(str)
	return err == nil && (strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://"))
}

// extractURL извлекает URL из содержимого
func (s *NoteService) extractURL(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if s.isURL(line) {
			return line
		}
	}
	return ""
}

// FormatNoteForDisplay форматирует заметку для отображения
func (s *NoteService) FormatNoteForDisplay(note *domain.Note) string {
	var builder strings.Builder

	// Заголовок с иконкой типа
	builder.WriteString(fmt.Sprintf("%s *%s*\n", note.GetDisplayType(), note.Title))

	// Категория
	if note.Category != domain.NoteCategoryGeneral {
		builder.WriteString(fmt.Sprintf("Категория: %s\n", note.GetDisplayCategory()))
	}

	// Содержимое
	if note.Content != "" {
		if len(note.Content) > 300 {
			builder.WriteString(fmt.Sprintf("```\n%s...\n```\n", note.Content[:300]))
		} else {
			builder.WriteString(fmt.Sprintf("```\n%s\n```\n", note.Content))
		}
	}

	// URL для ссылок
	if note.IsLink() && note.URL != "" {
		builder.WriteString(fmt.Sprintf("🔗 [Перейти по ссылке](%s)\n", note.URL))
	}

	// Информация о файле
	if note.IsFile() {
		builder.WriteString(fmt.Sprintf("📎 %s", note.FileName))
		if note.FileSize > 0 {
			builder.WriteString(fmt.Sprintf(" (%.1f KB)", float64(note.FileSize)/1024))
		}
		builder.WriteString("\n")
	}

	// Теги
	if note.Tags != "" {
		builder.WriteString(fmt.Sprintf("🏷️ %s\n", note.Tags))
	}

	// Избранное
	if note.IsFavorite {
		builder.WriteString("⭐ Избранное\n")
	}

	// Дата создания
	builder.WriteString(fmt.Sprintf("📅 %s", note.CreatedAt.Format("02.01.2006 15:04")))

	return builder.String()
}
