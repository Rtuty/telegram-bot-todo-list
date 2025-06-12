package domain

import "time"

// NoteType представляет тип заметки
type NoteType string

const (
	NoteTypeText     NoteType = "text"
	NoteTypeLink     NoteType = "link"
	NoteTypeDocument NoteType = "document"
	NoteTypeImage    NoteType = "image"
	NoteTypeVideo    NoteType = "video"
	NoteTypeAudio    NoteType = "audio"
)

// NoteCategory представляет категорию заметки
type NoteCategory string

const (
	NoteCategoryGeneral   NoteCategory = "general"
	NoteCategoryWork      NoteCategory = "work"
	NoteCategoryStudy     NoteCategory = "study"
	NoteCategoryPersonal  NoteCategory = "personal"
	NoteCategoryResources NoteCategory = "resources"
	NoteCategoryIdeas     NoteCategory = "ideas"
)

// Note представляет заметку/полезную информацию
type Note struct {
	ID         int          `json:"id" db:"id"`
	Title      string       `json:"title" db:"title"`
	Content    string       `json:"content" db:"content"`
	Type       NoteType     `json:"type" db:"type"`
	Category   NoteCategory `json:"category" db:"category"`
	URL        string       `json:"url,omitempty" db:"url"`
	FileID     string       `json:"file_id,omitempty" db:"file_id"`
	FileName   string       `json:"file_name,omitempty" db:"file_name"`
	FileSize   int64        `json:"file_size,omitempty" db:"file_size"`
	Tags       string       `json:"tags,omitempty" db:"tags"`
	IsFavorite bool         `json:"is_favorite" db:"is_favorite"`
	CreatedAt  time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at" db:"updated_at"`
	UserID     int64        `json:"user_id" db:"user_id"`
}

// IsLink проверяет, является ли заметка ссылкой
func (n *Note) IsLink() bool {
	return n.Type == NoteTypeLink
}

// IsFile проверяет, является ли заметка файлом
func (n *Note) IsFile() bool {
	return n.Type == NoteTypeDocument || n.Type == NoteTypeImage ||
		n.Type == NoteTypeVideo || n.Type == NoteTypeAudio
}

// ToggleFavorite переключает статус избранного
func (n *Note) ToggleFavorite() {
	n.IsFavorite = !n.IsFavorite
	n.UpdatedAt = time.Now()
}

// GetDisplayType возвращает отображаемый тип заметки
func (n *Note) GetDisplayType() string {
	switch n.Type {
	case NoteTypeText:
		return "📝 Текст"
	case NoteTypeLink:
		return "🔗 Ссылка"
	case NoteTypeDocument:
		return "📄 Документ"
	case NoteTypeImage:
		return "🖼️ Изображение"
	case NoteTypeVideo:
		return "🎥 Видео"
	case NoteTypeAudio:
		return "🎵 Аудио"
	default:
		return "📝 Заметка"
	}
}

// GetDisplayCategory возвращает отображаемую категорию заметки
func (n *Note) GetDisplayCategory() string {
	switch n.Category {
	case NoteCategoryGeneral:
		return "🗂️ Общее"
	case NoteCategoryWork:
		return "💼 Работа"
	case NoteCategoryStudy:
		return "📚 Учеба"
	case NoteCategoryPersonal:
		return "👤 Личное"
	case NoteCategoryResources:
		return "🔗 Ресурсы"
	case NoteCategoryIdeas:
		return "💡 Идеи"
	default:
		return "🗂️ Общее"
	}
}
