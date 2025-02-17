package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Google Sheets setup
const (
	spreadsheetID = "your_spreadsheet_id"
	rangeName     = "Orders!A:B" // Пример диапазона с колонками: ID | Статус
)

func getOrderStatus(orderID string) (string, error) {
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("credentials.json"))
	if err != nil {
		return "", err
	}

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, rangeName).Do()
	if err != nil {
		return "", err
	}

	for _, row := range resp.Values {
		if len(row) >= 2 && row[0] == orderID {
			return fmt.Sprintf("Статус заказа %s: %s", orderID, row[1]), nil
		}
	}

	return "Статус не найден", nil
}

func main() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			chatID := update.Message.Chat.ID
			switch update.Message.Text {
			case "/start":
				keyboard := tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("Статус заказа"),
					),
				)
				msg := tgbotapi.NewMessage(chatID, "Добро пожаловать! Нажмите 'Статус заказа', чтобы узнать статус.")
				msg.ReplyMarkup = keyboard
				bot.Send(msg)

			case "Статус заказа":
				msg := tgbotapi.NewMessage(chatID, "Пожалуйста, введите номер заказа.")
				bot.Send(msg)

			default:
				orderID := update.Message.Text
				status, err := getOrderStatus(orderID)
				if err != nil {
					msg := tgbotapi.NewMessage(chatID, "Ошибка при получении статуса заказа")
					bot.Send(msg)
					continue
				}
				msg := tgbotapi.NewMessage(chatID, status)
				bot.Send(msg)
			}
		}
	}
}
