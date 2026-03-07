package store

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

// UserState — прогресс одного пользователя.
type UserState struct {
	CurrentDay int
	Streak     int
	LastReadAt time.Time
	StartedAt  time.Time
}

type Store struct {
	db *sql.DB
}

func New(databaseURL string) (*Store, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS user_state (
			chat_id BIGINT PRIMARY KEY,
			current_day INTEGER NOT NULL DEFAULT 1,
			last_read_at TIMESTAMPTZ,
			streak INTEGER NOT NULL DEFAULT 0,
			started_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	return err
}

// EnsureUser создаёт запись пользователя, если её ещё нет (день 1, серия 0).
func (s *Store) EnsureUser(chatID int64) error {
	_, err := s.db.Exec(`
		INSERT INTO user_state (chat_id, current_day, streak, started_at)
		VALUES ($1, 1, 0, NOW())
		ON CONFLICT (chat_id) DO NOTHING;
	`, chatID)
	return err
}

// GetState возвращает состояние пользователя. Если нет — день 1, серия 0.
func (s *Store) GetState(chatID int64) (UserState, error) {
	if err := s.EnsureUser(chatID); err != nil {
		return UserState{}, err
	}
	var st UserState
	var lastReadAt, startedAt sql.NullTime
	err := s.db.QueryRow(`
		SELECT current_day, streak, last_read_at, started_at
		FROM user_state WHERE chat_id = $1
	`, chatID).Scan(&st.CurrentDay, &st.Streak, &lastReadAt, &startedAt)
	if err != nil {
		return UserState{}, err
	}
	if st.CurrentDay < 1 {
		st.CurrentDay = 1
	}
	if st.Streak < 0 {
		st.Streak = 0
	}
	if lastReadAt.Valid {
		st.LastReadAt = lastReadAt.Time
	}
	if startedAt.Valid {
		st.StartedAt = startedAt.Time
	} else {
		st.StartedAt = time.Now().UTC()
	}
	return st, nil
}

// MarkRead сохраняет прогресс: «прочитала сегодняшний день». Обновляет last_read_at и серию (streak).
// Номер дня (current_day) не меняется — его увеличивает только «Следующий день».
func (s *Store) MarkRead(chatID int64) error {
	if err := s.EnsureUser(chatID); err != nil {
		return err
	}
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	var lastReadAt sql.NullTime
	var streak int
	err := s.db.QueryRow(`SELECT streak, last_read_at FROM user_state WHERE chat_id = $1`, chatID).
		Scan(&streak, &lastReadAt)
	if err != nil {
		return err
	}

	if lastReadAt.Valid {
		lastDate := time.Date(lastReadAt.Time.Year(), lastReadAt.Time.Month(), lastReadAt.Time.Day(), 0, 0, 0, 0, time.UTC)
		daysDiff := int(today.Sub(lastDate).Hours() / 24)
		if daysDiff == 1 {
			streak++
		} else if daysDiff > 1 {
			streak = 1
		}
	} else {
		streak = 1
	}

	_, err = s.db.Exec(`
		UPDATE user_state SET last_read_at = $1, streak = $2 WHERE chat_id = $3
	`, now, streak, chatID)
	return err
}

// AdvanceDay увеличивает номер дня на 1 (например с 6 на 7) и возвращает новый день.
// Не трогает last_read_at и streak — прогресс сохраняется кнопкой «Прочитала!».
func (s *Store) AdvanceDay(chatID int64) (newDay int, planRestarted bool, err error) {
	if err := s.EnsureUser(chatID); err != nil {
		return 0, false, err
	}

	var currentDay int
	err = s.db.QueryRow(`SELECT current_day FROM user_state WHERE chat_id = $1`, chatID).Scan(&currentDay)
	if err != nil {
		return 0, false, err
	}

	newDay = currentDay + 1
	planRestarted = false
	if newDay > 365 {
		newDay = 1
		planRestarted = true
	}

	_, err = s.db.Exec(`UPDATE user_state SET current_day = $1 WHERE chat_id = $2`, newDay, chatID)
	if err != nil {
		return 0, false, err
	}
	return newDay, planRestarted, nil
}

// GetAllChatIDs возвращает все chat_id, для которых есть запись (для рассылки напоминаний).
func (s *Store) GetAllChatIDs() ([]int64, error) {
	rows, err := s.db.Query(`SELECT chat_id FROM user_state`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// LastReadWithin возвращает true, если пользователь нажимал «Прочитала!»/«Следующий день» в последние d.
func (s *Store) LastReadWithin(chatID int64, d time.Duration) bool {
	var lastReadAt sql.NullTime
	err := s.db.QueryRow(`SELECT last_read_at FROM user_state WHERE chat_id = $1`, chatID).Scan(&lastReadAt)
	if err != nil || !lastReadAt.Valid {
		return false
	}
	return time.Since(lastReadAt.Time) < d
}

func (s *Store) Close() error {
	return s.db.Close()
}
