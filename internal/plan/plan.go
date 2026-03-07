package plan

import "strconv"

// DayPlan represents one day of the reading plan.
type DayPlan struct {
	Day      int
	Passages []string
}

// AllDays returns the full 365-day Bible reading plan.
// Each day has 3–4 passages from different parts of the Bible (beginning, middle, end),
// so you get variety: not Genesis 1, 2, 3, 4 in a row.
func AllDays() []DayPlan {
	books := []struct {
		name  string
		chaps int
	}{
		{"Бытие", 50}, {"Исход", 40}, {"Левит", 27}, {"Числа", 36}, {"Второзаконие", 34},
		{"Иисус Навин", 24}, {"Судей", 21}, {"Руфь", 4}, {"1 Царств", 31}, {"2 Царств", 24},
		{"3 Царств", 22}, {"4 Царств", 25}, {"1 Паралипоменон", 29}, {"2 Паралипоменон", 36},
		{"Ездра", 10}, {"Неемия", 13}, {"Есфирь", 10}, {"Иов", 42}, {"Псалтирь", 150},
		{"Притчи", 31}, {"Екклесиаст", 12}, {"Песня песней", 8}, {"Исаия", 66},
		{"Иеремия", 52}, {"Плач Иеремии", 5}, {"Иезекииль", 48}, {"Даниил", 12},
		{"Осия", 14}, {"Иоиль", 3}, {"Амос", 9}, {"Авдий", 1}, {"Иона", 4},
		{"Михей", 7}, {"Наум", 3}, {"Аввакум", 3}, {"Софония", 3}, {"Аггей", 2},
		{"Захария", 14}, {"Малахия", 4},
		{"От Матфея", 28}, {"От Марка", 16}, {"От Луки", 24}, {"От Иоанна", 21},
		{"Деяния", 28}, {"Иакова", 5}, {"1 Петра", 5}, {"2 Петра", 3}, {"1 Иоанна", 5},
		{"2 Иоанна", 1}, {"3 Иоанна", 1}, {"Иуды", 1}, {"Римлянам", 16},
		{"1 Коринфянам", 16}, {"2 Коринфянам", 13}, {"Галатам", 6}, {"Ефесянам", 6},
		{"Филиппийцам", 4}, {"Колоссянам", 4}, {"1 Фессалоникийцам", 5},
		{"2 Фессалоникийцам", 3}, {"1 Тимофею", 6}, {"2 Тимофею", 4}, {"Титу", 3},
		{"Филимону", 1}, {"Евреям", 13}, {"Откровение", 22},
	}

	var chapters []string
	for _, b := range books {
		for c := 1; c <= b.chaps; c++ {
			chapters = append(chapters, b.name+" "+strconv.Itoa(c))
		}
	}

	// Каждый день — отрывки из разных частей Библии (не подряд Бытие 1,2,3).
	// Колонки: начало (0–364), середина (365–729), конец (730–1094); у 94 дней ещё 4-й отрывок.
	days := make([]DayPlan, 0, 365)
	for d := 0; d < 365; d++ {
		passages := []string{
			chapters[d],
			chapters[365+d],
			chapters[730+d],
		}
		if d < 94 {
			passages = append(passages, chapters[1095+d])
		}
		days = append(days, DayPlan{Day: d + 1, Passages: passages})
	}
	return days
}

// GetDay returns the plan for a given day (1-365). If day is out of range, wraps.
func GetDay(day int) DayPlan {
	all := AllDays()
	if day < 1 {
		day = 1
	}
	if day > 365 {
		day = ((day-1) % 365) + 1
	}
	return all[day-1]
}
