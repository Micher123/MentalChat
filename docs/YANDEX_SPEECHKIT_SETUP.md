# 🎙️ Настройка Yandex SpeechKit для MentalChat

## 📋 Требования

1. Аккаунт в Yandex Cloud
2. Платежный аккаунт (есть бесплатный лимит)
3. Установленный CLI yandex (опционально)

## 🚀 Пошаговая инструкция

### Шаг 1: Создание сервисного аккаунта

1. Откройте [Yandex Cloud Console](https://console.cloud.yandex.ru/)
2. Перейдите в нужный каталог (folder)
3. Нажмите **"Создать ресурс"** → **"Сервисный аккаунт"**
4. Заполните:
   - **Имя**: `mentalchat-speechkit`
   - **Описание**: `Сервисный аккаунт для транскрибации голоса`
5. Нажмите **"Создать"**

### Шаг 2: Назначение ролей

Сервисному аккаунту нужны роли:

1. **`speechkit.stt.user`** - доступ к SpeechKit STT
2. **`editor`** - для получения IAM токена

Для назначения:
```bash
# Через CLI
yc resource-manager folder add-access-binding <folder_id> \
  --subject serviceAccount:<service_account_id> \
  --role speechkit.stt.user

yc resource-manager folder add-access-binding <folder_id> \
  --subject serviceAccount:<service_account_id> \
  --role editor
```

Или через консоль:
- Откройте сервисный аккаунт
- Нажмите **"Назначить роли"**
- Выберите нужные роли

### Шаг 3: Получение IAM токена

#### Вариант A: Через OAuth токен (рекомендуется для разработки)

1. Получите OAuth токен:
   ```bash
   curl -X POST \
     'https://oauth.yandex.ru/token' \
     -H 'Content-Type: application/x-www-form-urlencoded' \
     --data-urlencode 'grant_type=authorization_code' \
     --data-urlencode 'code=<OAuth_code>' \
     --data-urlencode 'client_id=<client_id>' \
     --data-urlencode 'client_secret=<client_secret>'
   ```

2. Обменяйте OAuth токен на IAM:
   ```bash
   curl -X POST \
     'https://iam.api.cloud.yandex.net/iam/v1/tokens' \
     -H 'Content-Type: application/json' \
     -d '{
       "yandexPassportOauthToken": "<OAuth_token>"
     }'
   ```

3. Скопируйте `iamToken` из ответа

#### Вариант B: Через API ключ (для продакшена)

1. Создайте API ключ для сервисного аккаунта:
   ```bash
   yc iam key create --service-account-id <service_account_id> --output key.json
   ```

2. Используйте ключ для получения IAM токена

#### Вариант C: Через метаданные (для VM в Yandex Cloud)

Если приложение работает на VM в Yandex Cloud:
```bash
curl -H "Metadata-Flavor: Google" \
  'http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/token'
```

### Шаг 4: Получение Folder ID

1. Откройте [Yandex Cloud Console](https://console.cloud.yandex.ru/)
2. Перейдите в нужный каталог
3. Folder ID указан в URL или в карточке каталога

Пример URL: `https://console.cloud.yandex.ru/folder/<folder_id>`

### Шаг 5: Настройка .env.json

Откройте `.env.json` и заполните:

```json
{
  "ai": {
    "yandex_speechkit": {
      "enabled": true,
      "folder_id": "b1gxxxxxxxxxxxxxxxxxx",
      "iam_token": "t1.xxxxxxxxxxxxxxxxxxxxxxxx",
      "service_key": "y0_xxxxxxxxxxxxxxxxxxxxxxxx",
      "api_endpoint": "https://stt.api.cloud.yandex.net"
    }
  }
}
```

**Важно:**
- `iam_token` действует 1 час
- Для продакшена используйте `service_key` для авто-обновления
- В разработке можно обновлять токен вручную

### Шаг 6: Проверка работы

#### Тест через curl:

```bash
# Получите IAM токен
IAM_TOKEN="t1.xxxxxxxxxxxxxxxxxxxxxxxx"

# Отправьте аудио файл
curl -X POST \
  'https://stt.api.cloud.yandex.net/speech/v1/stt:recognize' \
  -H "Authorization: Bearer $IAM_TOKEN" \
  -H 'Content-Type: audio/ogg;codecs=opus' \
  --data-binary @test.ogg
```

#### Тест через приложение:

```bash
# Запустите бэкенд
cd Backend
go run ./cmd/main.go

# В другом терминале отправьте аудио
curl -X POST \
  'http://localhost:8080/api/v1/voice/transcribe' \
  -H 'Authorization: Bearer <your_jwt_token>' \
  -F 'file=@test.ogg'
```

## 💰 Тарифы и лимиты

### Бесплатный тариф (ежемесячно):
- **1 час** аудио в месяц бесплатно
- Далее: **0.06 ₽** за минуту

### Лимиты:
- Макс. длительность запроса: **60 секунд** (для синхронного API)
- Макс. размер файла: **10 MB**
- RPS: **10 запросов в секунду**

### Оптимизация затрат:

1. **Кэширование результатов** - не транскрибируйте одинаковые файлы
2. **Сжатие аудио** - используйте формат OPUS
3. **Ограничение длительности** - обрезайте аудио до 60 сек
4. **Пакетная обработка** - для длинных файлов используйте асинхронный API

## 🔧 Расширенные настройки

### Выбор модели распознавания

В конфиге можно указать модель:

```json
{
  "ai": {
    "yandex_speechkit": {
      "enabled": true,
      "model": "phonecall"  // или "general", "video"
    }
  }
}
```

**Доступные модели:**
- `general` - общая модель (по умолчанию)
- `phonecall` - для телефонных разговоров
- `video` - для видео с субтитрами

### Настройка языка

Поддерживаемые языки:
- `ru-RU` - Русский
- `en-US` - Английский
- `tr-TR` - Турецкий
- `kk-KZ` - Казахский
- `uk-UA` - Украинский
- `de-DE` - Немецкий
- `fr-FR` - Французский
- `es-ES` - Испанский
- `it-IT` - Итальянский
- `pt-PT` - Португальский

### Фильтр ненормативной лексики

```json
{
  "ai": {
    "yandex_speechkit": {
      "profanity_filter": true
    }
  }
}
```

## 🐛 Отладка

### Включение логов

В `.env.json`:
```json
{
  "server": {
    "debug": true
  }
}
```

### Просмотр логов Yandex SpeechKit

```bash
# Логи бэкенда
tail -f Backend/logs/app.log | grep -i speechkit

# Или через journalctl
journalctl -u mentalchat -f
```

### Частые ошибки

| Ошибка | Причина | Решение |
|--------|---------|---------|
| `UNAUTHENTICATED` | Неверный IAM токен | Обновите токен |
| `PERMISSION_DENIED` | Нет роли speechkit.stt.user | Назначьте роль |
| `RESOURCE_EXHAUSTED` | Превышен лимит RPS | Добавьте задержку |
| `INVALID_ARGUMENT` | Неверный формат аудио | Используйте OGG/OPUS |
| `DEADLINE_EXCEEDED` | Таймаут запроса | Увеличьте timeout |

## 📊 Мониторинг

### Метрики для отслеживания:

1. **Количество запросов** в день
2. **Средняя длительность** аудио
3. **Процент успешных** транскрибаций
4. **Время ответа** API
5. **Затраты** на транскрибацию

### Дашборд в Yandex Cloud:

1. Откройте **Monitoring** в консоли
2. Создайте дашборд
3. Добавьте метрики:
   - `speechkit.stt.requests`
   - `speechkit.stt.errors`
   - `speechkit.stt.duration`

## 🔐 Безопасность

### Рекомендации:

1. **Не храните IAM токены в коде**
2. **Используйте Secrets Manager** для хранения ключей
3. **Регулярно обновляйте** IAM токены
4. **Ограничьте доступ** к сервисному аккаунту
5. **Включите аудит** действий

### Rotation токенов:

Приложение автоматически обновляет IAM токены каждые 50 минут.

Для ручного обновления:
```bash
curl http://localhost:8080/api/v1/admin/refresh-iam-token \
  -H 'Authorization: Bearer <admin_token>'
```

## 📚 Дополнительные ресурсы

- [Документация Yandex SpeechKit](https://cloud.yandex.ru/docs/speechkit/stt/)
- [API Reference](https://cloud.yandex.ru/docs/speechkit/api-ref/)
- [Примеры кода](https://github.com/yandex-cloud/examples/tree/master/speechkit)
- [Калькулятор стоимости](https://cloud.yandex.ru/pricing?products=speechkit)

## 🆘 Поддержка

При проблемах:

1. Проверьте логи приложения
2. Проверьте квоты в Yandex Cloud
3. Убедитесь, что роли назначены правильно
4. Проверьте срок действия IAM токена
5. Обратитесь в поддержку Yandex Cloud

---

**Готово!** 🎉 Yandex SpeechKit настроен и готов к работе!
