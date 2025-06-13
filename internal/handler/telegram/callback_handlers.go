package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"todolist/internal/domain"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleCallbackQuery обрабатывает callback запросы от inline клавиатур
func (b *Bot) handleCallbackQuery(ctx context.Context, query *tgbotapi.CallbackQuery) {
	chatID := query.Message.Chat.ID
	userID := query.From.ID
	data := query.Data

	// Подтверждаем получение callback
	callback := tgbotapi.NewCallback(query.ID, "")
	b.api.Request(callback)

	// Получаем пользователя
	user, err := b.getUserFromTelegram(ctx, userID)
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка авторизации")
		return
	}

	// Обрабатываем команды
	switch {
	case data == "cmd_menu":
		b.handleMenuCallback(ctx, chatID, userID)
	case data == "cmd_tasks":
		b.handleTasksCallback(ctx, chatID, userID)
	case data == "cmd_add_task":
		b.handleAddTaskCallback(ctx, chatID, userID)
	case data == "cmd_notes":
		b.handleNotesCallback(ctx, chatID, userID)
	case data == "cmd_add_note":
		b.handleAddNoteCallback(ctx, chatID, userID)
	case data == "cmd_pending":
		b.handlePendingCallback(ctx, chatID, userID)
	case data == "cmd_completed":
		b.handleCompletedCallback(ctx, chatID, userID)
	case data == "cmd_search":
		b.handleSearchCallback(ctx, chatID, userID)
	case data == "cmd_favorites":
		b.handleFavoritesCallback(ctx, chatID, userID)
	case data == "cmd_help":
		b.handleHelpCallback(ctx, chatID)
	case data == "cmd_logout":
		b.handleLogoutCallback(ctx, chatID, userID)
	case strings.HasPrefix(data, "complete_"):
		b.handleCompleteTaskCallback(ctx, query, user)
	case strings.HasPrefix(data, "show_"):
		b.handleShowTaskCallback(ctx, query, user)
	case strings.HasPrefix(data, "delete_"):
		b.handleDeleteTaskCallback(ctx, query, user)
	case strings.HasPrefix(data, "notify_"):
		b.handleNotifyTaskCallback(ctx, query, user)
	case strings.HasPrefix(data, "show_note_"):
		b.handleShowNoteCallback(ctx, query, user)
	case strings.HasPrefix(data, "delete_note_"):
		b.handleDeleteNoteCallback(ctx, query, user)
	case strings.HasPrefix(data, "favorite_add_"):
		b.handleAddFavoriteCallback(ctx, query, user)
	case strings.HasPrefix(data, "favorite_remove_"):
		b.handleRemoveFavoriteCallback(ctx, query, user)
	case strings.HasPrefix(data, "priority_"):
		b.handlePriorityCallback(ctx, query, user)
	case strings.HasPrefix(data, "category_"):
		b.handleCategoryCallback(ctx, query, user)
	case strings.HasPrefix(data, "confirm_"):
		b.handleConfirmCallback(ctx, query, user)
	case strings.HasPrefix(data, "cancel_"):
		b.handleCancelCallback(ctx, query)
	default:
		b.sendMessage(chatID, "❓ Неизвестная команда")
	}
}

