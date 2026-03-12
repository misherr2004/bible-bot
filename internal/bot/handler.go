package bot

import (
	"bible-bot/internal/plan"
	"bible-bot/internal/store"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	btnReadDone      = "✅ Прочитал(а)!"
	btnNextDay       = "📖 Следующий день"
	callbackReadDone = "read_done"
	callbackNextDay  = "next_day"
)

type Handler struct {
	api   *tgbotapi.BotAPI
	store *store.Store
}

func NewHandler(api *tgbotapi.BotAPI, st *store.Store) *Handler {
	return &Handler{api: api, store: st}
}

func (h *Handler) HandleUpdate(update tgbotapi.Update) {
	if update.Message != nil {
		h.handleMessage(update.Message)
		return
	}
	if update.CallbackQuery != nil {
		h.handleCallback(update.CallbackQuery)
	}
}

func (h *Handler) handleMessage(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	if msg.Command() == "" {
		return
	}
	switch msg.Command() {
	case "start":
		_ = h.store.EnsureUser(chatID)
		h.cmdStart(chatID)
	case "read":
		h.cmdRead(chatID)
	case "progress":
		h.cmdProgress(chatID)
	case "reset_progress":
		h.cmdResetProgress(chatID)
	default:
		// Игнорируем неизвестные команды
	}
}

func (h *Handler) cmdStart(chatID int64) {
	text := `Привет! ✝

Здесь план чтения Библии на год: каждый день — три отрывка (Ветхий завет, Новый завет, Псалмы или Притчи). За 365 дней проходим всё Писание.

Когда прочитал(а) — нажми «Прочитал(а)!». На следующий день нажми «Следующий день» — пришлю план.

Команда /read покажет, что читать сегодня. Да благословит тебя Господь.`
	m := tgbotapi.NewMessage(chatID, text)
	_, _ = h.api.Send(m)
}

func (h *Handler) cmdRead(chatID int64) {
	st, err := h.store.GetState(chatID)
	if err != nil {
		_, _ = h.api.Send(tgbotapi.NewMessage(chatID, "Ошибка загрузки прогресса. Попробуй /start или позже."))
		return
	}
	dayPlan := plan.GetDay(st.CurrentDay)
	daysLeft := 365 - st.CurrentDay
	if daysLeft < 0 {
		daysLeft = 0
	}

	pct := plan.PercentRead(st.CurrentDay)
	var b strings.Builder
	b.WriteString(fmt.Sprintf("📖 День %d из 365 (%d%%) — осталось %d дней\n\n", st.CurrentDay, pct, daysLeft))
	b.WriteString("Сегодня читай:\n\n")
	b.WriteString("• Ветхий завет: " + dayPlan.OT + "\n")
	if dayPlan.NT != "" {
		b.WriteString("• Новый завет: " + dayPlan.NT + "\n")
	}
	b.WriteString("• " + dayPlan.PsalmsProverbs + "\n\n")
	if st.Streak > 0 {
		b.WriteString(fmt.Sprintf("🔥 Уже %d дней подряд в плане — продолжай в том же духе.\n\n", st.Streak))
	}
	b.WriteString("Когда прочитаешь — нажми кнопку ниже.")

	m := tgbotapi.NewMessage(chatID, b.String())
	m.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnReadDone, callbackReadDone),
		),
	)
	_, _ = h.api.Send(m)
}

func (h *Handler) cmdProgress(chatID int64) {
	st, err := h.store.GetState(chatID)
	if err != nil {
		_, _ = h.api.Send(tgbotapi.NewMessage(chatID, "Ошибка загрузки прогресса. Попробуй /start или позже."))
		return
	}
	daysLeft := 365 - st.CurrentDay
	if daysLeft < 0 {
		daysLeft = 0
	}
	pct := (st.CurrentDay * 100) / 365
	if pct > 100 {
		pct = 100
	}

	pctBible := plan.PercentRead(st.CurrentDay)
	var b strings.Builder
	b.WriteString("📊 Твой прогресс\n\n")
	b.WriteString(fmt.Sprintf("Текущий день: %d из 365 (%d%% от Библии)\n", st.CurrentDay, pctBible))
	b.WriteString(fmt.Sprintf("Осталось дней: %d\n", daysLeft))
	b.WriteString(fmt.Sprintf("Пройдено по плану: %d%%\n\n", pct))
	if st.Streak > 0 {
		b.WriteString(fmt.Sprintf("🔥 Серия: %d дней подряд\n", st.Streak))
	} else {
		b.WriteString("Серия: пока нет — начни читать и нажми «Прочитал(а)!».\n")
	}
	if !st.LastReadAt.IsZero() {
		b.WriteString(fmt.Sprintf("\nПоследнее чтение: %s", st.LastReadAt.Format("02.01.2006 15:04")))
	}
	if !st.StartedAt.IsZero() {
		b.WriteString(fmt.Sprintf("\nНачало: %s", st.StartedAt.Format("02.01.2006")))
	}

	_, _ = h.api.Send(tgbotapi.NewMessage(chatID, b.String()))
}

