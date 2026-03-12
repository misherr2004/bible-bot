package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	TelegramBotToken string
	DatabaseURL      string
	ReminderHour     int
	Timezone         string
}

func Load() (*Config, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required (PostgreSQL connection string)")
	}

	reminderHour := 15
	if h := os.Getenv("REMINDER_HOUR"); h != "" {
		if parsed, err := strconv.Atoi(h); err == nil && parsed >= 0 && parsed <= 23 {
			reminderHour = parsed
		}
	}

	timezone := "Europe/Moscow"
	if tz := os.Getenv("TIMEZONE"); tz != "" {
		timezone = tz
	}

	return &Config{
		TelegramBotToken: token,
		DatabaseURL:      databaseURL,
		ReminderHour:     reminderHour,
		Timezone:         timezone,
	}, nil
}