// handleMenuCallback обрабатывает возврат в главное меню
func (b *Bot) handleMenuCallback(ctx context.Context, chatID, userID int64) {
	text := "🏠 *Главное меню*\n\nВыберите действие:"
	keyboard := getMainMenuKeyboard()
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

// handleTasksCallback обрабатывает показ списка задач
func (b *Bot) handleTasksCallback(ctx context.Context, chatID, userID int64) {
	user, err := b.getUserFromTelegram(ctx, userID)
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка авторизации")
		return
	}

	tasks, err := b.taskService.GetTasks(ctx, user.ID)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка получения задач: %s", err.Error()))
		return
	}

	if len(tasks) == 0 {
		text := "📋 У вас пока нет задач\n\nНажмите кнопку ниже, чтобы создать первую задачу!"
		keyboard := tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{
					tgbotapi.InlineKeyboardButton{Text: "➕ Создать задачу", CallbackData: &[]string{"cmd_add_task"}[0]},
				},
				{
					tgbotapi.InlineKeyboardButton{Text: "🏠 Главное меню", CallbackData: &[]string{"cmd_menu"}[0]},
				},
			},
		}
		b.sendMessageWithKeyboard(chatID, text, keyboard)
		return
	}

	// Конвертируем задачи в формат для клавиатуры
	var taskItems []TaskListItem
	for _, task := range tasks {
		taskItems = append(taskItems, TaskListItem{
			ID:    task.ID,
			Title: task.Title,
		})
	}

	text := fmt.Sprintf("📋 *Ваши задачи* (%d)\n\nВыберите задачу для выполнения действий:", len(tasks))
	keyboard := getTaskListKeyboard(taskItems)
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

// handleAddTaskCallback обрабатывает начало создания задачи
func (b *Bot) handleAddTaskCallback(ctx context.Context, chatID, userID int64) {
	b.userStates[userID] = &UserState{
		Action:   "add_task",
		Step:     1,
		TaskData: make(map[string]string),
	}

	text := "📝 *Создание новой задачи*\n\n1️⃣ Введите название задачи:"
	keyboard := getBackToMenuKeyboard()
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

// handleNotesCallback обрабатывает показ списка заметок
func (b *Bot) handleNotesCallback(ctx context.Context, chatID, userID int64) {
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
		text := "📝 У вас пока нет заметок\n\nНажмите кнопку ниже, чтобы создать первую заметку!"
		keyboard := tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				{
					tgbotapi.InlineKeyboardButton{Text: "📄 Создать заметку", CallbackData: &[]string{"cmd_add_note"}[0]},
				},
				{
					tgbotapi.InlineKeyboardButton{Text: "🏠 Главное меню", CallbackData: &[]string{"cmd_menu"}[0]},
				},
			},
		}
		b.sendMessageWithKeyboard(chatID, text, keyboard)
		return
	}

	// Конвертируем заметки в формат для клавиатуры
	var noteItems []NoteListItem
	for _, note := range notes {
		noteItems = append(noteItems, NoteListItem{
			ID:         note.ID,
			Title:      note.Title,
			IsFavorite: note.IsFavorite,
		})
	}

	text := fmt.Sprintf("📝 *Ваши заметки* (%d)\n\nВыберите заметку для просмотра:", len(notes))
	keyboard := getNoteListKeyboard(noteItems)
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

// handleAddNoteCallback обрабатывает начало создания заметки
func (b *Bot) handleAddNoteCallback(ctx context.Context, chatID, userID int64) {
	b.userStates[userID] = &UserState{
		Action:   "add_note",
		Step:     1,
		NoteData: make(map[string]string),
	}

	text := "📄 *Создание новой заметки*\n\n1️⃣ Введите заголовок заметки:"
	keyboard := getBackToMenuKeyboard()
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

// handleCompleteTaskCallback обрабатывает завершение задачи
func (b *Bot) handleCompleteTaskCallback(ctx context.Context, query *tgbotapi.CallbackQuery, user *domain.User) {
	chatID := query.Message.Chat.ID
	taskIDStr := strings.TrimPrefix(query.Data, "complete_")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		b.sendMessage(chatID, "❌ Неверный ID задачи")
		return
	}

	task, err := b.taskService.CompleteTask(ctx, taskID, user.ID)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка: %s", err.Error()))
		return
	}

	text := fmt.Sprintf("✅ *Задача выполнена!*\n\n📌 [%d] %s", task.ID, task.Title)
	keyboard := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.InlineKeyboardButton{Text: "📋 К задачам", CallbackData: &[]string{"cmd_tasks"}[0]},
				tgbotapi.InlineKeyboardButton{Text: "🏠 Главное меню", CallbackData: &[]string{"cmd_menu"}[0]},
			},
		},
	}
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

