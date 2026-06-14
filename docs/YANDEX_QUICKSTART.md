# ⚡ Быстрая настройка Yandex SpeechKit

## За 5 минут

### 1. Получите данные из Yandex Cloud

```bash
# Folder ID (из URL консоли)
FOLDER_ID="b1gxxxxxxxxxxxxxxxxxx"

# OAuth токен (из Яндекс.Паспорта)
OAUTH_TOKEN="y0_xxxxxxxxxxxxxxxxxxxxxxxx"
```

### 2. Получите IAM токен

```bash
curl -X POST \
  'https://iam.api.cloud.yandex.net/iam/v1/tokens' \
  -H 'Content-Type: application/json' \
  -d "{\"yandexPassportOauthToken\": \"$OAUTH_TOKEN\"}" \
  | jq -r '.iamToken'
```

Скопируйте `iamToken` из ответа.

### 3. Обновите .env.json

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

### 4. Проверьте работу

```bash
# Перезапустите бэкенд
cd Backend && go run ./cmd/main.go

# В логах должно быть:
# "Yandex SpeechKit initialized successfully"
```

### 5. Протестируйте

```bash
# Отправьте голосовое сообщение
curl -X POST \
  'http://localhost:8080/api/v1/voice/transcribe' \
  -H 'Authorization: Bearer YOUR_JWT_TOKEN' \
  -F 'file=@test.ogg'
```

**Готово!** 🎉

---

## Если что-то не работает

### Ошибка: `UNAUTHENTICATED`
- IAM токен истек (действует 1 час)
- Получите новый токен командой выше

### Ошибка: `PERMISSION_DENIED`
- Назначьте роль `speechkit.stt.user` сервисному аккаунту
- Проверьте Folder ID

### Ошибка: `Yandex SpeechKit is not enabled`
- Установите `"enabled": true` в .env.json

## Прод для продакшена

Для автоматического обновления токена используйте `service_key`:

```bash
# Создайте API ключ
yc iam key create --service-account-id <sa_id> --output key.json

# Используйте в .env.json
"service_key": "содержимое_key.json"
```

Приложение будет автоматически обновлять IAM токен!