// cmdResetProgress обнуляет прогресс того, кто отправил команду. День 1, серия 0 — старт заново.
func (h *Handler) cmdResetProgress(chatID int64) {
	if err := h.store.ResetProgress(chatID); err != nil {
		_, _ = h.api.Send(tgbotapi.NewMessage(chatID, "Не удалось обнулить прогресс. Попробуй ещё раз."))
		return
	}
	_, _ = h.api.Send(tgbotapi.NewMessage(chatID, "Прогресс обнулён. Ты снова на дне 1. Нажми /read — пришлю план на сегодня."))
}

func (h *Handler) handleCallback(cq *tgbotapi.CallbackQuery) {
	chatID := cq.Message.Chat.ID
	callback := tgbotapi.NewCallback(cq.ID, "")
	_, _ = h.api.Request(callback)

	switch cq.Data {
	case callbackReadDone:
		h.doReadDone(chatID)
	case callbackNextDay:
		h.doNextDay(chatID)
	default:
		_, _ = h.api.Send(tgbotapi.NewMessage(chatID, "Неизвестное действие."))
	}
}

// doReadDone — «Прочитал(а)!»: только сохранение прогресса (last_read_at, серия). День не меняется.
func (h *Handler) doReadDone(chatID int64) {
	if err := h.store.MarkRead(chatID); err != nil {
		_, _ = h.api.Send(tgbotapi.NewMessage(chatID, "Не удалось сохранить прогресс. Попробуй ещё раз."))
		return
	}
	st, _ := h.store.GetState(chatID)
	praise := "Записал(а). Пусть Слово приносит плод в твоей жизни."
	if st.Streak > 0 {
		praise += fmt.Sprintf("\n\n🔥 Уже %d дней подряд — да хранит тебя Господь.", st.Streak)
	}
	praise += "\n\nКогда будешь готов(а) к новому дню — нажми «Следующий день», пришлю план."

	m := tgbotapi.NewMessage(chatID, praise)
	m.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnNextDay, callbackNextDay),
		),
	)
	_, _ = h.api.Send(m)
}

// doNextDay — «Следующий день»: день +1 (например 6→7), сразу присылаем план на новый день.
func (h *Handler) doNextDay(chatID int64) {
	newDay, planRestarted, err := h.store.AdvanceDay(chatID)
	if err != nil {
		_, _ = h.api.Send(tgbotapi.NewMessage(chatID, "Не удалось перейти к следующему дню. Попробуй ещё раз."))
		return
	}
	st, _ := h.store.GetState(chatID)
	dayPlan := plan.GetDay(newDay)
	daysLeft := 365 - newDay
	if daysLeft < 0 {
		daysLeft = 0
	}

	var b strings.Builder
	pct := plan.PercentRead(newDay)
	if planRestarted {
		b.WriteString("🎉 Ты прочитал(а) всю Библию за этот цикл. Слава Богу! Можем начать план заново — нажми «Следующий день».\n\n")
	}
	b.WriteString(fmt.Sprintf("📖 День %d из 365 (%d%%) — осталось %d дней\n\n", newDay, pct, daysLeft))
	b.WriteString("Сегодня читай:\n\n")
	b.WriteString("• Ветхий завет: " + dayPlan.OT + "\n")
	if dayPlan.NT != "" {
		b.WriteString("• Новый завет: " + dayPlan.NT + "\n")
	}
	b.WriteString("• " + dayPlan.PsalmsProverbs + "\n\n")
	if st.Streak > 0 {
		b.WriteString(fmt.Sprintf("🔥 Уже %d дней подряд в плане — продолжай в том же духе.\n\n", st.Streak))
	}
	b.WriteString("Когда прочитаешь — нажми кнопку ниже.")

	m := tgbotapi.NewMessage(chatID, b.String())
	m.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(btnReadDone, callbackReadDone),
		),
	)
	_, _ = h.api.Send(m)
}
