# 🧪 Руководство по тестированию Fingerprinting

## Быстрый старт

### 1. Автоматическая проверка

```bash
cd /home/n1/Projects/MentalChat
./test_fingerprinting.sh
```

**Что проверяет:**
- ✅ Наличие всех файлов fingerprinting
- ✅ Сборку backend и frontend
- ✅ Импорты и интеграцию
- ✅ Поля в моделях данных

---

### 2. Тест в браузере (без БД)

Откройте файл `test_fingerprint.html` в браузере:

```bash
# Вариант 1: Просто открыть файл
xdg-open test_fingerprint.html

# Вариант 2: Через локальный сервер
cd MentalChat
python3 -m http.server 8000
# Откройте http://localhost:8000/test_fingerprint.html
```

**Что проверяет:**
- ✅ Генерацию fingerprint
- ✅ Стабильность (одинаковый ли при повторной генерации)
- ✅ Сбор всех данных (Canvas, WebGL, Fonts, Plugins)
- ✅ Отправку на сервер (если backend запущен)

**Ожидаемый результат:**
- Fingerprint генерируется за ~1 секунду
- При повторной генерации fingerprint **одинаковый**
- Все хеши (Canvas, WebGL, Fonts, Plugins) заполнены

---

### 3. Полное тестирование (с БД)

#### Шаг 1: Настройте PostgreSQL

```bash
# Создайте базу данных
sudo -u postgres psql

CREATE DATABASE mentalchat;
CREATE USER mentalchat_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE mentalchat TO mentalchat_user;
\q
```

#### Шаг 2: Настройте .env.json

```json
{
  "database": {
    "host": "localhost",
    "port": 5432,
    "name": "mentalchat",
    "user": "mentalchat_user",
    "password": "your_password",
    "ssl_mode": "disable"
  }
}
```

#### Шаг 3: Запустите backend

```bash
cd Backend
./mentalchat
```

**Ожидаемый лог:**
```
INFO  Starting server address=localhost:8080
INFO  Yandex SpeechKit initialized successfully
```

#### Шаг 4: Запустите frontend

```bash
cd Frontend
npm run dev
```

Откройте http://localhost:3000/registration

#### Шаг 5: Проверка в браузере

1. Откройте консоль разработчика (F12)
2. Перейдите на вкладку **Console**
3. Должен появиться лог: `Fingerprint generated: abc123...`

4. Перейдите на вкладку **Network**
5. Заполните форму регистрации
6. Нажмите "Регистрация"
7. Найдите запрос к `/api/v1/auth/register`
8. Проверьте payload - должен содержать `fingerprint` и `fingerprint_data`

**Пример payload:**
```json
{
  "email": "test@example.com",
  "password": "password123",
  "display_name": "Test",
  "fingerprint": "abc123def456...",
  "fingerprint_data": {
    "ua": "Mozilla/5.0...",
    "sw": 1920,
    "sh": 1080,
    "ch": "canvas_hash",
    "wh": "webgl_hash",
    ...
  }
}
```

#### Шаг 6: Проверка в БД

```bash
sudo -u postgres psql mentalchat

# Проверка сохраненного fingerprint
SELECT id, email, fingerprint, trial_start, trial_end 
FROM users 
ORDER BY created_at DESC 
LIMIT 1;

# Проверка fingerprint_data
SELECT fingerprint_data FROM users 
WHERE email = 'test@example.com';
```

**Ожидаемый результат:**
- fingerprint сохранен (не пустой)
- fingerprint_data содержит JSON с данными устройства
- trial_end установлен через 3 дня от trial_start

---

### 4. Тест на повторную регистрацию

#### Цель: Проверить блокировку повторного trial

**Шаг 1:** Зарегистрируйтесь с первого устройства

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test1@example.com",
    "password": "password123",
    "display_name": "Test 1",
    "mental_state": "harmony",
    "marketing_email": false,
    "privacy_policy": true,
    "fingerprint": "test_fingerprint_123",
    "fingerprint_data": {"test": true}
  }'
```

**Ожидаемый ответ:** `201 Created`

**Шаг 2:** Попробуйте зарегистрироваться снова с тем же fingerprint

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test2@example.com",
    "password": "password123",
    "display_name": "Test 2",
    "mental_state": "harmony",
    "marketing_email": false,
    "privacy_policy": true,
    "fingerprint": "test_fingerprint_123",
    "fingerprint_data": {"test": true}
  }'
```

