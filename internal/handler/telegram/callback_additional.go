package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"todolist/internal/domain"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleSearchCallback обрабатывает кнопку поиска заметок
func (b *Bot) handleSearchCallback(ctx context.Context, chatID, userID int64) {
	b.userStates[userID] = &UserState{
		Action:   "search_notes",
		Step:     1,
		NoteData: make(map[string]string),
	}

	text := "🔍 *Поиск заметок*\n\nВведите поисковый запрос:"
	keyboard := getBackToMenuKeyboard()
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

// handleFavoritesCallback обрабатывает кнопку избранных заметок
func (b *Bot) handleFavoritesCallback(ctx context.Context, chatID, userID int64) {
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
		text := "⭐ У вас пока нет избранных заметок\n\nДобавьте заметки в избранное для быстрого доступа!"
		keyboard := getBackToMenuKeyboard()
		b.sendMessageWithKeyboard(chatID, text, keyboard)
		return
	}

	// Конвертируем заметки в формат для клавиатуры
	var noteItems []NoteListItem
	for _, note := range notes {
		noteItems = append(noteItems, NoteListItem{
			ID:         note.ID,
			Title:      note.Title,
			IsFavorite: true,
		})
	}

	text := fmt.Sprintf("⭐ *Избранные заметки* (%d)\n\nВаши любимые заметки:", len(notes))
	keyboard := getNoteListKeyboard(noteItems)
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

// handleShowNoteCallback обрабатывает показ заметки
func (b *Bot) handleShowNoteCallback(ctx context.Context, query *tgbotapi.CallbackQuery, user *domain.User) {
	chatID := query.Message.Chat.ID
	noteIDStr := strings.TrimPrefix(query.Data, "show_note_")
	noteID, err := strconv.Atoi(noteIDStr)
	if err != nil {
		b.sendMessage(chatID, "❌ Неверный ID заметки")
		return
	}

	note, err := b.noteService.GetNote(ctx, noteID)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка: %s", err.Error()))
		return
	}

	if note.UserID != user.ID {
		b.sendMessage(chatID, "❌ Заметка не найдена")
		return
	}

	text := b.noteService.FormatNoteForDisplay(note)
	keyboard := getNoteActionsKeyboard(noteID, note.IsFavorite)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	b.api.Send(msg)
}

// handleDeleteNoteCallback обрабатывает удаление заметки
func (b *Bot) handleDeleteNoteCallback(ctx context.Context, query *tgbotapi.CallbackQuery, user *domain.User) {
	chatID := query.Message.Chat.ID
	noteIDStr := strings.TrimPrefix(query.Data, "delete_note_")
	noteID, err := strconv.Atoi(noteIDStr)
	if err != nil {
		b.sendMessage(chatID, "❌ Неверный ID заметки")
		return
	}

	text := "🗑️ *Удаление заметки*\n\nВы уверены, что хотите удалить эту заметку?"
	keyboard := getConfirmationKeyboard("delete_note", noteID)
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

// handleAddFavoriteCallback обрабатывает добавление в избранное
func (b *Bot) handleAddFavoriteCallback(ctx context.Context, query *tgbotapi.CallbackQuery, user *domain.User) {
	chatID := query.Message.Chat.ID
	noteIDStr := strings.TrimPrefix(query.Data, "favorite_add_")
	noteID, err := strconv.Atoi(noteIDStr)
	if err != nil {
		b.sendMessage(chatID, "❌ Неверный ID заметки")
		return
	}

	// Проверяем принадлежность заметки пользователю
	note, err := b.noteService.GetNote(ctx, noteID)
	if err != nil || note.UserID != user.ID {
		b.sendMessage(chatID, "❌ Заметка не найдена")
		return
	}

	updatedNote, err := b.noteService.ToggleFavorite(ctx, noteID)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка: %s", err.Error()))
		return
	}

	text := fmt.Sprintf("⭐ *Заметка добавлена в избранное!*\n\n[%d] %s", noteID, updatedNote.Title)
	keyboard := getNoteActionsKeyboard(noteID, updatedNote.IsFavorite)
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

// handleRemoveFavoriteCallback обрабатывает удаление из избранного
func (b *Bot) handleRemoveFavoriteCallback(ctx context.Context, query *tgbotapi.CallbackQuery, user *domain.User) {
	chatID := query.Message.Chat.ID
	noteIDStr := strings.TrimPrefix(query.Data, "favorite_remove_")
	noteID, err := strconv.Atoi(noteIDStr)
	if err != nil {
		b.sendMessage(chatID, "❌ Неверный ID заметки")
		return
	}

	// Проверяем принадлежность заметки пользователю
	note, err := b.noteService.GetNote(ctx, noteID)
	if err != nil || note.UserID != user.ID {
		b.sendMessage(chatID, "❌ Заметка не найдена")
		return
	}

	updatedNote, err := b.noteService.ToggleFavorite(ctx, noteID)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка: %s", err.Error()))
		return
	}

	text := fmt.Sprintf("✨ *Заметка убрана из избранного*\n\n[%d] %s", noteID, updatedNote.Title)
	keyboard := getNoteActionsKeyboard(noteID, updatedNote.IsFavorite)
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

// handleDeleteTaskCallback обрабатывает удаление задачи
func (b *Bot) handleDeleteTaskCallback(ctx context.Context, query *tgbotapi.CallbackQuery, user *domain.User) {
	chatID := query.Message.Chat.ID
	taskIDStr := strings.TrimPrefix(query.Data, "delete_")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		b.sendMessage(chatID, "❌ Неверный ID задачи")
		return
	}

	text := "🗑️ *Удаление задачи*\n\nВы уверены, что хотите удалить эту задачу?"
	keyboard := getConfirmationKeyboard("delete_task", taskID)
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

// handleNotifyTaskCallback обрабатывает установку напоминания для задачи
func (b *Bot) handleNotifyTaskCallback(ctx context.Context, query *tgbotapi.CallbackQuery, user *domain.User) {
	chatID := query.Message.Chat.ID
	userID := query.From.ID
	taskIDStr := strings.TrimPrefix(query.Data, "notify_")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		b.sendMessage(chatID, "❌ Неверный ID задачи")
		return
	}

	// Запускаем интерактивную настройку уведомления
	b.userStates[userID] = &UserState{
		Action:   "set_notification",
		Step:     1,
		TaskID:   taskID,
		TaskData: make(map[string]string),
	}

	text := "⏰ *Настройка напоминания*\n\nВведите время уведомления:\n\n*Примеры:*\n• 15:30 - сегодня в 15:30\n• завтра 10:00\n• 25.12 14:00"
	keyboard := getBackToMenuKeyboard()
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

// handlePriorityCallback обрабатывает выбор приоритета
func (b *Bot) handlePriorityCallback(ctx context.Context, query *tgbotapi.CallbackQuery, user *domain.User) {
	chatID := query.Message.Chat.ID
	userID := query.From.ID
	priority := strings.TrimPrefix(query.Data, "priority_")

	if state, exists := b.userStates[userID]; exists && state.Action == "add_task" && state.Step == 3 {
		state.TaskData["priority"] = priority
		b.handleAddTaskState(ctx, &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: chatID},
			From: &tgbotapi.User{ID: userID},
			Text: priority,
		}, user, state)
	}
}

