package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"todolist/internal/domain"
)

// handleListNotesCommand обрабатывает команду /notes
func (b *Bot) handleListNotesCommand(ctx context.Context, chatID, userID int64) {
	user, err := b.getUserFromTelegram(ctx, userID)
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка авторизации")
		return
	}

	notes, err := b.noteService.GetUserNotes(ctx, user.ID)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка получения заметок: %s", err.Error()))
		return
	}

	if len(notes) == 0 {
		b.sendMessage(chatID, "📝 У вас пока нет заметок.\n\nИспользуйте /note для создания новой заметки или отправьте документ/изображение.")
		return
	}

	var response strings.Builder
	response.WriteString("📚 *Ваши заметки:*\n\n")

	for i, note := range notes {
		if i >= 10 { // Ограничиваем количество заметок в списке
			response.WriteString(fmt.Sprintf("... и еще %d заметок\n", len(notes)-i))
			break
		}

		response.WriteString(fmt.Sprintf("%s [%d] %s\n", note.GetDisplayType(), note.ID, note.Title))
		if note.IsFavorite {
			response.WriteString("⭐")
		}
		response.WriteString(fmt.Sprintf("📅 %s\n\n", note.CreatedAt.Format("02.01.2006")))
	}

	response.WriteString("💡 Используйте /nshow ID для просмотра заметки")

	msg := tgbotapi.NewMessage(chatID, response.String())
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}

// handleAddNoteCommand обрабатывает команду /note
func (b *Bot) handleAddNoteCommand(ctx context.Context, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID

	user, err := b.getUserFromTelegram(ctx, userID)
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка авторизации")
		return
	}

	args := strings.Fields(message.Text)
	if len(args) < 2 {
		// Запускаем интерактивное создание заметки
		b.userStates[userID] = &UserState{
			Action:   "add_note",
			Step:     1,
			NoteData: make(map[string]string),
		}
		b.sendMessage(chatID, "📝 Создание новой заметки\n\n1️⃣ Введите заголовок заметки:")
		return
	}

	title := strings.Join(args[1:], " ")

	note, err := b.noteService.CreateNote(ctx, user.ID, title, "", "general", "")
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка создания заметки: %s", err.Error()))
		return
	}

	b.sendMessage(chatID, fmt.Sprintf("✅ Заметка [%d] создана!\n📝 %s", note.ID, note.Title))
}

// handleShowNoteCommand обрабатывает команду /nshow
func (b *Bot) handleShowNoteCommand(ctx context.Context, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID

	user, err := b.getUserFromTelegram(ctx, userID)
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка авторизации")
		return
	}

	args := strings.Fields(message.Text)
	if len(args) < 2 {
		b.sendMessage(chatID, "❌ Укажите ID заметки: /nshow 123")
		return
	}

	noteID, err := strconv.Atoi(args[1])
	if err != nil {
		b.sendMessage(chatID, "❌ Неверный ID заметки")
		return
	}

	note, err := b.noteService.GetNote(ctx, noteID)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка: %s", err.Error()))
		return
	}

	// Проверяем принадлежность заметки пользователю
	if note.UserID != user.ID {
		b.sendMessage(chatID, "❌ Заметка не найдена")
		return
	}

	response := b.noteService.FormatNoteForDisplay(note)

	msg := tgbotapi.NewMessage(chatID, response)
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}

// handleDeleteNoteCommand обрабатывает команду /ndelete
func (b *Bot) handleDeleteNoteCommand(ctx context.Context, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID

	user, err := b.getUserFromTelegram(ctx, userID)
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка авторизации")
		return
	}

	args := strings.Fields(message.Text)
	if len(args) < 2 {
		b.sendMessage(chatID, "❌ Укажите ID заметки: /ndelete 123")
		return
	}

	noteID, err := strconv.Atoi(args[1])
	if err != nil {
		b.sendMessage(chatID, "❌ Неверный ID заметки")
		return
	}

	// Проверяем принадлежность заметки пользователю
	note, err := b.noteService.GetNote(ctx, noteID)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка: %s", err.Error()))
		return
	}

	if note.UserID != user.ID {
		b.sendMessage(chatID, "❌ Заметка не найдена")
		return
	}

	err = b.noteService.DeleteNote(ctx, noteID)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка удаления: %s", err.Error()))
		return
	}

	b.sendMessage(chatID, fmt.Sprintf("🗑️ Заметка [%d] удалена!", noteID))
}

