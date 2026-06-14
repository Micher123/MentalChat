# 🔄 Система синхронизации чатов

## Описание

MentalChat теперь сохраняет историю переписки **локально на устройстве** (IndexedDB) и автоматически синхронизирует её с сервером в настраиваемые промежутки времени.

## 📋 Возможности

### ✅ Локальное хранение
- Все сообщения сохраняются в браузере через IndexedDB
- Работает даже без интернета
- Мгновенная отправка сообщений без ожидания ответа сервера

### ✅ Автоматическая синхронизация
- Синхронизация в фоновом режиме
- Настраиваемый интервал (по умолчанию 5 минут)
- Повторные попытки при ошибке сети

### ✅ Двусторонняя синхронизация
- Сообщения сохраняются локально И на сервере
- Защита от потери данных
- Доступ к истории с любого устройства

## ⚙️ Настройка

### В `.env.json` (Backend):

```json
{
  "app": {
    "chat_sync_interval_seconds": 300
  }
}
```

**Возможные значения:**
- `60` - каждые 60 секунд (1 минута)
- `180` - каждые 180 секунд (3 минуты)  
- `300` - каждые 300 секунд (5 минут) ← **по умолчанию**
- `600` - каждые 600 секунд (10 минут)
- `3600` - каждые 3600 секунд (1 час)

### В браузере (Frontend):

Интервал можно изменить динамически через localStorage:

```javascript
localStorage.setItem('chat_sync_interval', '600') // 10 минут
```

## 🛠️ API Endpoints

### POST `/api/v1/chat/sync`

**Аутентификация:** Требуется JWT token

**Request:**
```json
{
  "messages": [
    {
      "local_id": 1234567890,
      "chat_id": "psychologist",
      "content": "Привет!",
      "is_from_ai": false,
      "role": "user",
      "timestamp": "2026-06-14T09:30:00.000Z"
    },
    {
      "local_id": 1234567891,
      "chat_id": "psychologist",
      "content": "Здравствуйте! Чем могу помочь?",
      "is_from_ai": true,
      "role": "ai",
      "timestamp": "2026-06-14T09:30:05.000Z"
    }
  ]
}
```

**Response:**
```json
{
  "message": "Sync completed",
  "synced": 2,
  "failed": 0,
  "synced_ids": [1234567890, 1234567891]
}
```

## 💻 Использование в коде

### Хук `useChatSync`

```typescript
import { useChatSync } from './hooks/useChatSync'

function ChatComponent() {
  const {
    isSyncing,
    lastSyncTime,
    unsyncedCount,
    syncError,
    syncNow,
    setSyncInterval,
    syncInterval
  } = useChatSync({
    enabled: true,
    userId: user.id,
    syncIntervalSeconds: 300
  })

  return (
    <div>
      <p>Не синхронизировано: {unsyncedCount}</p>
      {lastSyncTime && <p>Последняя синхронизация: {lastSyncTime.toLocaleTimeString()}</p>}
      {syncError && <p className="error">{syncError}</p>}
      
      <button onClick={syncNow} disabled={isSyncing}>
        {isSyncing ? 'Синхронизация...' : 'Синхронизировать'}
      </button>
      
      <select onChange={(e) => setSyncInterval(Number(e.target.value))}>
        <option value={60}>Каждую минуту</option>
        <option value={300}>Каждые 5 минут</option>
        <option value={600}>Каждые 10 минут</option>
      </select>
    </div>
  )
}
```

### Сервис `chatSyncService`

```typescript
import { chatSyncService } from './services/chatSyncService'

// Инициализация
await chatSyncService.init()

// Сохранить сообщение локально
const localId = await chatSyncService.addMessage(chatId, {
  userId: 1,
  content: 'Привет!',
  isFromAI: false,
  role: 'user'
})

// Получить неотсинхронизированные сообщения
const unsynced = await chatSyncService.getUnsyncedMessages()

// Пометить как синхронизированные
await chatSyncService.markMessagesSynced([localId])

// Получить статистику
const stats = await chatSyncService.getSyncStats()
console.log('Неотсинхронизировано:', stats.unsyncedMessages)
console.log('Последняя синхронизация:', stats.lastSync)

// Изменить интервал
chatSyncService.setSyncInterval(600) // 10 минут

// Остановить авто-синхронизацию
chatSyncService.stopAutoSync()

// Удалить все локальные данные
await chatSyncService.clearAll()
```