// handleCategoryCallback обрабатывает выбор категории заметки
func (b *Bot) handleCategoryCallback(ctx context.Context, query *tgbotapi.CallbackQuery, user *domain.User) {
	chatID := query.Message.Chat.ID
	userID := query.From.ID
	category := strings.TrimPrefix(query.Data, "category_")

	if state, exists := b.userStates[userID]; exists && state.Action == "add_note" && state.Step == 3 {
		state.NoteData["category"] = category
		state.Step = 4
		text := "4️⃣ Введите теги через запятую (или отправьте \"-\" чтобы пропустить):"
		keyboard := getBackToMenuKeyboard()
		b.sendMessageWithKeyboard(chatID, text, keyboard)
	}
}

// handleConfirmCallback обрабатывает подтверждение действий
func (b *Bot) handleConfirmCallback(ctx context.Context, query *tgbotapi.CallbackQuery, user *domain.User) {
	chatID := query.Message.Chat.ID
	userID := query.From.ID
	data := strings.TrimPrefix(query.Data, "confirm_")

	parts := strings.Split(data, "_")
	if len(parts) < 2 {
		b.sendMessage(chatID, "❌ Неверный формат команды")
		return
	}

	action := parts[0]

	switch action {
	case "delete":
		if len(parts) >= 3 && parts[1] == "task" {
			taskID, err := strconv.Atoi(parts[2])
			if err != nil {
				b.sendMessage(chatID, "❌ Неверный ID задачи")
				return
			}

			err = b.taskService.DeleteTask(ctx, taskID, user.ID)
			if err != nil {
				b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка: %s", err.Error()))
				return
			}

			text := fmt.Sprintf("🗑️ *Задача [%d] удалена!*", taskID)
			keyboard := tgbotapi.InlineKeyboardMarkup{
				InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
					{
						tgbotapi.InlineKeyboardButton{Text: "📋 К задачам", CallbackData: &[]string{"cmd_tasks"}[0]},
						tgbotapi.InlineKeyboardButton{Text: "🏠 Главное меню", CallbackData: &[]string{"cmd_menu"}[0]},
					},
				},
			}
			b.sendMessageWithKeyboard(chatID, text, keyboard)

		} else if len(parts) >= 3 && parts[1] == "note" {
			noteID, err := strconv.Atoi(parts[2])
			if err != nil {
				b.sendMessage(chatID, "❌ Неверный ID заметки")
				return
			}

			err = b.noteService.DeleteNote(ctx, noteID)
			if err != nil {
				b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка: %s", err.Error()))
				return
			}

			text := fmt.Sprintf("🗑️ *Заметка [%d] удалена!*", noteID)
			keyboard := tgbotapi.InlineKeyboardMarkup{
				InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
					{
						tgbotapi.InlineKeyboardButton{Text: "📝 К заметкам", CallbackData: &[]string{"cmd_notes"}[0]},
						tgbotapi.InlineKeyboardButton{Text: "🏠 Главное меню", CallbackData: &[]string{"cmd_menu"}[0]},
					},
				},
			}
			b.sendMessageWithKeyboard(chatID, text, keyboard)
		}

	case "logout":
		err := b.authService.Logout(ctx, userID)
		if err != nil {
			b.sendMessage(chatID, "❌ Ошибка при выходе")
			return
		}

		// Удаляем состояние пользователя
		delete(b.userStates, userID)

		b.sendMessage(chatID, "👋 Вы вышли из системы. Для повторной авторизации отправьте /start пароль")
	}
}

// handleCancelCallback обрабатывает отмену действий
func (b *Bot) handleCancelCallback(ctx context.Context, query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID

	text := "❌ *Действие отменено*\n\nВозвращаемся в главное меню:"
	keyboard := getMainMenuKeyboard()
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}
