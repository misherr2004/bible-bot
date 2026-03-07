package state

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type State struct {
	CurrentDay int       `json:"current_day"`
	LastReadAt time.Time `json:"last_read_at"`
	Streak     int       `json:"streak"`
	StartedAt  time.Time `json:"started_at"`
	ChatID     int64     `json:"chat_id,omitempty"`
	mu         sync.RWMutex
	filePath   string
}

type persistedState struct {
	CurrentDay int    `json:"current_day"`
	LastReadAt string `json:"last_read_at"`
	Streak     int    `json:"streak"`
	StartedAt  string `json:"started_at"`
	ChatID     int64  `json:"chat_id,omitempty"`
}

func New(filePath string) (*State, error) {
	s := &State{
		filePath: filePath,
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if s.CurrentDay == 0 {
		s.CurrentDay = 1
		s.StartedAt = time.Now().UTC()
		s.LastReadAt = time.Time{}
		s.Streak = 0
		_ = s.save()
	}
	return s, nil
}

func (s *State) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	var p persistedState
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}

	s.CurrentDay = p.CurrentDay
	if p.CurrentDay <= 0 {
		s.CurrentDay = 1
	}
	s.Streak = p.Streak
	if p.Streak < 0 {
		s.Streak = 0
	}

	if p.LastReadAt != "" {
		if t, err := time.Parse(time.RFC3339, p.LastReadAt); err == nil {
			s.LastReadAt = t
		}
	}
	if p.StartedAt != "" {
		if t, err := time.Parse(time.RFC3339, p.StartedAt); err == nil {
			s.StartedAt = t
		} else {
			s.StartedAt = time.Now().UTC()
		}
	} else {
		s.StartedAt = time.Now().UTC()
	}
	s.ChatID = p.ChatID

	return nil
}

func (s *State) GetChatID() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ChatID
}

func (s *State) SetChatID(chatID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ChatID = chatID
	return s.save()
}

// TransferToNewOwner clears the current user and resets reading progress.
// The next user who sends /start becomes the owner and starts from day 1.
func (s *State) TransferToNewOwner() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ChatID = 0
	s.CurrentDay = 1
	s.Streak = 0
	s.LastReadAt = time.Time{}
	s.StartedAt = time.Now().UTC()
	return s.save()
}

func (s *State) save() error {
	p := persistedState{
		CurrentDay: s.CurrentDay,
		Streak:     s.Streak,
		LastReadAt: s.LastReadAt.Format(time.RFC3339),
		StartedAt:  s.StartedAt.Format(time.RFC3339),
		ChatID:     s.ChatID,
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}

// AdvanceDay увеличивает текущий день на 1. Можно вызывать несколько раз в сутки — каждый раз +1 день.
// Серия (streak) обновляется только при первом нажатии за календарный день.
func (s *State) AdvanceDay() (newDay int, planRestarted bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	planRestarted = false
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Серия: только если последнее чтение было в другой календарный день
	if s.LastReadAt.IsZero() {
		s.Streak = 1
	} else {
		lastDate := time.Date(s.LastReadAt.Year(), s.LastReadAt.Month(), s.LastReadAt.Day(), 0, 0, 0, 0, time.UTC)
		daysDiff := int(today.Sub(lastDate).Hours() / 24)
		if daysDiff == 1 {
			s.Streak++
		} else if daysDiff > 1 {
			s.Streak = 1
		}
		// daysDiff == 0: уже нажимали сегодня — серию не меняем
	}

	s.LastReadAt = now
	s.CurrentDay++
	if s.CurrentDay > 365 {
		s.CurrentDay = 1
		planRestarted = true
	}
	newDay = s.CurrentDay
	return newDay, planRestarted, s.save()
}

func (s *State) GetSnapshot() (currentDay, streak int, lastReadAt, startedAt time.Time) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.CurrentDay, s.Streak, s.LastReadAt, s.StartedAt
}

func (s *State) LastReadWithin(d time.Duration) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.LastReadAt.IsZero() {
		return false
	}
	return time.Since(s.LastReadAt) < d
}
