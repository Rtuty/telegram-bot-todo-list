package telegram

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"todolist/internal/domain"
)

// handleUserState обрабатывает состояния пользователя для многошаговых операций
func (b *Bot) handleUserState(ctx context.Context, message *tgbotapi.Message, state *UserState) {
	chatID := message.Chat.ID
	userID := message.From.ID

	user, err := b.getUserFromTelegram(ctx, userID)
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка авторизации")
		return
	}

	switch state.Action {
	case "add_task":
		b.handleAddTaskState(ctx, message, user, state)
	case "add_note":
		b.handleAddNoteState(ctx, message, user, state)
	case "set_notification":
		b.handleSetNotificationState(ctx, message, user, state)
	default:
		delete(b.userStates, userID)
		b.sendMessage(chatID, "❌ Неизвестное состояние. Попробуйте еще раз.")
	}
}

// handleCreateTaskFromText создает задачу из произвольного текста
func (b *Bot) handleCreateTaskFromText(ctx context.Context, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID

	user, err := b.getUserFromTelegram(ctx, userID)
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка авторизации")
		return
	}

	title := strings.TrimSpace(message.Text)
	if title == "" {
		b.sendMessage(chatID, "❌ Название задачи не может быть пустым")
		return
	}

	task, err := b.taskService.CreateTask(ctx, user.ID, title, "", domain.TaskPriorityMedium)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка создания задачи: %s", err.Error()))
		return
	}

	b.sendMessage(chatID, fmt.Sprintf("✅ Задача [%d] создана!\n📌 %s", task.ID, task.Title))
}

// handleListTasksCommand обрабатывает команду /tasks
func (b *Bot) handleListTasksCommand(ctx context.Context, chatID, userID int64) {
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

	b.sendMessage(chatID, b.taskService.FormatTaskList(tasks))
}

// handleAddTaskCommand обрабатывает команду /add
func (b *Bot) handleAddTaskCommand(ctx context.Context, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID

	user, err := b.getUserFromTelegram(ctx, userID)
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка авторизации")
		return
	}

	args := strings.Fields(message.Text)
	if len(args) < 2 {
		// Запускаем интерактивное создание задачи
		b.userStates[userID] = &UserState{
			Action:   "add_task",
			Step:     1,
			TaskData: make(map[string]string),
		}
		b.sendMessage(chatID, "📝 Создание новой задачи\n\n1️⃣ Введите название задачи:")
		return
	}

	title := strings.Join(args[1:], " ")
	task, err := b.taskService.CreateTask(ctx, user.ID, title, "", domain.TaskPriorityMedium)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка создания задачи: %s", err.Error()))
		return
	}

	b.sendMessage(chatID, fmt.Sprintf("✅ Задача [%d] создана!\n📌 %s", task.ID, task.Title))
}

// handleCompleteTaskCommand обрабатывает команду /complete
func (b *Bot) handleCompleteTaskCommand(ctx context.Context, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID

	user, err := b.getUserFromTelegram(ctx, userID)
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка авторизации")
		return
	}

	args := strings.Fields(message.Text)
	if len(args) < 2 {
		b.sendMessage(chatID, "❌ Укажите ID задачи: /complete 123")
		return
	}

	taskID, err := strconv.Atoi(args[1])
	if err != nil {
		b.sendMessage(chatID, "❌ Неверный ID задачи")
		return
	}

	task, err := b.taskService.CompleteTask(ctx, taskID, user.ID)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка: %s", err.Error()))
		return
	}

	b.sendMessage(chatID, fmt.Sprintf("✅ Задача [%d] выполнена!\n📌 %s", task.ID, task.Title))
}

// handleDeleteTaskCommand обрабатывает команду /delete
func (b *Bot) handleDeleteTaskCommand(ctx context.Context, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID

	user, err := b.getUserFromTelegram(ctx, userID)
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка авторизации")
		return
	}

	args := strings.Fields(message.Text)
	if len(args) < 2 {
		b.sendMessage(chatID, "❌ Укажите ID задачи: /delete 123")
		return
	}

	taskID, err := strconv.Atoi(args[1])
	if err != nil {
		b.sendMessage(chatID, "❌ Неверный ID задачи")
		return
	}

	err = b.taskService.DeleteTask(ctx, taskID, user.ID)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка: %s", err.Error()))
		return
	}

	b.sendMessage(chatID, fmt.Sprintf("🗑️ Задача [%d] удалена!", taskID))
}

