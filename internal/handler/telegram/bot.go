package telegram

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"

	"todolist/config"
	"todolist/internal/domain"
	"todolist/internal/usecase"
)

// Bot представляет телеграм бота
type Bot struct {
	api                 *tgbotapi.BotAPI
	authService         *usecase.AuthService
	taskService         *usecase.TaskService
	noteService         *usecase.NoteService
	notificationService *usecase.NotificationService
	config              *config.Config
	logger              *zap.Logger
	userStates          map[int64]*UserState
}

// UserState хранит состояние пользователя для многошаговых операций
type UserState struct {
	Action      string
	Step        int
	TaskID      int
	NoteID      int
	TaskData    map[string]string
	NoteData    map[string]string
	LastCommand string
}

// NewBot создает новый экземпляр бота
func NewBot(
	api *tgbotapi.BotAPI,
	authService *usecase.AuthService,
	taskService *usecase.TaskService,
	noteService *usecase.NoteService,
	notificationService *usecase.NotificationService,
	config *config.Config,
	logger *zap.Logger,
) *Bot {
	return &Bot{
		api:                 api,
		authService:         authService,
		taskService:         taskService,
		noteService:         noteService,
		notificationService: notificationService,
		config:              config,
		logger:              logger,
		userStates:          make(map[int64]*UserState),
	}
}

// Start запускает бота
func (b *Bot) Start(ctx context.Context) error {
	b.logger.Info("bot starting...")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			b.logger.Info("bot stopping...")
			return ctx.Err()
		case update := <-updates:
			go b.handleUpdate(ctx, update)
		}
	}
}

// handleUpdate обрабатывает обновления от Telegram
func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			b.logger.Error("panic in update handler", zap.Any("panic", r))
		}
	}()

	if update.Message != nil {
		b.handleMessage(ctx, update.Message)
	} else if update.CallbackQuery != nil {
		b.handleCallbackQuery(ctx, update.CallbackQuery)
	}
}

// handleMessage обрабатывает входящие сообщения
func (b *Bot) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	userID := message.From.ID
	chatID := message.Chat.ID

	b.logger.Info("received message",
		zap.Int64("user_id", userID),
		zap.Int64("chat_id", chatID),
		zap.String("text", message.Text))

	// Проверяем авторизацию (кроме команды /start)
	if !strings.HasPrefix(message.Text, "/start") {
		user, err := b.authService.IsAuthenticated(ctx, userID)
		if err != nil {
			b.sendMessage(chatID, "🔐 Для использования бота необходимо авторизоваться. Используйте команду /start")
			return
		}
		userID = user.ID // Используем внутренний ID пользователя
	}

	// Обработка состояний пользователя
	if state, exists := b.userStates[userID]; exists && state.Action != "" {
		b.handleUserState(ctx, message, state)
		return
	}

	// Обработка команд
	if message.IsCommand() {
		b.handleCommand(ctx, message)
	} else if message.Document != nil || len(message.Photo) > 0 || message.Video != nil ||
		message.Audio != nil || message.Voice != nil {
		// Если это файл, создаем заметку из файла
		b.handleCreateNoteFromFile(ctx, message)
	} else {
		// Если это не команда, создаем задачу из текста
		b.handleCreateTaskFromText(ctx, message)
	}
}

// handleCommand обрабатывает команды бота
func (b *Bot) handleCommand(ctx context.Context, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID

	switch message.Command() {
	case "start":
		b.handleStartCommand(ctx, message)
	case "help":
		b.handleHelpCommand(chatID)
	case "tasks", "list":
		b.handleListTasksCommand(ctx, chatID, userID)
	case "add", "new":
		b.handleAddTaskCommand(ctx, message)
	case "complete", "done":
		b.handleCompleteTaskCommand(ctx, message)
	case "delete", "del":
		b.handleDeleteTaskCommand(ctx, message)
	case "show", "get":
		b.handleShowTaskCommand(ctx, message)
	case "pending":
		b.handlePendingTasksCommand(ctx, chatID, userID)
	case "completed":
		b.handleCompletedTasksCommand(ctx, chatID, userID)
	case "notify":
		b.handleSetNotificationCommand(ctx, message)
	case "notes":
		b.handleListNotesCommand(ctx, chatID, userID)
	case "note":
		b.handleAddNoteCommand(ctx, message)
	case "nshow":
		b.handleShowNoteCommand(ctx, message)
	case "ndelete":
		b.handleDeleteNoteCommand(ctx, message)
	case "favorites":
		b.handleFavoriteNotesCommand(ctx, chatID, userID)
	case "favorite":
		b.handleToggleFavoriteCommand(ctx, message)
	case "search":
		b.handleSearchNotesCommand(ctx, message)
	case "links":
		b.handleLinkNotesCommand(ctx, chatID, userID)
	case "files":
		b.handleFileNotesCommand(ctx, chatID, userID)
	case "logout":
		b.handleLogoutCommand(ctx, chatID, userID)
	default:
		b.sendMessage(chatID, "❓ Неизвестная команда. Используйте /help для просмотра доступных команд.")
	}
}

