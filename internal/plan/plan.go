package plan

import "strconv"

// DayPlan — план на один день: три трека (ВЗ без Псалмов/Притчей, НЗ, Псалмы/Притчи).
type DayPlan struct {
	Day        int
	OT         string // например "Числа 3–4"
	NT         string // например "Лука 5" или "" после дня 260
	PsalmsProverbs string // например "Псалом 47"
}

// Треки идут последовательно, без пропусков.
// Трек 1: ВЗ без Псалтири и Притчей — 748 глав. 18 дней по 3 главы, остальные по 2.
// Трек 2: НЗ — 260 глав, по 1 в день; после дня 260 — пусто.
// Трек 3: Псалтирь 1–150, затем Притчи 1–31, цикл 181 глава.

const (
	track1Chapters = 748
	track2Chapters = 260
	track3Cycle    = 181 // 150 + 31
	track1Days3    = 18
)

// Дни (1–365), в которые в треке 1 читаем по 3 главы (остальные — по 2).
var track1ThreeChapterDays = func() map[int]bool {
	m := make(map[int]bool)
	// Равномерно: 20, 40, 60, ..., 360
	for i := 20; i <= 360; i += 20 {
		m[i] = true
	}
	if len(m) != track1Days3 {
		// fallback: первые 18 дней с 3 главами для простоты
		m = make(map[int]bool)
		for i := 1; i <= track1Days3; i++ {
			m[i] = true
		}
	}
	return m
}()

// ВЗ без Псалтири и Притчей (русские названия книг и кол-во глав).
var track1Books = []struct {
	name  string
	chaps int
}{
	{"Бытие", 50}, {"Исход", 40}, {"Левит", 27}, {"Числа", 36}, {"Второзаконие", 34},
	{"Иисус Навин", 24}, {"Судей", 21}, {"Руфь", 4}, {"1 Царств", 31}, {"2 Царств", 24},
	{"3 Царств", 22}, {"4 Царств", 25}, {"1 Паралипоменон", 29}, {"2 Паралипоменон", 36},
	{"Ездра", 10}, {"Неемия", 13}, {"Есфирь", 10}, {"Иов", 42},
	{"Екклесиаст", 12}, {"Песня песней", 8}, {"Исаия", 66}, {"Иеремия", 52},
	{"Плач Иеремии", 5}, {"Иезекииль", 48}, {"Даниил", 12},
	{"Осия", 14}, {"Иоиль", 3}, {"Амос", 9}, {"Авдий", 1}, {"Иона", 4},
	{"Михей", 7}, {"Наум", 3}, {"Аввакум", 3}, {"Софония", 3}, {"Аггей", 2},
	{"Захария", 14}, {"Малахия", 4},
}

// НЗ (русские названия и кол-во глав).
var track2Books = []struct {
	name  string
	chaps int
}{
	{"От Матфея", 28}, {"От Марка", 16}, {"От Луки", 24}, {"От Иоанна", 21},
	{"Деяния", 28}, {"Иакова", 5}, {"1 Петра", 5}, {"2 Петра", 3}, {"1 Иоанна", 5},
	{"2 Иоанна", 1}, {"3 Иоанна", 1}, {"Иуды", 1}, {"Римлянам", 16},
	{"1 Коринфянам", 16}, {"2 Коринфянам", 13}, {"Галатам", 6}, {"Ефесянам", 6},
	{"Филиппийцам", 4}, {"Колоссянам", 4}, {"1 Фессалоникийцам", 5},
	{"2 Фессалоникийцам", 3}, {"1 Тимофею", 6}, {"2 Тимофею", 4}, {"Титу", 3},
	{"Филимону", 1}, {"Евреям", 13}, {"Откровение", 22},
}

// chapterRef возвращает ссылку вида "Книга N" или "Книга N–M" по индексу главы в слайсе (может быть через две книги).
func chapterRef(books []struct{ name string; chaps int }, startIdx int, count int) string {
	if count <= 0 || startIdx < 0 {
		return ""
	}
	var idx int
	for _, b := range books {
		if startIdx < idx+b.chaps {
			ch := startIdx - idx + 1
			if count == 1 {
				return b.name + " " + strconv.Itoa(ch)
			}
			chEnd := ch + count - 1
			if chEnd <= b.chaps {
				return b.name + " " + strconv.Itoa(ch) + "–" + strconv.Itoa(chEnd)
			}
			// диапазон переходит в следующую книгу
			rest := count - (b.chaps - ch + 1)
			nextStart := idx + b.chaps
			return b.name + " " + strconv.Itoa(ch) + "–" + strconv.Itoa(b.chaps) + ", " + chapterRef(books, nextStart, rest)
		}
		idx += b.chaps
	}
	return ""
}

// track1ChapterRef возвращает одну ссылку для трека 1 (например "Числа 3–4").
func track1ChapterRef(startIdx, count int) string {
	return chapterRef(track1Books, startIdx, count)
}

// track2ChapterRef — одна глава НЗ по индексу 0..259.
func track2ChapterRef(chapIdx int) string {
	if chapIdx < 0 {
		return ""
	}
	return chapterRef(track2Books, chapIdx, 1)
}

// track3ChapterRef — по индексу в цикле 0..180 (Псалтирь 1–150, Притчи 1–31).
func track3ChapterRef(cycleIdx int) string {
	if cycleIdx < 0 {
		return ""
	}
	cycleIdx = cycleIdx % track3Cycle
	if cycleIdx < 150 {
		return "Псалом " + strconv.Itoa(cycleIdx+1)
	}
	return "Притчи " + strconv.Itoa(cycleIdx-150+1)
}

// startIndexTrack1 возвращает индекс первой главы для дня d (1-based) в треке 1.
func startIndexTrack1(day int) int {
	if day <= 0 {
		return 0
	}
	n := (day - 1) * 2
	for i := 20; i < day; i += 20 {
		n++
	}
	return n
}

// GetDay возвращает план на день day (1–365).
func GetDay(day int) DayPlan {
	if day < 1 {
		day = 1
	}
	if day > 365 {
		day = ((day-1) % 365) + 1
	}

	// Трек 1
	start1 := startIndexTrack1(day)
	count1 := 2
	if track1ThreeChapterDays[day] {
		count1 = 3
	}
	ot := track1ChapterRef(start1, count1)

	// Трек 2
	var nt string
	if day <= track2Chapters {
		nt = track2ChapterRef(day - 1)
	}

	// Трек 3
	t3Idx := (day - 1) % track3Cycle
	psalms := track3ChapterRef(t3Idx)

	return DayPlan{
		Day:            day,
		OT:             ot,
		NT:             nt,
		PsalmsProverbs: psalms,
	}
}

// TotalBibleChapters — всего глав в Библии (для процента).
const TotalBibleChapters = 1189

// ChaptersReadByDay возвращает число «прочитанных» глав к концу дня day (1–365) по трём трекам.
func ChaptersReadByDay(day int) int {
	if day <= 0 {
		return 0
	}
	c1 := 2
	if track1ThreeChapterDays[day] {
		c1 = 3
	}
	t1 := startIndexTrack1(day) + c1
	if t1 > track1Chapters {
		t1 = track1Chapters
	}
	t2 := day
	if t2 > track2Chapters {
		t2 = track2Chapters
	}
	t3 := day
	return t1 + t2 + t3
}

// PercentRead возвращает процент прочитанного от всей Библии к концу дня day.
func PercentRead(day int) int {
	if day <= 0 {
		return 0
	}
	read := ChaptersReadByDay(day)
	if read > TotalBibleChapters {
		read = TotalBibleChapters
	}
	return read * 100 / TotalBibleChapters
}
