package main

import (
	"bible-bot/config"
	"bible-bot/internal/bot"
	"bible-bot/internal/store"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const reminderMessage = "Сегодня ты ещё не открывал(а) план чтения. Загляни в бота — там отрывки на этот день."

func main() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatalf("load .env: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	api, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		log.Fatalf("telegram bot: %v", err)
	}
	api.Debug = false
	log.Printf("Authorized as @%s", api.Self.UserName)

	st, err := store.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer st.Close()

	handler := bot.NewHandler(api, st)

	sendReminderTo := func(chatID int64) {
		msg := tgbotapi.NewMessage(chatID, reminderMessage)
		if _, err := api.Send(msg); err != nil {
			log.Printf("[reminder] chat %d: %v", chatID, err)
		}
	}

	go bot.StartScheduler(cfg, st, sendReminderTo)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := api.GetUpdatesChan(u)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-stop:
			log.Println("Shutting down...")
			return
		case update := <-updates:
			handler.HandleUpdate(update)
		}
	}
}