// handleShowTaskCommand обрабатывает команду /show
func (b *Bot) handleShowTaskCommand(ctx context.Context, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID

	user, err := b.getUserFromTelegram(ctx, userID)
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка авторизации")
		return
	}

	args := strings.Fields(message.Text)
	if len(args) < 2 {
		b.sendMessage(chatID, "❌ Укажите ID задачи: /show 123")
		return
	}

	taskID, err := strconv.Atoi(args[1])
	if err != nil {
		b.sendMessage(chatID, "❌ Неверный ID задачи")
		return
	}

	task, err := b.taskService.GetTaskByID(ctx, taskID, user.ID)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка: %s", err.Error()))
		return
	}

	b.sendMessage(chatID, b.taskService.FormatTask(task))
}

// handlePendingTasksCommand обрабатывает команду /pending
func (b *Bot) handlePendingTasksCommand(ctx context.Context, chatID, userID int64) {
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
		b.sendMessage(chatID, "🎉 У вас нет невыполненных задач!")
		return
	}

	message := "⏳ Невыполненные задачи:\n\n"
	message += b.taskService.FormatTaskList(tasks)
	b.sendMessage(chatID, message)
}

// handleCompletedTasksCommand обрабатывает команду /completed
func (b *Bot) handleCompletedTasksCommand(ctx context.Context, chatID, userID int64) {
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
		b.sendMessage(chatID, "📝 У вас нет выполненных задач.")
		return
	}

	message := "✅ Выполненные задачи:\n\n"
	message += b.taskService.FormatTaskList(tasks)
	b.sendMessage(chatID, message)
}

// handleSetNotificationCommand обрабатывает команду /notify
func (b *Bot) handleSetNotificationCommand(ctx context.Context, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID

	user, err := b.getUserFromTelegram(ctx, userID)
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка авторизации")
		return
	}

	args := strings.Fields(message.Text)
	if len(args) < 3 {
		// Запускаем интерактивную настройку уведомления
		b.userStates[userID] = &UserState{
			Action:   "set_notification",
			Step:     1,
			TaskData: make(map[string]string),
		}
		b.sendMessage(chatID, "⏰ Настройка уведомления\n\n1️⃣ Введите ID задачи:")
		return
	}

	taskID, err := strconv.Atoi(args[1])
	if err != nil {
		b.sendMessage(chatID, "❌ Неверный ID задачи")
		return
	}

	timeStr := strings.Join(args[2:], " ")
	notifyTime, err := b.parseTime(timeStr)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Неверный формат времени: %s", err.Error()))
		return
	}

	task, err := b.taskService.SetTaskNotification(ctx, taskID, user.ID, notifyTime)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка: %s", err.Error()))
		return
	}

	b.sendMessage(chatID, fmt.Sprintf("⏰ Уведомление установлено!\n📌 Задача [%d]: %s\n🕐 Время: %s",
		task.ID, task.Title, notifyTime.Format("02.01.2006 15:04")))
}

// parseTime парсит время из различных форматов
func (b *Bot) parseTime(timeStr string) (time.Time, error) {
	now := time.Now()
	timeStr = strings.ToLower(strings.TrimSpace(timeStr))

	// Регулярные выражения для разных форматов времени
	patterns := []struct {
		regex  *regexp.Regexp
		parser func([]string) (time.Time, error)
	}{
		// Сегодня HH:MM
		{
			regexp.MustCompile(`^(\d{1,2}):(\d{2})$`),
			func(matches []string) (time.Time, error) {
				hour, _ := strconv.Atoi(matches[1])
				minute, _ := strconv.Atoi(matches[2])
				return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location()), nil
			},
		},
		// Завтра HH:MM
		{
			regexp.MustCompile(`^завтра (\d{1,2}):(\d{2})$`),
			func(matches []string) (time.Time, error) {
				hour, _ := strconv.Atoi(matches[1])
				minute, _ := strconv.Atoi(matches[2])
				tomorrow := now.AddDate(0, 0, 1)
				return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), hour, minute, 0, 0, now.Location()), nil
			},
		},
		// DD.MM HH:MM
		{
			regexp.MustCompile(`^(\d{1,2})\.(\d{1,2}) (\d{1,2}):(\d{2})$`),
			func(matches []string) (time.Time, error) {
				day, _ := strconv.Atoi(matches[1])
				month, _ := strconv.Atoi(matches[2])
				hour, _ := strconv.Atoi(matches[3])
				minute, _ := strconv.Atoi(matches[4])
				year := now.Year()
				return time.Date(year, time.Month(month), day, hour, minute, 0, 0, now.Location()), nil
			},
		},
	}

	for _, pattern := range patterns {
		if matches := pattern.regex.FindStringSubmatch(timeStr); matches != nil {
			return pattern.parser(matches)
		}
	}

	return time.Time{}, fmt.Errorf("неподдерживаемый формат времени")
}

