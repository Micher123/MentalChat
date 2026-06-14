# 📋 Индекс проекта MentalChat

> **Дата индексации:** 14.06.2026
> **Версия:** 1.1.0
> **Обновление:** chat_index_service, context_service, pull-синхронизация, GetMessageByLocalID

---

## 1. Общая информация

**MentalChat** — веб-приложение для общения с AI-специалистами (психолог, таролог, сексолог, гадалка). Состоит из Go-бэкенда (REST API на Gin + GORM + PostgreSQL) и React-фронтенда (TypeScript, Vite, TailwindCSS, Zustand).

| Компонент | Технологии | Порт |
|-----------|-----------|------|
| Backend | Go 1.21+, Gin, GORM, PostgreSQL, Viper | 8080 |
| Frontend | React 18, TypeScript, Vite, TailwindCSS, Zustand, IndexedDB | 3000 |
| AI | ChadGPT API (chadgpt.ru), Yandex SpeechKit | — |
| Платежи | YooMoney (ЮKassa) | — |

### Типы чатов (AI-роли)
| Ключ | Название |
|------|----------|
| `psychologist` | Психолог |
| `tarot` | Таролог |
| `sexologist` | Сексолог |
| `fortune_teller` | Гадалка |

### Тарифные планы
| Уровень | Модель AI | Цена/мес | Цена/год | Триал |
|---------|-----------|----------|----------|-------|
| `free` | gpt-4o-mini | 0 ₽ | — | — |
| `pro` | gpt-4o | 499 ₽ | 4 990 ₽ | 3 дня |
| `ultra` | gpt-4-turbo | 999 ₽ | 9 990 ₽ | 1 день |

---

## 2. Структура проекта

```
MentalChat/
├── Backend/                          # Go-бэкенд
│   ├── go.mod / go.sum
│   ├── mentalchat                    # Скомпилированный бинарник
│   ├── cmd/main.go                   # Точка входа: Gin-сервер, CORS, graceful shutdown
│   └── internal/
│       ├── config/config.go          # Конфигурация (Viper + singleton, .env.json)
│       ├── model/models.go           # GORM-модели (10 таблиц)
│       ├── routes/routes.go          # Маршруты API v1 + middleware
│       ├── handler/
│       │   ├── auth_handler.go       # Регистрация, логин, верификация email, сброс пароля
│       │   ├── chat_handler.go       # Отправка сообщений, история, поиск, синхронизация
│       │   ├── user_handler.go       # Профиль, настройки, сессии чатов
│       │   └── voice_handler.go      # Транскрибация голоса через Yandex SpeechKit
│       ├── middleware/middleware.go   # AuthMiddleware (JWT), RateLimiter, DDoSProtection, RequestLogger
│       ├── storage/storage.go        # PostgreSQL через GORM: CRUD для всех моделей + AutoMigrate
│       └── service/
│           ├── auth_service.go       # Регистрация, логин, bcrypt, email-верификация, fingerprint-проверка
│           ├── ai_service.go         # ChadGPT API (4 модели), промпты для 4 ролей. GetAIResponseWithContext
│           ├── chat_index_service.go # 🆕 Индексация: sequence numbers, SHA256-хеши, дедупликация
│           ├── context_service.go    # 🆕 Сборка контекста: управление окном токенов, авто-суммаризация
│           ├── jwt_service.go        # Access/Refresh JWT токены (HMAC-SHA256)
│           ├── email_service.go      # SMTP-отправка писем
│           ├── fingerprint_service.go # Генерация/валидация fingerprint, вычисление схожести
│           ├── payment_service.go    # YooMoney: создание платежа, проверка статуса, webhook
│           ├── user_service.go       # CRUD пользователей, обновление профиля
│           └── yandex_speechkit_service.go  # STT (Speech-to-Text) + TTS (Text-to-Speech)
│
├── Frontend/                         # React-фронтенд
│   ├── package.json                  # Vite, React 18, Zustand, TailwindCSS, idb, axios
│   ├── vite.config.ts                # Vite + proxy на :8080
│   ├── tsconfig.json
│   ├── tailwind.config.js
│   ├── postcss.config.js
│   ├── index.html
│   └── src/
│       ├── main.tsx                  # Точка входа React
│       ├── App.tsx                   # Роутинг: / → редирект, /login, /register, /dashboard
│       ├── index.css                 # TailwindCSS + кастомные стили
│       ├── pages/
│       │   ├── LoginPage.tsx         # Форма логина (email + пароль)
│       │   ├── RegistrationPage.tsx  # Регистрация (3 шага + fingerprint)
│       │   └── DashboardPage.tsx     # Основной интерфейс: сайдбар + чат + голосовой ввод + синхронизация
│       ├── components/
│       │   ├── VoiceInputButton.tsx  # Кнопка голосового ввода с визуализацией
│       │   └── WelcomeModal.tsx      # Приветственное модальное окно
│       ├── hooks/
│       │   ├── useChatSync.ts        # Хук автосинхронизации IndexedDB ↔ Server
│       │   ├── useFingerprint.ts     # Сбор отпечатка устройства (canvas, WebGL, шрифты...)
│       │   └── useSpeechRecognition.ts  # Web Speech API / Yandex SpeechKit
│       ├── services/
│       │   ├── api.ts                # Axios-клиент: auth, user, chat, payment, voice endpoints
│       │   └── chatSyncService.ts    # IndexedDB: локальное хранение + синхронизация
│       ├── store/
│       │   └── useAppStore.ts        # Zustand store (persist): user, auth, chat, theme, UI
│       ├── types/
│       │   └── speech.d.ts           # Типы для Web Speech API
│       └── utils/                    # (пустая директория)
│
├── docs/                             # Документация
│   ├── CHAT_SYNC.md                  # Система синхронизации чатов (IndexedDB)
│   ├── FINGERPRINTING.md             # Отпечатки устройств
│   ├── VOICE_INPUT.md                # Голосовой ввод
│   ├── YANDEX_QUICKSTART.md          # Быстрый старт Yandex SpeechKit
│   ├── YANDEX_SPEECHKIT_SETUP.md     # Настройка Yandex SpeechKit
│   └── TESTING_GUIDE.md              # Инструкция по тестированию
│
├── .env.json                         # Основной конфиг (в .gitignore)
├── .env.example.json                 # Шаблон конфигурации
├── build.sh                          # Скрипт сборки и запуска
├── quick_test.sh                     # Быстрый тест API
├── test_fingerprint.html             # HTML-страница тестирования fingerprint
├── test_fingerprinting.sh            # Скрипт тестирования fingerprinting
├── README.md                         # Основной README
├── README.md.old                     # Старая версия README
├── QUICK_START.md                    # Быстрый старт
├── FIX_DATABASE.md                   # Инструкция по исправлению БД
└── PROJECT_INDEX.md                  # ← Этот файл
```