// handleFavoriteNotesCommand обрабатывает команду /favorites
func (b *Bot) handleFavoriteNotesCommand(ctx context.Context, chatID, userID int64) {
	user, err := b.getUserFromTelegram(ctx, userID)
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка авторизации")
		return
	}

	notes, err := b.noteService.GetFavoriteNotes(ctx, user.ID)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка получения избранных заметок: %s", err.Error()))
		return
	}

	if len(notes) == 0 {
		b.sendMessage(chatID, "⭐ У вас пока нет избранных заметок.\n\nИспользуйте /favorite ID для добавления заметки в избранное.")
		return
	}

	var response strings.Builder
	response.WriteString("⭐ *Избранные заметки:*\n\n")

	for _, note := range notes {
		response.WriteString(fmt.Sprintf("%s [%d] %s\n", note.GetDisplayType(), note.ID, note.Title))
		response.WriteString(fmt.Sprintf("📅 %s\n\n", note.CreatedAt.Format("02.01.2006")))
	}

	msg := tgbotapi.NewMessage(chatID, response.String())
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}

// handleToggleFavoriteCommand обрабатывает команду /favorite
func (b *Bot) handleToggleFavoriteCommand(ctx context.Context, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID

	user, err := b.getUserFromTelegram(ctx, userID)
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка авторизации")
		return
	}

	args := strings.Fields(message.Text)
	if len(args) < 2 {
		b.sendMessage(chatID, "❌ Укажите ID заметки: /favorite 123")
		return
	}

	noteID, err := strconv.Atoi(args[1])
	if err != nil {
		b.sendMessage(chatID, "❌ Неверный ID заметки")
		return
	}

	// Проверяем принадлежность заметки пользователю
	note, err := b.noteService.GetNote(ctx, noteID)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка: %s", err.Error()))
		return
	}

	if note.UserID != user.ID {
		b.sendMessage(chatID, "❌ Заметка не найдена")
		return
	}

	updatedNote, err := b.noteService.ToggleFavorite(ctx, noteID)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка: %s", err.Error()))
		return
	}

	if updatedNote.IsFavorite {
		b.sendMessage(chatID, fmt.Sprintf("⭐ Заметка [%d] добавлена в избранное!", noteID))
	} else {
		b.sendMessage(chatID, fmt.Sprintf("✨ Заметка [%d] убрана из избранного", noteID))
	}
}

// handleSearchNotesCommand обрабатывает команду /search
func (b *Bot) handleSearchNotesCommand(ctx context.Context, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID

	user, err := b.getUserFromTelegram(ctx, userID)
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка авторизации")
		return
	}

	args := strings.Fields(message.Text)
	if len(args) < 2 {
		b.sendMessage(chatID, "❌ Укажите поисковый запрос: /search текст")
		return
	}

	query := strings.Join(args[1:], " ")

	notes, err := b.noteService.SearchNotes(ctx, user.ID, query)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка поиска: %s", err.Error()))
		return
	}

	if len(notes) == 0 {
		b.sendMessage(chatID, fmt.Sprintf("🔍 По запросу \"%s\" ничего не найдено.", query))
		return
	}

	var response strings.Builder
	response.WriteString(fmt.Sprintf("🔍 *Результаты поиска по \"%s\":*\n\n", query))

	for i, note := range notes {
		if i >= 10 { // Ограничиваем количество результатов
			response.WriteString(fmt.Sprintf("... и еще %d результатов\n", len(notes)-i))
			break
		}

		response.WriteString(fmt.Sprintf("%s [%d] %s\n", note.GetDisplayType(), note.ID, note.Title))
		if note.IsFavorite {
			response.WriteString("⭐")
		}
		response.WriteString(fmt.Sprintf("📅 %s\n\n", note.CreatedAt.Format("02.01.2006")))
	}

	msg := tgbotapi.NewMessage(chatID, response.String())
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}

// handleLinkNotesCommand обрабатывает команду /links
func (b *Bot) handleLinkNotesCommand(ctx context.Context, chatID, userID int64) {
	user, err := b.getUserFromTelegram(ctx, userID)
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка авторизации")
		return
	}

	notes, err := b.noteService.GetNotesByType(ctx, user.ID, domain.NoteTypeLink)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка получения ссылок: %s", err.Error()))
		return
	}

	if len(notes) == 0 {
		b.sendMessage(chatID, "🔗 У вас пока нет сохраненных ссылок.\n\nОтправьте ссылку как текст для автоматического сохранения.")
		return
	}

	var response strings.Builder
	response.WriteString("🔗 *Сохраненные ссылки:*\n\n")

	for i, note := range notes {
		if i >= 10 {
			response.WriteString(fmt.Sprintf("... и еще %d ссылок\n", len(notes)-i))
			break
		}

		response.WriteString(fmt.Sprintf("[%d] %s\n", note.ID, note.Title))
		if note.URL != "" {
			response.WriteString(fmt.Sprintf("🔗 [Перейти](%s)\n", note.URL))
		}
		if note.IsFavorite {
			response.WriteString("⭐")
		}
		response.WriteString(fmt.Sprintf("📅 %s\n\n", note.CreatedAt.Format("02.01.2006")))
	}

	msg := tgbotapi.NewMessage(chatID, response.String())
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}

