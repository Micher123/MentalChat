# 🎙️ Голосовой ввод в MentalChat

## Обзор

Голосовой ввод позволяет пользователям отправлять сообщения голосом, используя Web Speech API браузера. Это обеспечивает удобство использования и доступность для всех пользователей.

## Архитектура

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Фронтенд      │────▶│   Бэкенд         │────▶│   Yandex        │
│   (React)       │     │   (Go)           │     │   SpeechKit     │
│                 │     │                  │     │                 │
│ - VoiceInput    │     │ - VoiceHandler   │     │ - Transcription │
│ - useSpeech     │     │ - AIService      │     │ - STT API       │
│   Recognition   │     │ - Transcribe     │     │                 │
└─────────────────┘     └──────────────────┘     └─────────────────┘
```

## Компоненты фронтенда

### 1. VoiceInputButton.tsx

Основной компонент кнопки голосового ввода.

**Функции:**
- Запрос разрешения на доступ к микрофону
- Отображение модального окна с запросом разрешения
- Сохранение выбора пользователя в БД
- Визуальная индикация записи (анимация)
- Обработка ошибок

**Пропсы:**
```typescript
interface VoiceInputButtonProps {
  onTranscript: (text: string) => void  // Callback с результатом
  disabled?: boolean                     // Состояние disabled
  className?: string                     // Дополнительные CSS классы
}
```

### 2. useSpeechRecognition.ts

Хук для работы с Web Speech API.

**Возвращаемые значения:**
```typescript
{
  isListening: boolean           // Статус прослушивания
  isSupported: boolean           // Поддержка браузером
  transcript: string             // Текущий текст
  startListening: () => void     // Начать запись
  stopListening: () => void      // Остановить запись
  clearTranscript: () => void    // Очистить текст
  requestMicrophonePermission: () => Promise<boolean>
}
```

**Опции:**
```typescript
{
  onResult?: (result) => void    // Обработчик результата
  onError?: (error) => void      // Обработчик ошибок
  onStart?: () => void           // Начало записи
  onEnd?: () => void             // Конец записи
  continuous?: boolean           // Продолжительная запись
  interimResults?: boolean       // Промежуточные результаты
  language?: string              // Язык (ru-RU)
}
```

## API Бэкенда

### POST /api/v1/voice/transcribe

Транскрибация голосового сообщения.

**Запрос:**
```
Content-Type: multipart/form-data

file: <audio_file.ogg>
```

**Ответ:**
```json
{
  "transcript": "Привет, как дела?",
  "file_name": "20240608_120000_voice.ogg",
  "file_size": 1024
}
```

**Коды ошибок:**
- `400` - Неверный формат файла
- `401` - Неавторизован
- `500` - Ошибка транскрибации

### POST /api/v1/user/microphone-permission

Сохранение разрешения на использование микрофона.

**Запрос:**
```json
{
  "granted": true
}
```

**Ответ:**
```json
{
  "message": "Microphone permission saved",
  "granted": true
}
```

## Поддерживаемые форматы аудио

- ✅ audio/ogg (рекомендуется)
- ✅ audio/webm
- ✅ audio/wav

## Интеграция с Yandex SpeechKit

### Настройка

1. Создайте сервисный аккаунт в Yandex Cloud
2. Получите IAM токен
3. Включите SpeechKit API

### Конфигурация

```json
{
  "ai": {
    "yandex_speechkit": {
      "folder_id": "your_folder_id",
      "api_key": "your_api_key"
    }
  }
}
```

### Пример запроса к Yandex STT

```bash
curl -X POST \
  'https://stt.api.cloud.yandex.net/speech/v1/stt:recognize' \
  -H 'Authorization: Bearer <IAM_TOKEN>' \
  -H 'Content-Type: audio/ogg;codecs=opus' \
  --data-binary @voice.ogg
```

## Безопасность

### Ограничения

1. **Размер файла**: Максимум 10 MB
2. **Длительность**: Максимум 60 секунд
3. **Частота запросов**: Rate limiting (100 запросов/мин)
4. **Проверка типа файла**: Только аудио форматы

### Хранение

- Голосовые файлы хранятся в `./storage/voices/`
- Имя файла: `<timestamp>_<original_name>.ogg`
- После транскрибации оригинал может быть удален (настраивается)

## UX Особенности

### Модальное окно разрешения

При первом использовании показывается модальное окно с:
- Иконкой микрофона 🎙️
- Объяснением преимуществ
- Чекбоксом "Запомнить выбор"
- Кнопками "Да" / "Нет"

### Визуальная обратная связь

- **Ожидание**: Серая кнопка с иконкой микрофона
- **Запись**: Пульсирующая красная кнопка с индикатором
- **Ошибка**: Всплывающее уведомление
- **Успех**: Текст вставляется в поле ввода

### Обработка ошибок

| Ошибка | Сообщение пользователю |
|--------|----------------------|
| No browser support | "Ваш браузер не поддерживает голосовой ввод" |
| Permission denied | "Разрешите доступ к микрофону" |
| No speech detected | "Речь не обнаружена, попробуйте снова" |
| Network error | "Ошибка сети, проверьте соединение" |

## Тестирование

### Ручное тестирование

1. Откройте приложение в браузере
2. Нажмите на иконку микрофона
3. Разрешите доступ к микрофону
4. Произнесите фразу
5. Проверьте, что текст появился в поле ввода

### Автоматическое тестирование

```typescript
// Пример теста
describe('VoiceInputButton', () => {
  it('should request microphone permission on first click', async () => {
    const { getByTestId } = render(<VoiceInputButton onTranscript={() => {}} />)
    const button = getByTestId('voice-input-button')
    
    fireEvent.click(button)
    
    expect(screen.getByText('Доступ к микрофону')).toBeInTheDocument()
  })
})
```

## Будущие улучшения

- [ ] Поддержка других языков (EN, ES, FR, DE)
- [ ] Распознавание команд ("отправить", "очистить", "отмена")
- [ ] Офлайн режим с кэшированием
- [ ] Улучшенное шумоподавление
- [ ] Интеграция с другими STT сервисами (Google, Azure)
- [ ] Статистика использования голосового ввода

## Поддерживаемые браузеры

| Браузер | Версия | Поддержка |
|---------|--------|-----------|
| Chrome | 25+ | ✅ Полная |
| Firefox | 49+ | ✅ Полная |
| Safari | 14.1+ | ✅ Полная |
| Edge | 79+ | ✅ Полная |
| Opera | 27+ | ✅ Полная |
| IE | Все | ❌ Нет |

## Ресурсы

- [Web Speech API Documentation](https://developer.mozilla.org/en-US/docs/Web/API/Web_Speech_API)
- [Yandex SpeechKit](https://cloud.yandex.ru/docs/speechkit/stt/)
- [W3C Speech Recognition API](https://wicg.github.io/speech-api/)