// handleShowTaskCallback обрабатывает показ детальной информации о задаче
func (b *Bot) handleShowTaskCallback(ctx context.Context, query *tgbotapi.CallbackQuery, user *domain.User) {
	chatID := query.Message.Chat.ID
	taskIDStr := strings.TrimPrefix(query.Data, "show_")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		b.sendMessage(chatID, "❌ Неверный ID задачи")
		return
	}

	task, err := b.taskService.GetTaskByID(ctx, taskID, user.ID)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка: %s", err.Error()))
		return
	}

	text := b.taskService.FormatTask(task)
	keyboard := getTaskActionsKeyboard(taskID)
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

// handlePendingCallback обрабатывает показ активных задач
func (b *Bot) handlePendingCallback(ctx context.Context, chatID, userID int64) {
	user, err := b.getUserFromTelegram(ctx, userID)
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка авторизации")
		return
	}

	tasks, err := b.taskService.GetTasksByStatus(ctx, user.ID, domain.TaskStatusPending)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка получения задач: %s", err.Error()))
		return
	}

	if len(tasks) == 0 {
		text := "⏰ Нет активных задач\n\nВсе задачи выполнены! 🎉"
		keyboard := getBackToMenuKeyboard()
		b.sendMessageWithKeyboard(chatID, text, keyboard)
		return
	}

	text := fmt.Sprintf("⏰ *Активные задачи* (%d)\n\n%s", len(tasks), b.taskService.FormatTaskList(tasks))
	keyboard := getBackToMenuKeyboard()
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

// handleCompletedCallback обрабатывает показ выполненных задач
func (b *Bot) handleCompletedCallback(ctx context.Context, chatID, userID int64) {
	user, err := b.getUserFromTelegram(ctx, userID)
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка авторизации")
		return
	}

	tasks, err := b.taskService.GetTasksByStatus(ctx, user.ID, domain.TaskStatusCompleted)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка получения задач: %s", err.Error()))
		return
	}

	if len(tasks) == 0 {
		text := "✅ Нет выполненных задач\n\nПора взяться за дело! 💪"
		keyboard := getBackToMenuKeyboard()
		b.sendMessageWithKeyboard(chatID, text, keyboard)
		return
	}

	text := fmt.Sprintf("✅ *Выполненные задачи* (%d)\n\n%s", len(tasks), b.taskService.FormatTaskList(tasks))
	keyboard := getBackToMenuKeyboard()
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

// handleHelpCallback обрабатывает показ справки
func (b *Bot) handleHelpCallback(ctx context.Context, chatID int64) {
	helpText := `❓ *Справка по командам*

🤖 *Основные функции:*
• Управление задачами с приоритетами
• Создание и организация заметок
• Установка напоминаний
• Поиск по заметкам

📋 *Работа с задачами:*
• Отправьте любой текст для быстрого создания задачи
• Используйте кнопки для управления задачами
• Устанавливайте приоритеты и напоминания

📝 *Работа с заметками:*
• Создавайте текстовые заметки и прикрепляйте файлы
• Организуйте заметки по категориям
• Добавляйте в избранное важные заметки

💡 *Советы:*
• Используйте кнопки для быстрой навигации
• Все ваши данные сохраняются автоматически
• Бот поддерживает различные типы файлов`

	keyboard := getBackToMenuKeyboard()
	b.sendMessageWithKeyboard(chatID, helpText, keyboard)
}

// handleLogoutCallback обрабатывает выход из системы
func (b *Bot) handleLogoutCallback(ctx context.Context, chatID, userID int64) {
	keyboard := getConfirmationKeyboard("logout", 0)
	text := "🚪 *Выход из системы*\n\nВы уверены, что хотите выйти?"
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}