---

## 3. Backend: Go API

### 3.1 Точка входа (`Backend/cmd/main.go`)

- Инициализация конфигурации (`config.LoadConfig()`)
- Подключение к PostgreSQL (`storage.NewPostgresDB()`)
- Создание сервисов: Auth, AI, Email, Fingerprint, Payment, User, JWT, YandexSpeechKit
- Настройка Gin с CORS
- Graceful shutdown с таймаутом 30с

### 3.2 Модели БД (`Backend/internal/model/models.go`)

| Модель | Таблица | Назначение |
|--------|---------|------------|
| `User` | `users` | Пользователи: email, пароль, tier, fingerprint, refresh_token |
| `Message` | `messages` | Сообщения чата: контент, роль, тип чата |
| `ChatSession` | `chat_sessions` | Сессии чатов: archived, last_message_time |
| `PaymentTransaction` | `payment_transactions` | Платежи YooMoney: статус, сумма, tier |
| `EmailLog` | `email_logs` | Лог отправки писем |
| `Subscription` | `subscriptions` | Подписки: статус, даты начала/окончания |
| `VoiceMessage` | `voice_messages` | Голосовые сообщения: путь к файлу, статус транскрибации |
| `UserCache` | `user_caches` | Кэш пользовательских данных |
| `RateLimit` | `rate_limits` | Ограничение частоты запросов по IP |
| `DDoSEntry` | `ddos_entries` | Защита от DDoS |

### 3.3 API Маршруты (`Backend/internal/routes/routes.go`)