## 📊 Структура базы данных (IndexedDB)

### Store: `chats`
```typescript
{
  id: string,              // Уникальный ID чата
  chatType: string,        // psychologist | tarot | sexologist | fortune_teller
  userId: number,          // ID пользователя
  messages: any[],         // Массив сообщений (опционально)
  lastMessageTime: string, // ISO timestamp последнего сообщения
  createdAt: string,       // ISO timestamp создания
  updatedAt: string,       // ISO timestamp обновления
  synced: boolean          // Статус синхронизации
}
```

### Store: `messages`
```typescript
{
  id: number,              // Временный ID (timestamp)
  chatId: string,          // ID чата
  userId: number,          // ID пользователя
  content: string,         // Текст сообщения
  isFromAI: boolean,       // От AI или пользователя
  role: 'user' | 'ai',     // Роль
  timestamp: string,       // ISO timestamp
  localId?: number,        // Локальный ID
  synced: boolean          // Статус синхронизации
}
```

### Store: `syncLog`
```typescript
{
  id: number,              // Автоинкремент
  action: string,          // 'auto-sync' | 'manual-sync'
  count: number,           // Количество сообщений
  timestamp: string,       // ISO timestamp
  success: boolean,        // Успешно или нет
  error?: string           // Ошибка (если failed)
}
```

## 🔒 Безопасность

- Все сообщения шифруются перед отправкой (стандартный HTTPS)
- JWT token для аутентификации
- Проверка ownership сообщений (только свои)
- Защита от дубликатов (по local_id)

## 🐛 Диагностика

### Проверка IndexedDB:
```javascript
// В консоли браузера
const req = indexedDB.open('MentalChatDB')
req.onsuccess = (event) => {
  const db = event.target.result
  console.log('Stores:', db.objectStoreNames)
  
  // Просмотр сообщений
  const tx = db.transaction('messages', 'readonly')
  const store = tx.objectStore('messages')
  store.getAll().onsuccess = (e) => {
    console.log('Messages:', e.target.result)
  }
}
```

### Проверка синхронизации:
```javascript
// Статистика синхронизации
import { chatSyncService } from './services/chatSyncService'
const stats = await chatSyncService.getSyncStats()
console.log(stats)

// Просмотр лога
const tx = db.transaction('syncLog', 'readonly')
const store = tx.objectStore('syncLog')
store.getAll().onsuccess = (e) => {
  console.log('Sync log:', e.target.result)
}
```

### Очистка данных:
```javascript
// Очистить всё
await chatSyncService.clearAll()

// Очистить IndexedDB вручную
indexedDB.deleteDatabase('MentalChatDB')
```

## 📝 Логирование

**Backend логи:**
```
[INFO] Sync completed - synced: 5, failed: 0
[ERROR] Sync failed - User not found
```

**Frontend логи (Console):**
```
✅ ChatSyncService: IndexedDB initialized
🔄 ChatSyncService: Auto-sync starting...
🔄 ChatSyncService: Syncing 5 messages...
💾 ChatSyncService: Message saved locally 1234567890
```

## 🔄 Сценарии использования

### 1. Регистрация нового пользователя:
```
1. Пользователь регистрируется
2. userId = 42
3. Авто-синхронизация стартует (300s интервал)
4. Сообщения сохраняются в IndexedDB
5. Каждые 5 минут → синхронизация с сервером
```

### 2. Офлайн режим:
```
1. Нет интернета
2. Сообщения сохраняются в IndexedDB
3. unsyncedCount увеличивается
4. При появлении сети → авто-синхронизация
```

### 3. Смена устройства:
```
1. Пользователь входит с нового устройства
2. Загрузка истории с сервера
3. Новые сообщения синхронизируются
4. История доступна на обоих устройствах
```

## 📦 Зависимости

```json
{
  "idb": "^8.0.0"
}
```

Установлено автоматически через `npm install idb`

## ✅ Чеклист внедрения

- [x] IndexedDB для локального хранения
- [x] Автосинхронизация в фоне
- [x] Ручная синхронизация
- [x] Настройка интервала через .env
- [x] Настройка интервала через localStorage
- [x] API endpoint для синхронизации
- [x] Защита от дубликатов
- [x] Логирование синхронизации
- [x] Статистика синхронизации
- [x] Документация

---

**Версия:** 1.0.0  
**Дата:** 2026-06-14  
**Автор:** NLP-Core-Team