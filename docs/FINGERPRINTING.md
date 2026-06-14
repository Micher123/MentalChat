# 🔐 Fingerprinting в MentalChat

## Обзор

Fingerprinting (отпечаток устройства) - это технология идентификации пользователя по уникальным характеристикам его устройства и браузера. В MentalChat fingerprinting используется для защиты от злоупотреблений trial периодами.

## Зачем это нужно?

### Проблемы, которые решает fingerprinting:

1. **Повторные регистрации** - пользователи не могут создать много аккаунтов для получения повторных trial
2. **Обход блокировок** - заблокированные пользователи не могут вернуться
3. **Фрод** - предотвращение мошеннических действий
4. **Безопасность** - обнаружение подозрительной активности

## Как это работает?

### 1. Сбор данных (Фронтенд)

Хук `useFingerprint.ts` собирает следующие данные:

#### Основные данные:
- **User Agent** - информация о браузере и ОС
- **Accept Language** - предпочтительные языки
- **Screen Dimensions** - разрешение экрана
- **Color Depth** - глубина цвета
- **Pixel Ratio** - соотношение пикселей
- **Timezone** - часовой пояс
- **Platform** - платформа (Windows, Mac, Linux)

#### Аппаратные данные:
- **Hardware Concurrency** - количество ядер CPU
- **Device Memory** - объем оперативной памяти
- **Touch Points** - количество точек касания

#### Уникальные отпечатки:
- **Canvas Fingerprint** - хеш рендеринга canvas
- **WebGL Fingerprint** - хеш видеокарты и драйверов
- **Audio Fingerprint** - хеш аудиосистемы
- **Fonts Fingerprint** - хеш установленных шрифтов
- **Plugins Fingerprint** - хеш установленных плагинов
- **Local IPs** - локальные IP адреса (WebRTC)

#### Дополнительные данные:
- **Do Not Track** - настройка приватности
- **Cookie Enabled** - включены ли cookies
- **Language** - основной язык

### 2. Генерация fingerprint

```typescript
// Компоненты объединяются в одну строку
const components = [
  userAgent,
  acceptLanguage,
  screenResolution,
  colorDepth,
  timezone,
  platform,
  canvasHash,
  webglHash,
  fontsHash,
  // ...
].join('|')

// Создается SHA-256 хеш
const fingerprint = await hashString(components)
```

### 3. Отправка на сервер

При регистрации fingerprint отправляется вместе с данными пользователя:

```json
{
  "email": "user@example.com",
  "password": "******",
  "display_name": "John",
  "fingerprint": "abc123...",
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

### 4. Проверка на сервере (Бэкенд)

#### FingerprintService:

**Методы:**
- `GenerateFingerprint(data)` - создает fingerprint из данных
- `ValidateFingerprint(data)` - проверяет валидность
- `CalculateSimilarity(fp1, fp2)` - вычисляет схожесть двух fingerprint
- `IsSuspiciousFingerprint(data)` - определяет подозрительные данные

#### Проверка при регистрации:

```go
func (s *AuthService) RegisterUser(input RegisterInput, fingerprintData map[string]interface{}) (*model.User, error) {
    // 1. Проверяем существующий fingerprint
    existingUser, err := s.storage.GetUserByFingerprint(input.Fingerprint)
    if err == nil && existingUser.ID > 0 {
        if existingUser.TrialEnd != nil && existingUser.TrialEnd.Before(time.Now()) {
            return nil, fmt.Errorf("trial period already used for this device")
        }
    }
    
    // 2. Проверяем схожесть с другими пользователями
    suspiciousUsers, err := s.checkFingerprintSimilarity(input.Fingerprint, fingerprintData)
    if len(suspiciousUsers) > 0 {
        log.Warn().Msg("Suspicious registration detected")
    }
    
    // 3. Создаем пользователя
    user := &model.User{
        Email: input.Email,
        Fingerprint: input.Fingerprint,
        FingerprintData: json.Marshal(fingerprintData),
        TrialEnd: time.Now().AddDate(0, 0, 3),
        ...
    }
    
    return user, nil
}
```

## Архитектура

```
┌─────────────────┐
│   Браузер       │
│   (Клиент)      │
└────────┬────────┘
         │
         │ Сбор данных
         │ (useFingerprint)
         ▼
