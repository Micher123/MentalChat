# MentalChat 🌸

**Персональный AI-помощник для ментального здоровья женщин**

MentalChat - это инновационный сервис, который заменяет психолога, таролога, сексолога и гадалку, помогая женщинам в их ежедневных вопросах, связанных с ментальной составляющей жизни.

## ✨ Возможности

### Для пользователей:
- 🧠 **4 типа специалистов**: Психолог, Таролог, Сексолог, Гадалка
- 💎 **3 тарифа**: FREE, PRO (3 дня trial), ULTRA (1 день trial)
- 🎨 **7 тем оформления**: Sage Green, Dusty Rose, Light Lilac, Storm Blue, French Gray, Almond, Oat
- 🎙️ **Голосовой ввод** сообщений с Web Speech API
- 🔍 **Поиск по истории** чатов
- 📧 **Email уведомления** и рассылки
- 🔐 **Безопасная аутентификация** с JWT токенами

### Технические особенности:
- ⚡ **Быстрый бэкенд** на Go (Gin + GORM + PostgreSQL)
- 🚀 **Современный фронтенд** на React + TypeScript + Vite
- 🎭 **Liquid Glass дизайн** с полупрозрачными элементами
- 🛡️ **Защита от DDoS** и Rate Limiting
- 📱 **Адаптивный дизайн** для всех устройств

## 🏗️ Архитектура проекта

```
MentalChat/
├── Backend/                 # Go бэкенд
│   ├── cmd/
│   │   └── main.go         # Точка входа
│   ├── internal/
│   │   ├── config/         # Конфигурация
│   │   ├── handler/        # HTTP обработчики
│   │   ├── middleware/     # Middleware (auth, rate limit, DDoS)
│   │   ├── model/          # Модели данных
│   │   ├── routes/         # Роутинг
│   │   ├── service/        # Бизнес-логика
│   │   │   ├── ai_service.go       # AI интеграция
│   │   │   ├── auth_service.go     # Аутентификация
│   │   │   ├── email_service.go    # Email рассылки
│   │   │   ├── jwt_service.go      # JWT токены
│   │   │   ├── payment_service.go  # Платежи
│   │   │   └── user_service.go     # Пользователи
│   │   ├── storage/        # Работа с БД
│   │   └── utils/          # Утилиты
│   └── go.mod
├── Frontend/               # React фронтенд
│   ├── src/
│   │   ├── components/     # UI компоненты
│   │   ├── pages/          # Страницы
│   │   ├── services/       # API клиент
│   │   ├── store/          # Zustand store
│   │   └── styles/         # Стили
│   └── package.json
├── .env.example.json       # Пример конфигурации
├── build.sh                # Скрипт сборки
└── README.md
```

## 🚀 Быстрый старт

### Требования:
- Go 1.22+
- Node.js 18+
- PostgreSQL 14+
- Git

### 1. Клонирование репозитория:
```bash
git clone <repository-url>
cd MentalChat
```

### 2. Настройка конфигурации:
```bash
cp .env.example.json .env.json
# Отредактируйте .env.json с вашими настройками
```

### 3. Запуск через build.sh:
```bash
chmod +x build.sh
./build.sh --build --run
```

Или по отдельности:
```bash
# Только сборка
./build.sh --build

# Только запуск (после сборки)
./build.sh --run
```

### 4. Ручной запуск:

**Бэкенд:**
```bash
cd Backend
go mod tidy
go run ./cmd/main.go
```

**Фронтенд:**
```bash
cd Frontend
npm install
npm run dev
```

## ⚙️ Конфигурация (.env.json)

### Основные параметры:

```json
{
  "server": {
    "host": "localhost",
    "port": 8080,
    "debug": false
  },
  "database": {
    "host": "localhost",
    "port": 5432,
    "name": "mentalchat",
    "user": "mentalchat_user",
    "password": "your_password",
    "ssl_mode": "disable"
  },
  "ai": {
    "chad_api_url": "https://ask.chadgpt.ru/api/public",
    "models": {
      "free": "gpt-4o-mini",
      "pro": "gpt-4o",
      "ultra": "gpt-4-turbo"
    },
    "yandex_speechkit": {
      "enabled": true,
      "folder_id": "your_yandex_folder_id",
      "iam_token": "your_yandex_iam_token",
      "service_key": "your_yandex_oauth_token",
      "api_endpoint": "https://stt.api.cloud.yandex.net"
    }
  },
  "payment": {
    "provider": "yoomoney",
    "yoomoney_shop_id": "your_shop_id",
    "yoomoney_secret": "your_secret",
    "prices": {
      "pro_monthly": 499,
      "pro_yearly": 4990,
      "ultra_monthly": 999,
      "ultra_yearly": 9990
    }
  },
  "email": {
    "smtp_host": "smtp.example.com",
    "smtp_port": 587,
    "smtp_user": "noreply@example.com",
    "smtp_password": "your_password",
    "from_email": "noreply@example.com",
    "from_name": "MentalChat"
  },
  "security": {
    "jwt_secret": "your_super_secret_key",
    "jwt_expiration_hours": 168,
    "rate_limit_requests": 100,
    "rate_limit_window_seconds": 60,
    "ddos_protection_enabled": true,
    "max_concurrent_requests_per_ip": 50
  },
  "storage": {
    "voice_storage_path": "./storage/voices",
    "avatar_storage_path": "./storage/avatars",
    "max_file_size_mb": 10
  },
  "app": {
    "frontend_url": "http://localhost:3000",
    "app_name": "MentalChat",
    "support_email": "support@mentalchat.com"
  }
}
```

