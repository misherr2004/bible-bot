# Bible Bot — Yearly Bible Reading Plan Bot

Telegram bot that helps you read through the whole Bible in one year with a structured 3–track plan, personal progress tracking, and gentle daily reminders.

---

### Badges

<p align="left">
  <img src="https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go version" />
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License: MIT" />
  <img src="https://img.shields.io/badge/Telegram-Bot-2CA5E0?style=for-the-badge&logo=telegram&logoColor=white" alt="Telegram Bot" />
</p>

---

### Features

- ✝️ **Year‑long Bible reading plan** — 365‑day plan that covers the whole Bible in a balanced rhythm.
- 👤 **Personal progress per user** — each Telegram user has their own independent reading state stored in PostgreSQL.
- 📚 **3‑track daily reading** — Old Testament, New Testament, and Psalms/Proverbs combined into one thoughtful daily portion.
- 🔔 **Smart daily reminders** — one reminder per day only if the user has not interacted with the bot yet.
- 📈 **Readable progress metrics** — shows the current day, remaining days, and % of the entire Bible already read.
- 🔁 **Streaks and last activity** — tracks continuous reading days and last read date for motivation.
- ▶️ **Inline actions** — buttons like “Read for today”, “I’ve read”, and “Next day” directly under the plan message.
- 🧹 **Reset progress safely** — `/reset_progress` lets a user start over from day 1 without affecting others.
- 📣 **Admin broadcast** — admin‑only `/broadcast` command to send announcements to all active users (e.g., when the plan is updated).

---

### How It Works

- **Three daily tracks**
  - **Track 1 — Old Testament (without Psalms & Proverbs):** 748 chapters, usually 2 chapters per day, some days 3 to fit in 365 days.
  - **Track 2 — New Testament:** 260 chapters, 1 chapter per day; after day 260 the NT part is no longer shown.
  - **Track 3 — Psalms & Proverbs:** Psalms 1–150, then Proverbs 1–31 in a cycle (181 chapters looping through the year).

- **Per‑user state**
  - Each user is identified by their Telegram `chat_id`.
  - The bot stores `current_day`, `last_read_at`, `streak`, and `started_at` in PostgreSQL.

- **Streaks & reminders**
  - Any interaction with the reading (marking as read, moving to the next day) counts as activity for that calendar day.
  - If a day passes with no activity, the streak is reset on the next interaction.
  - Once per day (at `REMINDER_HOUR` in `TIMEZONE`), the scheduler checks users without activity today and sends them a reminder.

- **Admin broadcast**
  - A single admin (configured via `ADMIN_CHAT_ID`) can run `/broadcast <message>`.
  - The message is delivered to all known `chat_id`s from the database and the admin gets a short delivery summary.

---

### Tech Stack

