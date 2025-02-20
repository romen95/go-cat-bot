package bot

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"gocarbot/internal"
	"gocarbot/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotHandler struct {
	Bot           *tgbotapi.BotAPI
	UserStates    map[int64]string
	lastMessageID int
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
	switch {
	case callback.Data == "get_main":
		h.HandleGetMain(callback)
	case callback.Data == "get_orders":
		h.HandleGetOrders(callback)
	case strings.HasPrefix(callback.Data, "order_"):
		h.HandleOrderInfo(callback)
	}
}

func (h *BotHandler) HandleStart(message *tgbotapi.Message) {
	userID := strconv.Itoa(int(message.Chat.ID))

	_, err := internal.GetUserByID(userID)
	if err != nil {
		if strings.Contains(err.Error(), "пользователь с ID") {
			// Создаём кнопку "Поделиться контактом"
			text := "👋🏻 Здравствуйте! Пожалуйста, нажмите кнопку ниже, чтобы отправить ваш номер телефона."

			buttonRequestContact := tgbotapi.NewKeyboardButtonContact("📲 Поделиться контактом")
			keyboard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(buttonRequestContact),
			)

			msg := tgbotapi.NewMessage(message.Chat.ID, text)
			msg.ReplyMarkup = keyboard

			msgSend, err := h.Bot.Send(msg)
			if err != nil {
				log.Printf("Ошибка отправки сообщения: %v", err)
			}
			h.lastMessageID = msgSend.MessageID

			// Устанавливаем состояние ожидания контакта
			h.UserStates[message.Chat.ID] = StateWaitingForPhone
			return
		} else {
			log.Printf("Ошибка при запросе данных: %v", err)
			return
		}
	}
	text := "🏠 Главное меню\n\nЗдесь вы можете получить информацию о статусе выполнения ваших заказов, а также связаться с нами"

	buttonOrders := tgbotapi.NewInlineKeyboardButtonData("📦 Мои заказы", "get_orders")
	buttonSupport := tgbotapi.NewInlineKeyboardButtonURL("🆘 Связаться с нами", "https://yandex.ru")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(buttonOrders),
		tgbotapi.NewInlineKeyboardRow(buttonSupport),
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyMarkup = keyboard
	if _, err := h.Bot.Send(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}

func (h *BotHandler) HandleGetMain(callback *tgbotapi.CallbackQuery) {
	text := "🏠 Главное меню\n\nЗдесь вы можете получить информацию о статусе выполнения ваших заказов, а также связаться с нами"

	buttonOrders := tgbotapi.NewInlineKeyboardButtonData("📦 Мои заказы", "get_orders")
	buttonSupport := tgbotapi.NewInlineKeyboardButtonURL("🆘 Связаться с нами", "https://yandex.ru")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(buttonOrders),
		tgbotapi.NewInlineKeyboardRow(buttonSupport),
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

func (h *BotHandler) HandleGetOrders(callback *tgbotapi.CallbackQuery) {
	chatID := strconv.Itoa(int(callback.Message.Chat.ID))
	user, err := internal.GetUserByID(chatID)
	if err != nil {
		log.Printf("Ошибка поиска пользователя: %v", err)
		return
	}

	text := "📦 Мои заказы\n\nВыберите заказ для просмотра информации 👇🏻"

	// Создаем кнопки только для непустых заказов
	var buttons [][]tgbotapi.InlineKeyboardButton

	// Функция для формирования текста кнопки
	getOrderButtonText := func(order models.Order, orderNumber string) string {
		if order.OrderNumber != "" {
			return fmt.Sprintf("🚗 Заказ №%s", order.OrderNumber)
		}
		return fmt.Sprintf("🚗 Заказ №%s", orderNumber)
	}

	if !internal.IsOrderEmpty(user.Order1) {
		buttonText := getOrderButtonText(user.Order1, "1")
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(buttonText, "order_1")})
	}
	if !internal.IsOrderEmpty(user.Order2) {
		buttonText := getOrderButtonText(user.Order2, "2")
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(buttonText, "order_2")})
	}

	// Если нет ни одного заказа, показываем сообщение
	if len(buttons) == 0 {
		text = "🏠 Главное меню\n\nУ вас пока нет активных заказов"
		editMsg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			text,
		)
		if _, err := h.Bot.Send(editMsg); err != nil {
			log.Printf("Ошибка отправки сообщения: %v", err)
		}
		return
	}

	buttonMain := tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "get_main")
	buttons = append(buttons, []tgbotapi.InlineKeyboardButton{buttonMain})

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

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