**Ожидаемый ответ:** `400 Bad Request`
```json
{
  "error": "trial period already used for this device"
}
```

---

### 5. Тест на подозрительную активность

#### Цель: Проверка обнаружения ботов

**Тест 1: Пустой User Agent**

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -H "User-Agent: " \
  -d '{
    "email": "bot@example.com",
    "password": "password123",
    "display_name": "Bot",
    "fingerprint": "",
    "fingerprint_data": {}
  }'
```

**Ожидаемый результат:** Логирование подозрительной активности

**Тест 2: Bot в User Agent**

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -H "User-Agent: Googlebot/2.1" \
  -d '{
    "email": "bot2@example.com",
    "password": "password123",
    "display_name": "Bot 2",
    "fingerprint": "",
    "fingerprint_data": {}
  }'
```

**Ожидаемый результат:** Логирование "Suspicious fingerprint: bot detected"

---

### 6. Проверка логов

```bash
# Логи backend
tail -f /tmp/backend.log | grep -i fingerprint

# Ожидаемые сообщения:
# - "Generated fingerprint: abc123..."
# - "User registered successfully"
# - "Suspicious registration detected" (если подозрительно)
```

---

## Чек-лист успешного тестирования

- [ ] Fingerprint генерируется в браузере
- [ ] Fingerprint стабильный (одинаковый при повторной генерации)
- [ ] Все компоненты собираются (Canvas, WebGL, Fonts, Plugins)
- [ ] Fingerprint отправляется при регистрации
- [ ] Fingerprint сохраняется в БД
- [ ] Повторная регистрация с тем же fingerprint блокируется
- [ ] Подозрительные активности логируются
- [ ] Документация читаемая и полная

---

## Отладка

### Fingerprint не генерируется

**Проблема:** Ошибка в консоли браузера

**Решение:**
1. Проверьте поддержку WebRTC
2. Включите JavaScript
3. Проверьте консоль на ошибки

### Fingerprint меняется

**Проблема:** Нестабильный fingerprint

**Возможные причины:**
- Изменение размера окна (используйте screen.width, не window.innerWidth)
- Динамические шрифты
- WebGL рендеринг меняется

**Решение:** Проверьте `useFingerprint.ts` - все ли данные стабильные

### БД не сохраняет fingerprint

**Проблема:** Поле fingerprint пустое в БД

**Решение:**
1. Проверьте миграции БД
2. Убедитесь, что поля есть в модели
3. Проверьте логи backend

### Повторная регистрация не блокируется

**Проблема:** Можно создать много аккаунтов

**Решение:**
1. Проверьте `GetUserByFingerprint()` в storage
2. Убедитесь, что проверка есть в `RegisterUser()`
3. Проверьте логи на наличие ошибок

---

## Инструменты для тестирования

### 1. Browser DevTools

**Console:**
```javascript
// Проверка fingerprint
const { fingerprint, fingerprintData } = window.testFingerprint
console.log(fingerprint)
```

**Network:**
- Проверка запросов к API
- Анализ payload

### 2. PostgreSQL

```sql
-- Все пользователи с fingerprint
SELECT email, fingerprint, trial_end 
FROM users 
WHERE fingerprint IS NOT NULL;

-- Поиск дубликатов fingerprint
SELECT fingerprint, COUNT(*) as count
FROM users
WHERE fingerprint IS NOT NULL
GROUP BY fingerprint
HAVING COUNT(*) > 1;

-- Статистика по trial
SELECT 
  COUNT(*) as total_users,
  COUNT(fingerprint) as with_fingerprint,
  AVG(EXTRACT(EPOCH FROM (trial_end - trial_start))/86400) as avg_trial_days
FROM users;
```

### 3. curl для API тестов

См. раздел "Тест на повторную регистрацию" выше.

---

## Поддержка

При возникновении проблем:

1. Проверьте логи backend
2. Проверьте консоль браузера
3. Убедитесь, что БД доступна
4. Проверьте .env.json конфигурацию

**Документация:**
- [FINGERPRINTING.md](docs/FINGERPRINTING.md) - полное руководство
- [API.md](docs/API.md) - API документация

---

**Успешного тестирования!** 🧪✅