| Метод | Путь | Хендлер | Middleware | Назначение |
|-------|------|---------|------------|------------|
| `POST` | `/api/v1/auth/register` | `Register` | — | Регистрация |
| `POST` | `/api/v1/auth/login` | `Login` | — | Логин |
| `POST` | `/api/v1/auth/refresh` | `RefreshToken` | — | Обновление токена |
| `POST` | `/api/v1/auth/verify-email` | `VerifyEmail` | — | Подтверждение email |
| `POST` | `/api/v1/auth/request-password-reset` | `RequestPasswordReset` | — | Запрос сброса пароля |
| `POST` | `/api/v1/auth/reset-password` | `ResetPassword` | — | Сброс пароля |
| `GET` | `/api/v1/config/trial` | `GetTrialInfo` | — | Информация о триалах |
| `GET` | `/api/v1/user/profile` | `GetProfile` | Auth | Профиль пользователя |
| `PUT` | `/api/v1/user/profile` | `UpdateProfile` | Auth | Обновление профиля |
| `DELETE` | `/api/v1/user/profile` | `DeleteProfile` | Auth | Удаление аккаунта |
| `GET` | `/api/v1/user/chat-sessions` | `GetChatSessions` | Auth | Сессии чатов |
| `POST` | `/api/v1/user/chat-sessions/archive` | `ArchiveChatSession` | Auth | Архивация сессии |
| `POST` | `/api/v1/user/microphone-permission` | `SaveMicrophonePermission` | Auth | Сохранение разрешения микрофона |
| `POST` | `/api/v1/chat` | `SendMessage` | Auth | Отправка сообщения AI |
| `POST` | `/api/v1/chat/history` | `GetHistory` | Auth | История сообщений |
| `POST` | `/api/v1/chat/search` | `SearchMessages` | Auth | Поиск по сообщениям |
| `POST` | `/api/v1/chat/sync` | `SyncMessages` | Auth | Push-синхронизация с проверкой хешей |
| `POST` | `/api/v1/chat/pull` | `PullMessages` | Auth | 🆕 Pull-синхронизация по sequence numbers |
| `POST` | `/api/v1/payment/initiate` | `InitiatePayment` | Auth | Создание платежа |
| `POST` | `/api/v1/payment/webhook` | `PaymentWebhook` | — | Webhook YooMoney |
| `POST` | `/api/v1/voice/transcribe` | `TranscribeVoice` | Auth | Транскрибация голоса |

### 3.4 Конфигурация (`Backend/internal/config/config.go`)

Singleton-паттерн через `sync.Once`. Читает `.env.json` через Viper. Секции конфигурации:

- **Server**: host, port, debug
- **Database**: host, port, name, user, password, ssl_mode
- **AI**: chad_api_url, models (free/pro/ultra), yandex_speechkit
- **Payment**: provider, yoomoney (shop_id, secret, scid), prices
- **Email**: SMTP (host, port, user, password, from)
- **Security**: JWT secret, rate limits, DDoS protection
- **Storage**: voice/avatar paths, max file size
- **App**: frontend URL, app name, support email

### 3.5 Хранилище (`Backend/internal/storage/storage.go`)

- PostgreSQL через GORM
- DSN формируется из конфигурации с fallback-дефолтами (localhost:5432)
- AutoMigrate всех 10 моделей
- CRUD операции для всех сущностей
- RateLimit / DDoS — in-memory singleton с `sync.Mutex`, скользящее окно, автоочистка

### 3.6 Сервисы

| Сервис | Файл | Ключевые функции |
|--------|------|------------------|
| **AuthService** | `auth_service.go` | RegisterUser, Login, VerifyEmail, ResetPassword, проверка fingerprint-схожести |
| **AIService** | `ai_service.go` | SendMessage — отправка в ChadGPT API, 4 роли специалистов с системными промптами |
| **JWTService** | `jwt_service.go` | Access токены (168ч), Refresh токены (720ч), HMAC-SHA256 |
| **EmailService** | `email_service.go` | SMTP отправка через Go `net/smtp` |
| **FingerprintService** | `fingerprint_service.go` | Генерация SHA-256 отпечатка, валидация, расчёт схожести, детекция подозрительных |
| **PaymentService** | `payment_service.go` | YooMoney: формирование формы оплаты, проверка подписи, webhook |
| **UserService** | `user_service.go` | Обновление профиля, управление подписками |
| **YandexSpeechKitService** | `yandex_speechkit_service.go` | STT (Speech-to-Text) и TTS (Text-to-Speech) через Yandex Cloud |
| **ChatIndexService** 🆕 | `chat_index_service.go` | IndexAndStoreMessage с SHA256-хешированием, sequence numbers, дедупликация |
| **ContextBuilder** 🆕 | `context_service.go` | BuildContext: скользящее окно токенов, авто-суммаризация длинных диалогов |