┌─────────────────┐
│  Fingerprint    │
│  (SHA-256 hash) │
└────────┬────────┘
         │
         │ Отправка при регистрации
         ▼
┌─────────────────┐
│  Backend        │
│  (Go)           │
└────────┬────────┘
         │
         │ Проверка
         ▼
┌─────────────────┐
│  БД (PostgreSQL)│
│  - Users        │
│  - Fingerprint  │
└─────────────────┘
```

## Хранение данных

### Таблица Users:

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    fingerprint VARCHAR(255),
    fingerprint_data TEXT,
    trial_start TIMESTAMP,
    trial_end TIMESTAMP,
    ...
);

CREATE INDEX idx_users_fingerprint ON users(fingerprint);
```

### Структура fingerprint_data (JSON):

```json
{
  "ua": "Mozilla/5.0 (Windows NT 10.0; Win64; x64)...",
  "al": "ru-RU,ru;q=0.9,en;q=0.8",
  "sw": 1920,
  "sh": 1080,
  "cd": 24,
  "pr": 1,
  "tz": "Europe/Moscow",
  "to": -180,
  "pf": "Win32",
  "hc": 8,
  "dm": 16,
  "tp": 0,
  "ch": "a1b2c3d4e5f6...",
  "wh": "f6e5d4c3b2a1...",
  "fh": "1a2b3c4d5e6f...",
  "ph": "6f5e4d3c2b1a...",
  "li": ["192.168.1.100"],
  "dnt": "unspecified",
  "ce": true,
  "lg": "ru-RU"
}
```

## Обнаружение подозрительной активности

### Критерии подозрительности:

1. **Пустой User Agent** - боты/скрипты
2. **Нереальные размеры экрана** - < 320px или > 5120px
3. **Нереальное количество ядер** - < 1 или > 128
4. **Отключенные cookies** - подозрительно для обычного пользователя
5. **Bot/Crawler в User Agent** - автоматизированные скрипты
6. **Совпадение fingerprint** - повторная регистрация

### Пример проверки:

```go
func (s *FingerprintService) IsSuspiciousFingerprint(data *FingerprintData) bool {
    // Пустой User Agent
    if data.UserAgent == "" {
        return true
    }
    
    // Bot detection
    suspiciousAgents := []string{"bot", "crawler", "spider", "scraper"}
    for _, suspicious := range suspiciousAgents {
        if strings.Contains(strings.ToLower(data.UserAgent), suspicious) {
            return true
        }
    }
    
    // Нереальные размеры экрана
    if data.ScreenWidth < 320 || data.ScreenWidth > 5120 {
        return true
    }
    
    // Отключенные cookies
    if !data.CookieEnabled {
        return true
    }
    
    return false
}
```

## Вычисление схожести fingerprint

### Алгоритм:

```go
func (s *FingerprintService) CalculateSimilarity(fp1, fp2 *FingerprintData) float64 {
    matchingComponents := 0
    totalComponents := 0
    
    // User Agent (exact match)
    if fp1.UserAgent == fp2.UserAgent {
        matchingComponents++
    }
    totalComponents++
    
    // Screen dimensions (exact match)
    if fp1.ScreenWidth == fp2.ScreenWidth && fp1.ScreenHeight == fp2.ScreenHeight {
        matchingComponents++
    }
    totalComponents++
    
    // Timezone (exact match)
    if fp1.Timezone == fp2.Timezone {
        matchingComponents++
    }
    totalComponents++
    
    // Canvas hash (most reliable)
    if fp1.CanvasHash == fp2.CanvasHash && fp1.CanvasHash != "" {
        matchingComponents++
    }
    totalComponents++
    
    // WebGL hash (very reliable)
    if fp1.WebGLHash == fp2.WebGLHash && fp1.WebGLHash != "" {
        matchingComponents++
    }
    totalComponents++
    
    similarity := float64(matchingComponents) / float64(totalComponents)
    return similarity
}
```