- **Language:** Go (1.22+)
- **Telegram API:** [`github.com/go-telegram-bot-api/telegram-bot-api/v5`](https://github.com/go-telegram-bot-api/telegram-bot-api/v5)
- **Database:** PostgreSQL
- **Migrations & persistence:** custom store package (`internal/store`)
- **Scheduling:** in‑process daily scheduler (`internal/bot/scheduler.go`)
- **Containerization:** Docker (optional but recommended for deployment)

---

### Getting Started (Local Development)

#### 1. Prerequisites

- Go **1.22+**
- PostgreSQL **13+** (local or in Docker)
- Telegram bot token from **@BotFather**

#### 2. Clone the repository

```bash
git clone https://github.com/your-username/bible-bot.git
cd bible-bot
```

#### 3. Create `.env`

Use `.env.example` as a reference:

```bash
cp .env.example .env
```

Edit `.env` with your values:

```env
TELEGRAM_BOT_TOKEN=your_bot_token_here
DATABASE_URL=postgres://user:password@localhost:5432/bible_bot?sslmode=disable
REMINDER_HOUR=15
TIMEZONE=Europe/Moscow
ADMIN_CHAT_ID=123456789  # optional: your personal Telegram chat id for /broadcast
```

#### 4. Prepare PostgreSQL

If PostgreSQL is installed locally:

```bash
createdb bible_bot
```

Make sure `DATABASE_URL` in `.env` points to this database.

#### 5. Run the bot

```bash
go mod download
go run ./cmd/bot
```

You should see something like:

```text
Authorized as @your_bot_username
[scheduler] daily reminder at 15:00 Europe/Moscow
```

Now open Telegram, start a chat with your bot, and send `/start`.

---

### Docker

You can run both PostgreSQL and the bot via Docker or docker-compose.

#### 1. Run PostgreSQL via Docker

```bash
docker run -d --name bible-postgres \
  -e POSTGRES_USER=bot \
  -e POSTGRES_PASSWORD=bot \
  -e POSTGRES_DB=bible_bot \
  -p 5432:5432 \
  postgres:16-alpine
```

Then set in `.env`:

```env
DATABASE_URL=postgres://bot:bot@host.docker.internal:5432/bible_bot?sslmode=disable
```

or `localhost` if the bot runs on the host machine.

#### 2. Build and run the bot container

If the repository contains a `Dockerfile` similar to:

```bash
docker build -t bible-bot .
```

Then run:

```bash
docker run -d --name bible-bot \
  --env-file .env \
  --link bible-postgres:postgres \
  bible-bot
```

Alternatively, you can wire everything through `docker-compose` (PostgreSQL service + bot service) to keep configuration in one place.

---

### Project Structure

```text
.
├── cmd/
│   └── bot/                  # Main entrypoint (Telegram bot)
├── config/
│   └── config.go             # Environment-based configuration (tokens, DB, timezone, admin, etc.)
├── internal/
│   ├── bot/
│   │   ├── handler.go        # Telegram updates, commands, inline buttons
│   │   └── scheduler.go      # Daily reminder scheduler
│   ├── plan/
│   │   └── plan.go           # 3-track Bible reading plan and progress calculation
│   └── store/
│       └── store.go          # PostgreSQL access, user state, streak logic
├── .env.example              # Example environment configuration
├── Dockerfile                # (Optional) Container image for the bot
└── README.md
```

---

### Author

This project is created and maintained by a practicing Baptist who uses this bot personally and within a real Christian community.  
It is not just a technical pet project, but a tool born out of faith and a desire to help believers consistently read and meditate on Scripture throughout the year.

If you have suggestions, feedback, or want to use this bot in your church or home group, feel free to open an issue or reach out.

---

### License

This project is licensed under the **MIT License**. See the `LICENSE` file for details.

# Bible Bot — план чтения Библии на год в Telegram

Бот для ежедневного плана чтения Библии (365 дней). Несколько пользователей, у каждого свой прогресс в PostgreSQL, напоминания в 12:00.

## Локальный запуск (тест)

### Вариант 1: PostgreSQL уже установлен

```bash
cd /Users/user/bible-bot
createdb bible_bot
# В .env пропиши TELEGRAM_BOT_TOKEN и DATABASE_URL=postgres://ТВОЙ_ЮЗЕР:ПАРОЛЬ@localhost:5432/bible_bot?sslmode=disable
go mod download
go run ./cmd/bot
```

### Вариант 2: PostgreSQL в Docker (если нет установленного Postgres)

```bash
cd /Users/user/bible-bot
docker run -d --name bible-postgres -e POSTGRES_USER=bot -e POSTGRES_PASSWORD=bot -e POSTGRES_DB=bible_bot -p 5432:5432 postgres:16-alpine
# Подожди пару секунд, пока БД поднимется
# В .env: DATABASE_URL=postgres://bot:bot@localhost:5432/bible_bot?sslmode=disable
go mod download
go run ./cmd/bot
```

### После запуска

В консоли должно быть: `Authorized as @...`, `[scheduler] daily reminder at ...`. В Telegram: **Start** → `/read` → «Прочитал(а)!» → «Следующий день» → проверь план на следующий день. Остановка: `Ctrl+C`.

**Про Docker для самого бота:** запуск в Docker не обязателен для локального теста — достаточно `go run ./cmd/bot` и Postgres (локально или в контейнере). Docker для бота полезен на сервере (один образ с бинарником). При желании можно добавить `Dockerfile` и `docker-compose` для деплоя.

## Пуш в GitHub

**.env в репозиторий не пушить** — там токен бота и пароль БД. В `.gitignore` уже есть `.env`, он не попадёт в коммит.

Если репозиторий ещё не создан на GitHub — создай пустой репозиторий (без README), затем:

```bash
cd /Users/user/bible-bot
git init
git add .
git commit -m "Bible reading plan Telegram bot"
git branch -M main
git remote add origin https://github.com/ТВОЙ_ЛОГИН/bible-bot.git
git push -u origin main
```

Подставь свой логин GitHub вместо `ТВОЙ_ЛОГИН`. Если репо уже есть и ты просто добавляешь remote — достаточно `git remote add origin ...`, потом `git push -u origin main`.

---

## Бесплатный хостинг навсегда (без триала)

Чтобы бот работал 24/7 и не зависел от твоего ноутбука, нужен свой сервер. **По-настоящему бесплатно без срока** даёт:

### Oracle Cloud Free Tier

- **Бессрочно бесплатно**: 1–2 виртуальные машины (VPS), без ограничения по времени.
- Нужна банковская карта при регистрации (списаний не будет, если не выходить за лимиты).
- Регистрация: [cloud.oracle.com](https://cloud.oracle.com) → Create a free account.

**Кратко по шагам:**

1. **Зарегистрироваться** в Oracle Cloud, выбрать регион (можно оставить по умолчанию).
2. **Создать инстанс (VM):**
   - Menu → Compute → Instances → Create instance.
   - Name: например `bible-bot`.
   - Image: **Ubuntu 22.04**.
   - Shape: оставить **Ampere** (ARM) или **VM.Standard.E2.1.Micro** (AMD) — оба входят в Always Free.
   - Создать/выбрать SSH-ключ или сгенерировать и скачать — понадобится для входа.
   - Create.
3. **Открыть порт SSH (22):**
   - Instance → Subnet → Default Security List → Add Ingress Rule: Source `0.0.0.0/0`, port 22, TCP.
4. **Подключиться по SSH** (в консоли Oracle указан пример, что-то вроде):
   ```bash
   ssh -i путь/до/ключа ubuntu@ПУБЛИЧЕСКИЙ_IP_ИНСТАНСА
   ```
   Публичный IP смотри на странице инстанса.

Дальше на этой VM делаешь всё как в разделе **«Деплой на отдельный сервер»** ниже: установка Go и Docker, клонирование репо, Postgres в Docker, `.env`, сборка бота и systemd. Один сервер — один раз настроил, бот крутится бесплатно и постоянно.

**Если Oracle недоступен (например из РФ даже с VPN):** бери недорогой VPS у российского провайдера — платно, но дёшево (обычно от ~200–300 ₽/мес за минимальный тариф). Ниже раздел про деплой на таком VPS.

---

## Деплой на VPS (в т.ч. российский)

Подойдёт любой VPS с Ubuntu 22.04: **Timeweb**, **Selectel**, **Reg.ru**, **Beget**, **Hetzner** и т.п. Российские хостинги работают без VPN, оплата картой РФ; минимальный тариф обычно 200–400 ₽/мес — хватит для бота с малым числом пользователей.

### 1. На сервере: установить Go и Docker (для Postgres)

```bash
sudo apt update && sudo apt install -y golang-go docker.io
sudo systemctl start docker
sudo systemctl enable docker
```

### 2. Клонировать репозиторий

```bash
cd ~
git clone https://github.com/ТВОЙ_ЛОГИН/bible-bot.git
cd bible-bot
```

### 3. Запустить PostgreSQL в Docker

```bash
docker run -d --name bible-postgres \
  -e POSTGRES_USER=bot -e POSTGRES_PASSWORD=ПРИДУМАЙ_НАДЁЖНЫЙ_ПАРОЛЬ \
  -e POSTGRES_DB=bible_bot -p 5432:5432 \
  postgres:16-alpine
```

### 4. Создать .env на сервере (вручную)

```bash
nano .env
```

Вставь (подставь свой токен и пароль от Postgres):

```
TELEGRAM_BOT_TOKEN=твой_токен_от_BotFather
DATABASE_URL=postgres://bot:ПРИДУМАЙ_НАДЁЖНЫЙ_ПАРОЛЬ@localhost:5432/bible_bot?sslmode=disable
REMINDER_HOUR=12
TIMEZONE=Europe/Moscow
```

Сохрани (Ctrl+O, Enter, Ctrl+X).

### 5. Собрать и запустить бота

```bash
go build -o bible-bot ./cmd/bot
nohup ./bible-bot > bot.log 2>&1 &
```

Проверка: `tail -f bot.log` — должны быть строки `Authorized as @...` и про scheduler.

Чтобы бот перезапускался после перезагрузки сервера и падений — настрой systemd (см. ниже).

### Автозапуск через systemd (опционально)

Создай файл `/etc/systemd/system/bible-bot.service`:

```ini
[Unit]
Description=Bible reading plan Telegram bot
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/bible-bot
ExecStart=/home/ubuntu/bible-bot/bible-bot
Restart=always
RestartSec=5
EnvironmentFile=/home/ubuntu/bible-bot/.env

[Install]
WantedBy=multi-user.target
```

Подставь свой путь и пользователя. Затем:

```bash
sudo systemctl daemon-reload
sudo systemctl enable bible-bot
sudo systemctl start bible-bot
sudo systemctl status bible-bot
```

Логи: `journalctl -u bible-bot -f`