// handleFileNotesCommand обрабатывает команду /files
func (b *Bot) handleFileNotesCommand(ctx context.Context, chatID, userID int64) {
	user, err := b.getUserFromTelegram(ctx, userID)
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка авторизации")
		return
	}

	// Получаем заметки с файлами (документы, изображения, видео, аудио)
	var allFiles []*domain.Note

	for _, noteType := range []domain.NoteType{domain.NoteTypeDocument, domain.NoteTypeImage, domain.NoteTypeVideo, domain.NoteTypeAudio} {
		notes, err := b.noteService.GetNotesByType(ctx, user.ID, noteType)
		if err != nil {
			b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка получения файлов: %s", err.Error()))
			return
		}
		allFiles = append(allFiles, notes...)
	}

	if len(allFiles) == 0 {
		b.sendMessage(chatID, "📎 У вас пока нет сохраненных файлов.\n\nОтправьте документ или изображение для автоматического сохранения.")
		return
	}

	var response strings.Builder
	response.WriteString("📎 *Сохраненные файлы:*\n\n")

	for i, note := range allFiles {
		if i >= 10 {
			response.WriteString(fmt.Sprintf("... и еще %d файлов\n", len(allFiles)-i))
			break
		}

		response.WriteString(fmt.Sprintf("%s [%d] %s\n", note.GetDisplayType(), note.ID, note.Title))
		if note.FileName != "" {
			response.WriteString(fmt.Sprintf("📄 %s", note.FileName))
			if note.FileSize > 0 {
				response.WriteString(fmt.Sprintf(" (%.1f KB)", float64(note.FileSize)/1024))
			}
			response.WriteString("\n")
		}
		if note.IsFavorite {
			response.WriteString("⭐")
		}
		response.WriteString(fmt.Sprintf("📅 %s\n\n", note.CreatedAt.Format("02.01.2006")))
	}

	msg := tgbotapi.NewMessage(chatID, response.String())
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}

// handleCreateNoteFromFile создает заметку из файла
func (b *Bot) handleCreateNoteFromFile(ctx context.Context, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID

	user, err := b.getUserFromTelegram(ctx, userID)
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка авторизации")
		return
	}

	var fileID, fileName string
	var fileSize int64
	var noteType domain.NoteType
	var title string

	// Определяем тип файла и получаем информацию
	if message.Document != nil {
		fileID = message.Document.FileID
		fileName = message.Document.FileName
		fileSize = int64(message.Document.FileSize)
		noteType = domain.NoteTypeDocument
		title = fileName
		if message.Caption != "" {
			title = message.Caption
		}
	} else if len(message.Photo) > 0 {
		// Берем самое большое изображение
		photo := message.Photo[len(message.Photo)-1]
		fileID = photo.FileID
		fileName = fmt.Sprintf("photo_%s.jpg", photo.FileUniqueID)
		fileSize = int64(photo.FileSize)
		noteType = domain.NoteTypeImage
		title = "Изображение"
		if message.Caption != "" {
			title = message.Caption
		}
	} else if message.Video != nil {
		fileID = message.Video.FileID
		fileName = message.Video.FileName
		if fileName == "" {
			fileName = fmt.Sprintf("video_%s.mp4", message.Video.FileUniqueID)
		}
		fileSize = int64(message.Video.FileSize)
		noteType = domain.NoteTypeVideo
		title = fileName
		if message.Caption != "" {
			title = message.Caption
		}
	} else if message.Audio != nil {
		fileID = message.Audio.FileID
		fileName = message.Audio.FileName
		if fileName == "" {
			fileName = fmt.Sprintf("audio_%s", message.Audio.FileUniqueID)
		}
		fileSize = int64(message.Audio.FileSize)
		noteType = domain.NoteTypeAudio
		title = fileName
		if message.Caption != "" {
			title = message.Caption
		}
	} else if message.Voice != nil {
		fileID = message.Voice.FileID
		fileName = fmt.Sprintf("voice_%s.ogg", message.Voice.FileUniqueID)
		fileSize = int64(message.Voice.FileSize)
		noteType = domain.NoteTypeAudio
		title = "Голосовое сообщение"
	} else {
		return // Неподдерживаемый тип файла
	}

	note, err := b.noteService.CreateNoteFromFile(ctx, user.ID, title, fileID, fileName, fileSize, noteType, "general", "")
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка создания заметки: %s", err.Error()))
		return
	}

	response := fmt.Sprintf("✅ %s [%d] сохранен!\n📎 %s", note.GetDisplayType(), note.ID, note.Title)
	if note.FileSize > 0 {
		response += fmt.Sprintf(" (%.1f KB)", float64(note.FileSize)/1024)
	}

	b.sendMessage(chatID, response)
}