func (h *BotHandler) HandleOrderInfo(callback *tgbotapi.CallbackQuery) {
	chatID := strconv.Itoa(int(callback.Message.Chat.ID))
	user, err := internal.GetUserByID(chatID)
	if err != nil {
		log.Printf("Ошибка поиска пользователя: %v", err)
		return
	}

	var order models.Order
	switch callback.Data {
	case "order_1":
		order = user.Order1
	case "order_2":
		order = user.Order2
	default:
		log.Printf("Неизвестный заказ: %s", callback.Data)
		return
	}

	text := fmt.Sprintf(
		"ℹ️ Информация о заказе №%s\n\n"+
			"📅 Дата заказа:\n%s\n"+
			"🚛 Ориентировочная дата прибытия:\n%s\n"+
			"🚗 Марка и модель:\n%s\n"+
			"🌍 Страна отправления:\n%s\n"+
			"📍 Текущее местоположение:\n%s\n"+
			"🏁 Пункт назначения:\n%s",
		order.OrderNumber,
		order.OrderDate,
		order.EstimatedArrival,
		order.CarModel,
		order.CountryOfOrigin,
		order.CurrentLocation,
		order.Destination,
	)

	buttonOrders := tgbotapi.NewInlineKeyboardButtonData("◀️ Назад", "get_orders")
	buttonMain := tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "get_main")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(buttonOrders),
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
	if message.Contact == nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "⚠️ Пожалуйста, используйте кнопку для отправки номера.")
		deleteMsg := tgbotapi.NewDeleteMessage(message.Chat.ID, h.lastMessageID)
		h.Bot.Send(deleteMsg)
		msgSend, err := h.Bot.Send(msg)
		if err != nil {
			log.Printf("Ошибка отправки сообщения: %v", err)
		}
		h.lastMessageID = msgSend.MessageID
		return
	}

	// Удаляем сообщение с контактом
	deleteContactMsg := tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID)
	h.Bot.Send(deleteContactMsg)

	phoneNumber := message.Contact.PhoneNumber
	phoneNumber = strings.TrimPrefix(phoneNumber, "+")

	phoneRegex := `^7\d{10}$`
	matched, _ := regexp.MatchString(phoneRegex, phoneNumber)
	if !matched {
		msg := tgbotapi.NewMessage(message.Chat.ID, "⚠️ Номер телефона должен быть в формате 79999999999.")
		deleteMsg := tgbotapi.NewDeleteMessage(message.Chat.ID, h.lastMessageID)
		h.Bot.Send(deleteMsg)
		msgSend, err := h.Bot.Send(msg)
		if err != nil {
			log.Printf("Ошибка отправки сообщения: %v", err)
		}
		h.lastMessageID = msgSend.MessageID
		return
	}

	userID := strconv.Itoa(int(message.Chat.ID))
	err := internal.RequestPhoneAndUpdateID(userID, phoneNumber)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "❌ Вашего номера нет в базе. Обратитесь в поддержку.")
		deleteMsg := tgbotapi.NewDeleteMessage(message.Chat.ID, h.lastMessageID)
		h.Bot.Send(deleteMsg)
		msgSend, err := h.Bot.Send(msg)
		if err != nil {
			log.Printf("Ошибка отправки сообщения: %v", err)
		}
		h.lastMessageID = msgSend.MessageID
		return
	}

	// Удаление клавиатуры и отправка сообщения с инлайн-кнопкой
	buttonMain := tgbotapi.NewInlineKeyboardButtonData("🏠 В главное меню", "get_main")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(buttonMain),
	)

	// Сообщение "Вы успешно вошли в систему!" с инлайн-кнопкой
	msg := tgbotapi.NewMessage(message.Chat.ID, "✅ Вы успешно вошли в систему!")
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false) // Убираем клавиатуру для ввода
	msg.ReplyMarkup = keyboard                          // Добавляем инлайн-клавиатуру

	deleteMsg := tgbotapi.NewDeleteMessage(message.Chat.ID, h.lastMessageID)
	h.Bot.Send(deleteMsg)

	msgSend, err := h.Bot.Send(msg)
	if err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
	h.lastMessageID = msgSend.MessageID

	h.UserStates[message.Chat.ID] = StateNone
}
