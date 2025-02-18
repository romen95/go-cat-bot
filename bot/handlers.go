package bot

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"gocarbot/internal"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotHandler struct {
	Bot        *tgbotapi.BotAPI
	UserStates map[int64]string
}

const (
	StateWaitingForPhone = "waiting_for_phone"
	StateNone            = "none"
)

func (h *BotHandler) HandleUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		h.HandleMessage(update.Message)
	}

	if update.CallbackQuery != nil {
		h.HandleCallbackQuery(update.CallbackQuery)
		return
	}
}

func (h *BotHandler) HandleMessage(message *tgbotapi.Message) {
	userID := message.Chat.ID

	// Если пользователь находится в состоянии ожидания ввода телефона, проверяем его ввод
	if h.UserStates[userID] == StateWaitingForPhone {
		h.HandlePhoneNumberInput(message)
		return
	}

	switch message.Text {
	case "/start":
		h.HandleStart(message)
	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Неизвестная команда. Введите /start")
		if _, err := h.Bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки сообщения: %v", err)
		}
	}
}

func (h *BotHandler) HandleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	switch callback.Data {
	case "get_info":
		h.HandleGetInfo(callback)
	case "get_main":
		h.HandleGetMain(callback)
	}
}

func (h *BotHandler) HandleStart(message *tgbotapi.Message) {
	userID := strconv.Itoa(int(message.Chat.ID))

	// Проверяем, есть ли пользователь в базе данных
	_, err := internal.GetUserByID(userID)
	if err != nil {
		if strings.Contains(err.Error(), "пользователь с ID") {
			// Если пользователь не найден, запрашиваем номер телефона
			text := "Добро пожаловать! Пожалуйста, введите ваш номер телефона в формате 79999999999"
			msg := tgbotapi.NewMessage(message.Chat.ID, text)
			if _, err := h.Bot.Send(msg); err != nil {
				log.Printf("Ошибка отправки сообщения: %v", err)
			}

			// Сохраняем состояние, чтобы ожидать номер телефона
			h.UserStates[message.Chat.ID] = StateWaitingForPhone
			return
		} else {
			log.Printf("Ошибка при запросе данных: %v", err)
			return
		}
	}

	text := "Главное меню\n\nДобро пожаловать! Здесь вы можете отслеживать статус ваших заказов"
	buttonInfo := tgbotapi.NewInlineKeyboardButtonData("Показать информацию", "get_info")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(buttonInfo),
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyMarkup = keyboard
	if _, err := h.Bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}

func (h *BotHandler) HandleGetMain(callback *tgbotapi.CallbackQuery) {
	text := "Главное меню\n\nДобро пожаловать! Здесь вы можете отслеживать статус ваших заказов"
	buttonInfo := tgbotapi.NewInlineKeyboardButtonData("Показать информацию", "get_info")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(buttonInfo),
	)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)

	if _, err := h.Bot.Send(editMsg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
		return
	}
}

func (h *BotHandler) HandleGetInfo(callback *tgbotapi.CallbackQuery) {
	chatID := strconv.Itoa(int(callback.Message.Chat.ID))
	user, err := internal.GetUserByID(chatID)
	if err != nil {
		log.Printf("Ошибка поиска пользователя: %v", err)
		return
	}

	text := fmt.Sprintf("Информация\n\nID: %s\nНомер телефона: %s\nАвтомобиль: %s", user.ID, user.PhoneNumber, user.Car)
	buttonMain := tgbotapi.NewInlineKeyboardButtonData("В главное меню", "get_main")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(buttonMain),
	)

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		text,
		keyboard,
	)

	if _, err := h.Bot.Send(editMsg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
		return
	}
}

func (h *BotHandler) HandlePhoneNumberInput(message *tgbotapi.Message) {
	// Проверка формата номера телефона
	phoneRegex := `^7\d{10}$`
	matched, _ := regexp.MatchString(phoneRegex, message.Text)

	if !matched {
		// Номер не соответствует формату
		msg := tgbotapi.NewMessage(message.Chat.ID, "⚠️ Неверный формат! Введите номер в формате: 79999999999")
		if _, err := h.Bot.Send(msg); err != nil {
			log.Printf("Ошибка отправки сообщения: %v", err)
		}
		return
	}

	// Номер корректен — добавляем в базу
	userID := strconv.Itoa(int(message.Chat.ID))
	err := internal.RequestPhoneAndUpdateID(userID, message.Text)
	if err != nil {
		// Проверяем содержимое ошибки
		if strings.Contains(err.Error(), "номер не найден") {
			msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Вашего номера нет в базе. Обратитесь в поддержку.")
			if _, sendErr := h.Bot.Send(msg); sendErr != nil {
				log.Printf("Ошибка отправки сообщения: %v", sendErr)
			}
		} else {
			log.Printf("Ошибка при обновлении телефона: %v", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "🚨 Произошла ошибка. Попробуйте позже.")
			if _, sendErr := h.Bot.Send(msg); sendErr != nil {
				log.Printf("Ошибка отправки сообщения: %v", sendErr)
			}
		}
		return
	}

	text := "✅ Вы успешно вошли в систему!"
	buttonMain := tgbotapi.NewInlineKeyboardButtonData("В главное меню", "get_main")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(buttonMain),
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyMarkup = keyboard
	if _, err := h.Bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
	h.UserStates[message.Chat.ID] = StateNone

}
