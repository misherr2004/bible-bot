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