## 🔐 Безопасность

### Реализованные механизмы:
- ✅ **JWT аутентификация** с access и refresh токенами
- ✅ **Bcrypt хеширование** паролей
- ✅ **Rate Limiting** для защиты от злоупотреблений
- ✅ **DDoS защита** с автоматической блокировкой IP
- ✅ **Email верификация** пользователей
- ✅ **Защита от повторных trial периодов** (fingerprinting)
- ✅ **CORS** и **Secure Headers**

## 📡 API Endpoints

### Публичные маршруты:
- `POST /api/v1/auth/register` - Регистрация
- `POST /api/v1/auth/login` - Вход
- `POST /api/v1/auth/verify-email` - Подтверждение email
- `POST /api/v1/auth/refresh` - Обновление токена
- `POST /api/v1/auth/request-password-reset` - Запрос сброса пароля
- `POST /api/v1/auth/reset-password` - Сброс пароля
- `GET /api/v1/config/trial` - Информация о тарифах

### Защищенные маршруты (требуется JWT):
- `POST /api/v1/auth/logout` - Выход
- `GET /api/v1/user/profile` - Получить профиль
- `PUT /api/v1/user/profile` - Обновить профиль
- `DELETE /api/v1/user/profile` - Удалить профиль
- `GET /api/v1/user/chat-sessions` - История чатов
- `POST /api/v1/user/microphone-permission` - Сохранение разрешения на микрофон
- `POST /api/v1/chat/` - Отправить сообщение
- `POST /api/v1/chat/history` - История сообщений
- `POST /api/v1/chat/search` - Поиск по сообщениям
- `POST /api/v1/payment/initiate` - Инициировать платеж
- `POST /api/v1/voice/transcribe` - Транскрибация голоса

## 🎨 Темы оформления

Доступные цветовые схемы:
1. **Sage Green** (#8FBC8F)
2. **Dusty Rose** (#CAC0B5)
3. **Light Lilac** (#C8A2C6)
4. **Storm Blue** (#708090)
5. **French Gray** (#BEBEBE)
6. **Almond** (#EEDC82)
7. **Oat** (#F5F5DC)

## 🧪 Тестирование

### Бэкенд:
```bash
cd Backend
go test ./...
```

### Фронтенд:
```bash
cd Frontend
npm test
```

## 📦 Сборка для продакшена

```bash
./build.sh --build
```

После сборки:
- Бэкенд: `./mentalchat-backend`
- Фронтенд: `Frontend/dist/`

## 🔧 Технологии

### Бэкенд:
- **Go 1.22** - язык программирования
- **Gin** - HTTP фреймворк
- **GORM** - ORM для PostgreSQL
- **PostgreSQL** - база данных
- **JWT** - аутентификация
- **Viper** - конфигурация
- **Zerolog** - логирование

### Фронтенд:
- **React 18** - UI библиотека
- **TypeScript** - типизация
- **Vite** - сборщик
- **TailwindCSS** - стилизация
- **Zustand** - state management
- **Framer Motion** - анимации
- **React Router** - роутинг
- **Axios** - HTTP клиент

## 📝 Roadmap

### Приоритет 1 (MVP) ✅:
- [x] JWT аутентификация
- [x] Email верификация
- [x] Базовая интеграция с AI
- [x] Восстановление пароля
- [x] Платежная система
- [x] Trial периоды

### Приоритет 2 (В РЕАЛИЗАЦИИ ✅):
- [x] Голосовой ввод (Web Speech API)
- [x] Yandex SpeechKit интеграция
- [x] Страница профиля с настройками
- [x] Полное переключение тем
- [x] Поиск по истории чатов (UI готов)
- [x] Анимация лотоса при входе
- [ ] Fingerprinting для защиты trial

### Приоритет 3:
- [ ] IndexedDB для кэширования
- [ ] Service Worker для офлайн работы
- [ ] Маркетинговые рассылки
- [ ] Улучшенная DDoS защита (Redis)
- [ ] Аналитика и мониторинг
- [ ] Мобильное приложение

## 📚 Документация

- [Настройка Yandex SpeechKit](docs/YANDEX_SPEECHKIT_SETUP.md) - полная инструкция по интеграции
- [Голосовой ввод](docs/VOICE_INPUT.md) - архитектура и API голосового ввода

## 🤝 Вклад в проект

1. Fork репозитория
2. Создайте feature branch (`git checkout -b feature/amazing-feature`)
3. Commit изменения (`git commit -m 'Add amazing feature'`)
4. Push в branch (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

## 📄 Лицензия

Этот проект является собственностью NLP-Core-Team. Все права защищены.

## 📞 Контакты

- **Email**: support@mentalchat.com
- **Telegram**: @mentalchat_support

---

**MentalChat** - Твоя красота исходит изнутри! 🌸