// handleStartCommand обрабатывает команду /start
func (b *Bot) handleStartCommand(ctx context.Context, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID

	// Проверяем, авторизован ли пользователь
	if _, err := b.authService.IsAuthenticated(ctx, userID); err == nil {
		b.sendMessage(chatID, "👋 Вы уже авторизованы! Используйте /help для просмотра команд.")
		return
	}

	args := strings.Fields(message.Text)
	if len(args) < 2 {
		b.sendMessage(chatID, "🔐 Для авторизации введите: /start ваш_пароль")
		return
	}

	password := strings.Join(args[1:], " ")
	username := ""
	if message.From.UserName != "" {
		username = "@" + message.From.UserName
	}

	user, err := b.authService.Login(ctx, userID, username,
		message.From.FirstName, message.From.LastName, password)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка авторизации: %s", err.Error()))
		return
	}

	welcomeMsg := fmt.Sprintf("🎉 Добро пожаловать, %s!\n\n", user.FirstName)
	welcomeMsg += "Вы успешно авторизованы в Todo Bot.\n\n"
	welcomeMsg += "📝 Основные команды:\n"
	welcomeMsg += "• Отправьте любой текст для создания задачи\n"
	welcomeMsg += "• /tasks - список всех задач\n"
	welcomeMsg += "• /help - справка по командам\n\n"
	welcomeMsg += "Начните управлять своими задачами уже сейчас! 🚀"

	b.sendMessage(chatID, welcomeMsg)
}

// handleHelpCommand обрабатывает команду /help
func (b *Bot) handleHelpCommand(chatID int64) {
	helpText := `📖 *Справка по командам Todo Bot*

🔐 *Авторизация:*
/start пароль - авторизация в системе

📝 *Работа с задачами:*
/tasks, /list - показать все задачи
/pending - показать невыполненные задачи
/completed - показать выполненные задачи
/add название - создать новую задачу
/complete ID - отметить задачу как выполненную
/delete ID - удалить задачу
/show ID - показать подробную информацию о задаче

📚 *Работа с заметками:*
/notes - показать все заметки
/note заголовок - создать новую заметку
/nshow ID - показать заметку
/ndelete ID - удалить заметку
/favorites - показать избранные заметки
/favorite ID - добавить/убрать из избранного
/search запрос - поиск заметок
/links - показать все ссылки
/files - показать все файлы

⏰ *Уведомления:*
/notify ID время - установить напоминание
   Примеры времени:
   • 15:30 - сегодня в 15:30
   • завтра 10:00 - завтра в 10:00
   • 25.12 14:00 - 25 декабря в 14:00

🔧 *Прочее:*
/help - показать эту справку
/logout - выйти из системы

💡 *Быстрое создание:*
Просто отправьте текст - он станет новой задачей!
Отправьте документ/изображение - станет заметкой!

📋 *Приоритеты задач:*
🔴 высокий | 🟡 средний | 🟢 низкий`

	msg := tgbotapi.NewMessage(chatID, helpText)
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}

// sendMessage отправляет сообщение пользователю
func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := b.api.Send(msg); err != nil {
		b.logger.Error("failed to send message", zap.Error(err))
	}
}

// sendMessageWithKeyboard отправляет сообщение с клавиатурой
func (b *Bot) sendMessageWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	if _, err := b.api.Send(msg); err != nil {
		b.logger.Error("failed to send message with keyboard", zap.Error(err))
	}
}

// getUserFromTelegram получает пользователя из контекста телеграма
func (b *Bot) getUserFromTelegram(ctx context.Context, telegramUserID int64) (*domain.User, error) {
	return b.authService.IsAuthenticated(ctx, telegramUserID)
}