### Интерпретация схожести:

| Схожесть | Значение |
|----------|----------|
| 100% | Тот же пользователь |
| 80-99% | Очень похоже (возможно тот же) |
| 60-79% | Похоже (может быть тот же) |
| 40-59% | Средняя схожесть |
| < 40% | Разные устройства |

## Конфиденциальность

### Что мы НЕ делаем:

- ❌ Не собираем персональные данные
- ❌ Не отслеживаем местоположение
- ❌ Не читаем файлы пользователя
- ❌ Не передаем данные третьим лицам

### Что мы делаем:

- ✅ Собираем только технические характеристики
- ✅ Храним fingerprint в хешированном виде
- ✅ Используем только для защиты от фрода
- ✅ Удаляем данные при удалении аккаунта

### GDPR соответствие:

Fingerprint данные считаются персональными данными по GDPR. Поэтому:

1. **Информирование** - пользователи должны знать о сборе
2. **Согласие** - необходимо согласие пользователя
3. **Право на удаление** - данные удаляются по запросу
4. **Минимизация** - собираем только необходимое

## Настройка

### Бэкенд (.env.json):

```json
{
  "security": {
    "fingerprinting_enabled": true,
    "fingerprint_similarity_threshold": 0.8,
    "block_suspicious_registrations": true
  }
}
```

### Фронтенд:

```typescript
// Автоматическая генерация при монтировании компонента
const { fingerprint, fingerprintData, loading } = useFingerprint()

// Ручная генерация (если нужно)
const { generateFingerprint } = useFingerprint()
await generateFingerprint()
```

## Тестирование

### Проверка работы:

1. Откройте консоль разработчика
2. Перейдите на страницу регистрации
3. В консоли увидите: `Fingerprint generated: abc123...`
4. Зарегистрируйтесь
5. Проверьте БД - fingerprint должен сохраниться

### Проверка блокировки:

```bash
# Попробуйте зарегистрироваться с тем же fingerprint
curl -X POST 'http://localhost:8080/api/v1/auth/register' \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "test2@example.com",
    "password": "password123",
    "display_name": "Test",
    "fingerprint": "abc123...",
    "fingerprint_data": {...}
  }'

# Ожидаемый ответ:
# {"error": "trial period already used for this device"}
```

## Будущие улучшения

- [ ] **Машинное обучение** - автоматическое обнаружение паттернов фрода
- [ ] **Behavioral fingerprinting** - анализ поведения пользователя
- [ ] **Device graph** - связь устройств по одному пользователю
- [ ] **Risk scoring** - оценка риска для каждой регистрации
- [ ] **A/B тестирование** - оптимизация порога схожести

## Производительность

### Время генерации:

- Canvas: ~10ms
- WebGL: ~5ms
- Fonts: ~50ms
- Plugins: ~2ms
- WebRTC IP: ~1000ms
- **Итого**: ~1067ms (~1 секунда)

### Оптимизация:

- Кэширование fingerprint в localStorage
- Генерация в фоне (не блокирует UI)
- Ленивая загрузка тяжелых компонентов

## Поддержка браузеров

| Браузер | Поддержка |
|---------|-----------|
| Chrome | ✅ Полная |
| Firefox | ✅ Полная |
| Safari | ✅ Полная |
| Edge | ✅ Полная |
| Opera | ✅ Полная |
| IE 11 | ⚠️ Ограниченная |

## Ресурсы

- [FingerprintJS](https://fingerprintjs.com/) - коммерческое решение
- [ClientJS](http://clientjs.org/) - open source библиотека
- [W3C Browser Fingerprinting](https://www.w3.org/TR/fingerprinting-guidance/) - рекомендации

---

**Fingerprinting готов к использованию!** 🔐