---

## 4. Frontend: React SPA

### 4.1 Страницы

| Страница | Файл | Маршрут | Назначение |
|----------|------|---------|------------|
| **LoginPage** | `LoginPage.tsx` | `/login` | Форма входа: email + пароль, ссылки на регистрацию/сброс пароля |
| **RegistrationPage** | `RegistrationPage.tsx` | `/register` | 3-шаговая регистрация: email/пароль → профиль → настройки (с fingerprint) |
| **DashboardPage** | `DashboardPage.tsx` | `/dashboard` | Основной интерфейс: сайдбар с чатами и настройками, область чата, голосовой ввод, синхронизация |

### 4.2 Компоненты

| Компонент | Файл | Назначение |
|-----------|------|------------|
| **VoiceInputButton** | `VoiceInputButton.tsx` | Кнопка записи голоса с анимацией, интеграция с Web Speech API / Yandex |
| **WelcomeModal** | `WelcomeModal.tsx` | Приветственное окно при первом входе |

### 4.3 Хуки

| Хук | Файл | Назначение |
|-----|------|------------|
| **useChatSync** | `useChatSync.ts` | Автоматическая фоновая синхронизация IndexedDB → Server с настраиваемым интервалом |
| **useFingerprint** | `useFingerprint.ts` | Сбор отпечатка устройства (canvas, WebGL, шрифты, плагины, WebRTC IP...) |
| **useSpeechRecognition** | `useSpeechRecognition.ts` | Распознавание речи: Web Speech API (браузер) или Yandex SpeechKit |

### 4.4 Сервисы

| Сервис | Файл | Назначение |
|--------|------|------------|
| **api** | `api.ts` | Axios-клиент с интерсепторами (JWT в заголовках, редирект при 401). Экспортирует: `authApi`, `userApi`, `chatApi`, `paymentApi`, `voiceApi` |
| **chatSyncService** | `chatSyncService.ts` | IndexedDB-сервис: сохранение сообщений локально, автосинхронизация, статистика, логирование |

### 4.5 Управление состоянием (Zustand)

Файл: `src/store/useAppStore.ts`

Хранилище с персистентностью в localStorage (`mentalchat-storage`):

- **User**: id, email, displayName, mentalState, tier, verified, avatarURL
- **Auth**: isAuthenticated, token (JWT)
- **Chat**: currentChatType, chatHistory[]
- **Theme**: name, primaryColor, secondaryColor, accentColor (default: Sage Green)
- **Settings**: microphonePerm, marketingEmail, privacyPolicy
- **UI**: sidebarOpen, welcomeModalOpen

Персистятся: user, theme, microphonePerm, marketingEmail, privacyPolicy, sidebarOpen

### 4.6 Типы чатов

```typescript
type ChatType = 'psychologist' | 'tarot' | 'sexologist' | 'fortune_teller'
```

### 4.7 Система синхронизации чатов

Асинхронная синхронизация IndexedDB ↔ Backend:

1. Сообщения сохраняются **локально** в IndexedDB (БД: `MentalChatDB`, stores: `chats`, `messages`, `syncLog`)
2. Фоновая синхронизация с интервалом (по умолчанию 300с, настраивается через `localStorage`)
3. Ручная синхронизация через кнопку "Синхронизировать сейчас"
4. Защита от дубликатов по `localId`
5. Индикатор неотсинхронизированных сообщений в сайдбаре

---

## 5. Ключевые архитектурные решения

### 5.1 Аутентификация
- JWT (Access + Refresh токены)
- Access: 168 часов (7 дней), Refresh: 720 часов (30 дней)
- Хранение refresh-токена в БД (поле `refresh_token` в users)
- Middleware извлекает user_id, email, tier из claims

### 5.2 Fingerprinting
- Сбор 15+ параметров устройства на фронтенде
- SHA-256 хеширование
- Отправка при регистрации
- Серверная проверка схожести (>80% = подозрительно)
- Детекция ботов (пустой UA, нереальные параметры экрана)