// handleAddTaskState обрабатывает состояние создания задачи
func (b *Bot) handleAddTaskState(ctx context.Context, message *tgbotapi.Message, user *domain.User, state *UserState) {
	chatID := message.Chat.ID

	switch state.Step {
	case 1: // Название задачи
		state.TaskData["title"] = message.Text
		state.Step = 2
		b.sendMessage(chatID, "2️⃣ Введите описание задачи (или отправьте \"-\" чтобы пропустить):")

	case 2: // Описание задачи
		description := message.Text
		if description == "-" {
			description = ""
		}
		state.TaskData["description"] = description
		state.Step = 3

		keyboard := getPriorityKeyboard()
		b.sendMessageWithKeyboard(chatID, "3️⃣ Выберите приоритет задачи:", keyboard)

	default:
		// Завершаем создание задачи
		priority := domain.TaskPriorityMedium
		if message.Text == "high" {
			priority = domain.TaskPriorityHigh
		} else if message.Text == "low" {
			priority = domain.TaskPriorityLow
		}

		task, err := b.taskService.CreateTask(ctx, user.ID,
			state.TaskData["title"],
			state.TaskData["description"],
			priority)

		delete(b.userStates, user.TelegramID)

		if err != nil {
			b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка создания задачи: %s", err.Error()))
			return
		}

		b.sendMessage(chatID, fmt.Sprintf("✅ Задача [%d] создана!\n%s", task.ID, b.taskService.FormatTask(task)))
	}
}

// handleSetNotificationState обрабатывает состояние установки уведомления
func (b *Bot) handleSetNotificationState(ctx context.Context, message *tgbotapi.Message, user *domain.User, state *UserState) {
	chatID := message.Chat.ID

	switch state.Step {
	case 1: // ID задачи
		taskID, err := strconv.Atoi(message.Text)
		if err != nil {
			b.sendMessage(chatID, "❌ Неверный ID задачи. Попробуйте еще раз:")
			return
		}
		state.TaskID = taskID
		state.Step = 2
		b.sendMessage(chatID, "2️⃣ Введите время уведомления:\n\nПримеры:\n• 15:30 - сегодня в 15:30\n• завтра 10:00\n• 25.12 14:00")

	case 2: // Время уведомления
		notifyTime, err := b.parseTime(message.Text)
		if err != nil {
			b.sendMessage(chatID, fmt.Sprintf("❌ Неверный формат времени: %s\nПопробуйте еще раз:", err.Error()))
			return
		}

		task, err := b.taskService.SetTaskNotification(ctx, state.TaskID, user.ID, notifyTime)
		delete(b.userStates, user.TelegramID)

		if err != nil {
			b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка: %s", err.Error()))
			return
		}

		b.sendMessage(chatID, fmt.Sprintf("⏰ Уведомление установлено!\n📌 Задача [%d]: %s\n🕐 Время: %s",
			task.ID, task.Title, notifyTime.Format("02.01.2006 15:04")))
	}
}

// handleAddNoteState обрабатывает состояние создания заметки
func (b *Bot) handleAddNoteState(ctx context.Context, message *tgbotapi.Message, user *domain.User, state *UserState) {
	chatID := message.Chat.ID

	switch state.Step {
	case 1: // Заголовок заметки
		state.NoteData["title"] = message.Text
		state.Step = 2
		b.sendMessage(chatID, "2️⃣ Введите содержимое заметки (или отправьте \"-\" чтобы пропустить):")

	case 2: // Содержимое заметки
		content := message.Text
		if content == "-" {
			content = ""
		}
		state.NoteData["content"] = content
		state.Step = 3

		keyboard := getCategoryKeyboard()
		b.sendMessageWithKeyboard(chatID, "3️⃣ Выберите категорию заметки:", keyboard)

	case 3: // Теги
		state.Step = 4
		b.sendMessage(chatID, "4️⃣ Введите теги через запятую (или отправьте \"-\" чтобы пропустить):")

	case 4: // Завершение создания заметки
		tags := message.Text
		if tags == "-" {
			tags = ""
		}

		category := state.NoteData["category"]
		if category == "" {
			category = "general"
		}

		note, err := b.noteService.CreateNote(ctx, user.ID,
			state.NoteData["title"],
			state.NoteData["content"],
			category,
			tags)

		delete(b.userStates, user.TelegramID)

		if err != nil {
			b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка создания заметки: %s", err.Error()))
			return
		}

		response := fmt.Sprintf("✅ Заметка [%d] создана!\n\n%s", note.ID, b.noteService.FormatNoteForDisplay(note))
		msg := tgbotapi.NewMessage(chatID, response)
		msg.ParseMode = "Markdown"
		b.api.Send(msg)
	}
}
