package telegram

import (
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// getMainMenuKeyboard возвращает главное меню бота
func getMainMenuKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.InlineKeyboardButton{Text: "📋 Мои задачи", CallbackData: &[]string{"cmd_tasks"}[0]},
				tgbotapi.InlineKeyboardButton{Text: "➕ Добавить задачу", CallbackData: &[]string{"cmd_add_task"}[0]},
			},
			{
				tgbotapi.InlineKeyboardButton{Text: "📝 Мои заметки", CallbackData: &[]string{"cmd_notes"}[0]},
				tgbotapi.InlineKeyboardButton{Text: "📄 Добавить заметку", CallbackData: &[]string{"cmd_add_note"}[0]},
			},
			{
				tgbotapi.InlineKeyboardButton{Text: "⏰ Активные задачи", CallbackData: &[]string{"cmd_pending"}[0]},
				tgbotapi.InlineKeyboardButton{Text: "✅ Выполненные", CallbackData: &[]string{"cmd_completed"}[0]},
			},
			{
				tgbotapi.InlineKeyboardButton{Text: "🔍 Поиск заметок", CallbackData: &[]string{"cmd_search"}[0]},
				tgbotapi.InlineKeyboardButton{Text: "⭐ Избранные", CallbackData: &[]string{"cmd_favorites"}[0]},
			},
			{
				tgbotapi.InlineKeyboardButton{Text: "❓ Справка", CallbackData: &[]string{"cmd_help"}[0]},
				tgbotapi.InlineKeyboardButton{Text: "🚪 Выйти", CallbackData: &[]string{"cmd_logout"}[0]},
			},
		},
	}
}

// getTaskActionsKeyboard возвращает клавиатуру для действий с задачей
func getTaskActionsKeyboard(taskID int) tgbotapi.InlineKeyboardMarkup {
	taskIDStr := strconv.Itoa(taskID)
	return tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.InlineKeyboardButton{Text: "✅ Выполнить", CallbackData: &[]string{"complete_" + taskIDStr}[0]},
				tgbotapi.InlineKeyboardButton{Text: "👀 Подробнее", CallbackData: &[]string{"show_" + taskIDStr}[0]},
			},
			{
				tgbotapi.InlineKeyboardButton{Text: "⏰ Напоминание", CallbackData: &[]string{"notify_" + taskIDStr}[0]},
				tgbotapi.InlineKeyboardButton{Text: "🗑️ Удалить", CallbackData: &[]string{"delete_" + taskIDStr}[0]},
			},
			{
				tgbotapi.InlineKeyboardButton{Text: "🔙 Назад к задачам", CallbackData: &[]string{"cmd_tasks"}[0]},
			},
		},
	}
}

// getPriorityKeyboard возвращает клавиатуру для выбора приоритета задачи
func getPriorityKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.InlineKeyboardButton{Text: "🔴 Высокий", CallbackData: &[]string{"priority_high"}[0]},
				tgbotapi.InlineKeyboardButton{Text: "🟡 Средний", CallbackData: &[]string{"priority_medium"}[0]},
				tgbotapi.InlineKeyboardButton{Text: "🟢 Низкий", CallbackData: &[]string{"priority_low"}[0]},
			},
		},
	}
}

// getCategoryKeyboard возвращает клавиатуру для выбора категории заметки
func getCategoryKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.InlineKeyboardButton{Text: "🗂️ Общее", CallbackData: &[]string{"category_general"}[0]},
				tgbotapi.InlineKeyboardButton{Text: "💼 Работа", CallbackData: &[]string{"category_work"}[0]},
			},
			{
				tgbotapi.InlineKeyboardButton{Text: "📚 Учеба", CallbackData: &[]string{"category_study"}[0]},
				tgbotapi.InlineKeyboardButton{Text: "👤 Личное", CallbackData: &[]string{"category_personal"}[0]},
			},
			{
				tgbotapi.InlineKeyboardButton{Text: "🔗 Ресурсы", CallbackData: &[]string{"category_resources"}[0]},
				tgbotapi.InlineKeyboardButton{Text: "💡 Идеи", CallbackData: &[]string{"category_ideas"}[0]},
			},
		},
	}
}