### 5.3 AI Integration
- Единый endpoint ChadGPT API: `https://ask.chadgpt.ru/api/public`
- 3 модели: gpt-4o-mini (free), gpt-4o (pro), gpt-4-turbo (ultra)
- 4 системных промпта для каждой роли специалиста
- Роли: психолог, таролог, сексолог, гадалка

### 5.4 Голосовой ввод
- Web Speech API (браузерный) как основной
- Yandex SpeechKit как fallback/альтернатива
- Отправка аудио → транскрибация → вставка текста в поле ввода

### 5.5 Платежи
- YooMoney (ЮKassa) через HTML-форму
- Webhook для подтверждения платежей
- Подписки: active/expired/cancelled

---

## 6. База данных

**СУБД:** PostgreSQL 15+
**База:** `mentalchat`
**Пользователь:** `mentalchat` / `mentalchat`

### Схема (GORM AutoMigrate)

```sql
-- Основные таблицы
users (id, email, password, display_name, mental_state, tier, verified,
       email_token, email_token_exp, fingerprint, fingerprint_data,
       trial_start, trial_end, refresh_token, refresh_token_exp, ...)
messages (id, user_id, chat_type, content, is_from_ai, role, timestamp, ...)
chat_sessions (id, user_id, chat_type, last_message_time, archived, ...)
payment_transactions (id, user_id, transaction_id, amount, currency, status, tier, ...)
email_logs (id, user_id, email_type, recipient, subject, status, ...)
subscriptions (id, user_id, tier, status, start_date, end_date, ...)
voice_messages (id, user_id, message_id, file_path, status, transcript, ...)
user_caches (id, user_id, cache_key, cache_value, ...)
rate_limits (id, ip, request_count, window_start, ...)
ddos_entries (id, ip, request_count, locked_until, ...)
```

**Важно:** DSN формируется из `config.DatabaseConfig` через `buildDSN()` с дефолтными значениями для локальной разработки.

---

## 7. Запуск проекта

```bash
# Быстрый старт
./build.sh              # Сборка backend + frontend и запуск
./build.sh --build      # Только сборка
./build.sh --run        # Только запуск

# Или вручную:
# Backend
cd Backend && go run ./cmd/main.go

# Frontend
cd Frontend && npm run dev
```

**Требования:**
- Go 1.21+
- Node.js 18+
- PostgreSQL 15+ (с базой `mentalchat`)
- `.env.json` в корне проекта

---

## 8. Технический долг / TODO

1. ~~**DSN захардкожен** — `storage.go:22` не использует конфигурацию из `.env.json`~~ ✅ исправлено (buildDSN)
2. ~~**RateLimit / DDoS — заглушки** — всегда разрешают запросы~~ ✅ исправлено (in-memory singleton)
3. ~~**generateRandomString** — возвращает `"placeholder-token"`, не использует `crypto/rand`~~ ✅ исправлено
4. ~~**ResetPassword** — вызывает `GetUserByEmail("")` с пустым email (баг)~~ ✅ исправлено (GetUserByEmailToken)
5. **GetMessageByLocalID** — не реализован поиск по local_id, возвращает первое сообщение пользователя
6. **Yandex SpeechKit** — отключен по умолчанию (`enabled: false`)
7. **Тесты** — отсутствуют (есть только `quick_test.sh` и `test_fingerprinting.sh`)
8. **Docker** — нет Dockerfile/docker-compose
9. **README.md.old** — устаревшая версия, можно удалить
10. **Frontend** — `DashboardPage.tsx` содержит прямые fetch-запросы вместо использования `chatApi` из `api.ts`

---

## 9. Полезные команды

```bash
# Backend
cd Backend && go mod tidy && go build -o mentalchat ./cmd/main.go
cd Backend && go run ./cmd/main.go

# Frontend
cd Frontend && npm install && npm run dev
cd Frontend && npm run build

# Тесты
./quick_test.sh           # Быстрый тест API (регистрация + чат)
./test_fingerprinting.sh  # Тест fingerprinting
```

---

## 10. Ресурсы и ссылки

- **ChadGPT API:** https://ask.chadgpt.ru/api/public
- **Yandex SpeechKit:** https://stt.api.cloud.yandex.net
- **YooMoney (ЮKassa):** https://yookassa.ru
- **Внутренняя документация:** `docs/` (6 файлов)