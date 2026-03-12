package bot

import (
	"bible-bot/config"
	"bible-bot/internal/store"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

// StartScheduler запускает ежедневное напоминание. Каждый день в REMINDER_HOUR (например 15:00)
// проверяет: если пользователь до этого времени не нажимал ни «Прочитал(а)!», ни «Следующий день» —
// отправляется напоминание. Стрик обнуляется, если до 00:00 не было ни одного нажатия (логика в store).
func StartScheduler(cfg *config.Config, st *store.Store, sendReminderTo func(chatID int64)) {
	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		log.Printf("[scheduler] invalid timezone %q, using UTC: %v", cfg.Timezone, err)
		loc = time.UTC
	}

	c := cron.New(cron.WithLocation(loc))
	_, err = c.AddFunc(cronSpec(cfg.ReminderHour), func() {
		chatIDs, err := st.GetAllChatIDs()
		if err != nil {
			log.Printf("[scheduler] get chat ids: %v", err)
			return
		}
		for _, chatID := range chatIDs {
			if st.HadActivityToday(chatID, loc) {
				continue
			}
			sendReminderTo(chatID)
		}
	})
	if err != nil {
		log.Printf("[scheduler] failed to add cron job: %v", err)
		return
	}

	c.Start()
	log.Printf("[scheduler] daily reminder at %d:00 %s (all users)", cfg.ReminderHour, cfg.Timezone)
}

func cronSpec(hour int) string {
	return fmt.Sprintf("0 %d * * *", hour)
}