// getNoteActionsKeyboard возвращает клавиатуру для действий с заметкой
func getNoteActionsKeyboard(noteID int, isFavorite bool) tgbotapi.InlineKeyboardMarkup {
	noteIDStr := strconv.Itoa(noteID)
	favoriteText := "⭐ В избранное"
	favoriteAction := "favorite_add_" + noteIDStr

	if isFavorite {
		favoriteText = "✨ Убрать из избранного"
		favoriteAction = "favorite_remove_" + noteIDStr
	}

	return tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.InlineKeyboardButton{Text: favoriteText, CallbackData: &favoriteAction},
				tgbotapi.InlineKeyboardButton{Text: "📝 Редактировать", CallbackData: &[]string{"edit_note_" + noteIDStr}[0]},
			},
			{
				tgbotapi.InlineKeyboardButton{Text: "🗑️ Удалить", CallbackData: &[]string{"delete_note_" + noteIDStr}[0]},
				tgbotapi.InlineKeyboardButton{Text: "🔙 К заметкам", CallbackData: &[]string{"cmd_notes"}[0]},
			},
		},
	}
}

// getTaskListKeyboard возвращает клавиатуру для списка задач с кнопками действий
func getTaskListKeyboard(tasks []TaskListItem) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Добавляем кнопки для каждой задачи (максимум 5)
	for i, task := range tasks {
		if i >= 5 {
			break
		}

		taskIDStr := strconv.Itoa(task.ID)
		completeBtn := tgbotapi.InlineKeyboardButton{
			Text:         "✅",
			CallbackData: &[]string{"complete_" + taskIDStr}[0],
		}
		showBtn := tgbotapi.InlineKeyboardButton{
			Text:         fmt.Sprintf("👀 [%d] %s", task.ID, truncateString(task.Title, 20)),
			CallbackData: &[]string{"show_" + taskIDStr}[0],
		}

		rows = append(rows, []tgbotapi.InlineKeyboardButton{completeBtn, showBtn})
	}

	// Добавляем кнопки управления
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.InlineKeyboardButton{Text: "➕ Добавить задачу", CallbackData: &[]string{"cmd_add_task"}[0]},
		tgbotapi.InlineKeyboardButton{Text: "🔄 Обновить", CallbackData: &[]string{"cmd_tasks"}[0]},
	})

	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.InlineKeyboardButton{Text: "🏠 Главное меню", CallbackData: &[]string{"cmd_menu"}[0]},
	})

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// getNoteListKeyboard возвращает клавиатуру для списка заметок
func getNoteListKeyboard(notes []NoteListItem) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Добавляем кнопки для каждой заметки (максимум 5)
	for i, note := range notes {
		if i >= 5 {
			break
		}

		noteIDStr := strconv.Itoa(note.ID)
		favoriteIcon := ""
		if note.IsFavorite {
			favoriteIcon = "⭐"
		}

		showBtn := tgbotapi.InlineKeyboardButton{
			Text:         fmt.Sprintf("%s📝 [%d] %s", favoriteIcon, note.ID, truncateString(note.Title, 18)),
			CallbackData: &[]string{"show_note_" + noteIDStr}[0],
		}

		rows = append(rows, []tgbotapi.InlineKeyboardButton{showBtn})
	}

	// Добавляем кнопки управления
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.InlineKeyboardButton{Text: "📄 Добавить заметку", CallbackData: &[]string{"cmd_add_note"}[0]},
		tgbotapi.InlineKeyboardButton{Text: "🔄 Обновить", CallbackData: &[]string{"cmd_notes"}[0]},
	})

	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.InlineKeyboardButton{Text: "🏠 Главное меню", CallbackData: &[]string{"cmd_menu"}[0]},
	})

	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// getConfirmationKeyboard возвращает клавиатуру подтверждения действия
func getConfirmationKeyboard(action string, itemID int) tgbotapi.InlineKeyboardMarkup {
	itemIDStr := strconv.Itoa(itemID)
	return tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.InlineKeyboardButton{Text: "✅ Да", CallbackData: &[]string{"confirm_" + action + "_" + itemIDStr}[0]},
				tgbotapi.InlineKeyboardButton{Text: "❌ Отмена", CallbackData: &[]string{"cancel_" + action}[0]},
			},
		},
	}
}

// getBackToMenuKeyboard возвращает кнопку возврата в главное меню
func getBackToMenuKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			{
				tgbotapi.InlineKeyboardButton{Text: "🏠 Главное меню", CallbackData: &[]string{"cmd_menu"}[0]},
			},
		},
	}
}

// TaskListItem представляет элемент списка задач для клавиатуры
type TaskListItem struct {
	ID    int
	Title string
}

// NoteListItem представляет элемент списка заметок для клавиатуры
type NoteListItem struct {
	ID         int
	Title      string
	IsFavorite bool
}

// truncateString обрезает строку до указанной длины
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
