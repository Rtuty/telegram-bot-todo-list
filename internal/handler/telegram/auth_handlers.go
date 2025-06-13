package telegram

import (
	"context"
	"errors"
	"strings"

	"todolist/internal/domain"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleStartCommand обрабатывает команду /start
func (b *Bot) handleStartCommand(ctx context.Context, message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID

	args := strings.Fields(message.Text)
	if len(args) < 2 {
		b.sendMessage(chatID, "🔐 Для авторизации отправьте: /start пароль")
		return
	}

	password := args[1]

	// Проверяем, есть ли уже активная сессия
	_, err := b.authService.IsAuthenticated(ctx, userID)
	if err == nil {
		text := "✅ *Вы уже авторизованы!*\n\nВыберите действие в главном меню:"
		keyboard := getMainMenuKeyboard()
		b.sendMessageWithKeyboard(chatID, text, keyboard)
		return
	}

	// Аутентификация через Login
	username := message.From.UserName
	firstName := message.From.FirstName
	lastName := message.From.LastName

	_, err = b.authService.Login(ctx, userID, username, firstName, lastName, password)
	if err != nil {
		b.sendMessage(chatID, "❌ Неверный пароль")
		return
	}

	welcomeMsg := `🎉 *Добро пожаловать в TodoList Bot!*

Вы успешно авторизованы в системе.

💡 *Быстрые советы:*
• Отправьте любой текст для создания задачи
• Используйте кнопки ниже для навигации
• Прикрепите файл для создания заметки

Выберите действие в главном меню:`

	keyboard := getMainMenuKeyboard()
	b.sendMessageWithKeyboard(chatID, welcomeMsg, keyboard)
}

// handleLogoutCommand обрабатывает команду /logout
func (b *Bot) handleLogoutCommand(ctx context.Context, chatID, userID int64) {
	err := b.authService.Logout(ctx, userID)
	if err != nil {
		b.sendMessage(chatID, "❌ Ошибка при выходе")
		return
	}

	// Удаляем состояние пользователя
	delete(b.userStates, userID)

	b.sendMessage(chatID, "👋 Вы вышли из системы. Для повторной авторизации отправьте /start пароль")
}

// getUserFromTelegram получает пользователя по Telegram ID
func (b *Bot) getUserFromTelegram(ctx context.Context, telegramID int64) (*domain.User, error) {
	user, err := b.authService.IsAuthenticated(ctx, telegramID)
	if err != nil {
		return nil, errors.New("пользователь не авторизован")
	}

	return user, nil
}
