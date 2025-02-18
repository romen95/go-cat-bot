package main

import (
	"log"
	"os"

	"gocarbot/bot"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Ошибка загрузки файла .env")
	}
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	handler := bot.NewBotHandler(botToken)

	log.Println("Запуск Telegram-бота...")
	handler.Run()
}
