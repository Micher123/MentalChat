import { openDB } from 'idb'

interface ChatMessage {
  id: number
  chatId: string
  userId: number
  content: string
  isFromAI: boolean
  role: 'user' | 'ai'
  timestamp: string
  synced: boolean
  localId?: number
}

interface ChatStore {
  id: string
  chatType: string
  userId: number
  messages: any[]
  lastMessageTime: string
  createdAt: string
  updatedAt: string
  synced: boolean
}

class ChatSyncService {
  private db: any = null
  private syncInterval: any = null
  private syncIntervalSeconds: number = 300 // Default 5 minutes

  async init() {
    this.db = await openDB('MentalChatDB', 1, {
      upgrade(db: any) {
        // Chats store
        const chatStore = db.createObjectStore('chats', { keyPath: 'id' })
        chatStore.createIndex('by-type', 'chatType')
        chatStore.createIndex('by-sync', 'synced')

        // Messages store
        const msgStore = db.createObjectStore('messages', { keyPath: 'id' })
        msgStore.createIndex('by-chat', 'chatId')
        msgStore.createIndex('by-sync', 'synced')
        msgStore.createIndex('by-time', 'timestamp')

        // Sync log store
        db.createObjectStore('syncLog', { keyPath: 'id', autoIncrement: true })
      }
    })
    console.log('✅ ChatSyncService: IndexedDB initialized')
  }

  async saveChat(chatType: string, userId: number) {
    if (!this.db) await this.init()
    
    const chatId = `${userId}_${chatType}_${Date.now()}`
    const now = new Date().toISOString()

    await this.db!.put('chats', {
      id: chatId,
      chatType,
      userId,
      messages: [],
      lastMessageTime: now,
      createdAt: now,
      updatedAt: now,
      synced: false
    })

    return chatId
  }

  // Получить или создать чат для пары userId + chatType
  async getOrCreateChat(userId: number, chatType: string): Promise<string> {
    if (!this.db) await this.init()

    // Ищем существующий чат этого типа для пользователя
    const allChats = await this.db!.getAll('chats')
    const existing = (allChats as ChatStore[]).find(
      (c: ChatStore) => c.userId === userId && c.chatType === chatType
    )

    if (existing) {
      return existing.id
    }

    // Создаём новый
    const chatId = `${userId}_${chatType}`
    const now = new Date().toISOString()
    await this.db!.put('chats', {
      id: chatId,
      chatType,
      userId,
      messages: [],
      lastMessageTime: now,
      createdAt: now,
      updatedAt: now,
      synced: false
    })

    return chatId
  }

  async addMessage(chatId: string, message: Omit<ChatMessage, 'id' | 'synced'>) {
    if (!this.db) await this.init()

    const localId = Date.now()
    const msg = {
      ...message,
      chatId,
      id: localId,
      synced: false,
      timestamp: new Date().toISOString()
    }

    await this.db!.put('messages', msg)
    
    // Автосоздание чата, если его нет
    let chat = await this.db!.get('chats', chatId)
    if (!chat) {
      // Создаём чат на основе chatId (формат: userId_chatType)
      const parts = chatId.split('_')
      chat = {
        id: chatId,
        chatType: parts.length >= 2 ? parts[1] : 'unknown',
        userId: parts.length >= 1 ? parseInt(parts[0]) : 0,
        messages: [],
        lastMessageTime: msg.timestamp,
        createdAt: msg.timestamp,
        updatedAt: msg.timestamp,
        synced: false
      }
      await this.db!.put('chats', chat)
    } else {
      chat.lastMessageTime = msg.timestamp
      chat.updatedAt = new Date().toISOString()
      await this.db!.put('chats', chat)
    }

    console.log('💾 ChatSyncService: Message saved locally', msg.id)
    return localId
  }

  async getUnsyncedMessages(_chatId?: string) {
    if (!this.db) await this.init()
    const allMessages = await this.db!.getAll('messages')
    return (allMessages as ChatMessage[]).filter((m: ChatMessage) => !m.synced)
  }

  async markMessagesSynced(messageIds: number[]) {
    if (!this.db) return

    for (const id of messageIds) {
      const msg = await this.db!.get('messages', id)
      if (msg) {
        msg.synced = true
        await this.db!.put('messages', msg)
      }
    }
  }

  async markChatSynced(chatId: string) {
    if (!this.db) return
    const chat = await this.db!.get('chats', chatId)
    if (chat) {
      chat.synced = true
      await this.db!.put('chats', chat)
    }
  }

  async getLocalChats(userId: number) {
    if (!this.db) await this.init()

    const allChats = await this.db!.getAll('chats')
    return (allChats as ChatStore[]).filter((c: ChatStore) => c.userId === userId)
  }

  async getLocalMessages(chatId: string) {
    if (!this.db) await this.init()

    const messages = await this.db!.getAllFromIndex('messages', 'by-chat', chatId)
    return (messages as ChatMessage[]).sort((a: ChatMessage, b: ChatMessage) => 
      new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
    )
  }

  async deleteLocalChat(chatId: string) {
    if (!this.db) return
    await this.db!.delete('chats', chatId)
    const messages = await this.db!.getAllFromIndex('messages', 'by-chat', chatId)
    for (const msg of (messages as ChatMessage[])) {
      await this.db!.delete('messages', msg.id)
    }
  }

  async logSync(action: string, count: number, success: boolean, error?: string) {
    if (!this.db) return
    await this.db!.add('syncLog', {
      action,
      count,
      timestamp: new Date().toISOString(),
      success,
      error
    })
  }

  // Авто-синхронизация
  startAutoSync(_userId: number, syncCallback: (messages: any[]) => Promise<void>) {
    // Читаем интервал из localStorage (можно установить из .env при загрузке)
    const storedInterval = localStorage.getItem('chat_sync_interval')
    this.syncIntervalSeconds = storedInterval ? parseInt(storedInterval) : 300

    this.stopAutoSync()

    this.syncInterval = setInterval(async () => {
      console.log('🔄 ChatSyncService: Auto-sync starting...')
      
      try {
        const unsynced = await this.getUnsyncedMessages()
        if (unsynced.length > 0) {
          console.log(`🔄 ChatSyncService: Syncing ${unsynced.length} messages...`)
          await syncCallback(unsynced)
          await this.logSync('auto-sync', unsynced.length, true)
        } else {
          console.log('🔄 ChatSyncService: Nothing to sync')
        }
      } catch (err: any) {
        console.error('❌ ChatSyncService: Sync failed', err)
        await this.logSync('auto-sync', 0, false, err.message)
      }
    }, this.syncIntervalSeconds * 1000)

    console.log(`✅ ChatSyncService: Auto-sync started (${this.syncIntervalSeconds}s interval)`)
  }

  stopAutoSync() {
    if (this.syncInterval) {
      clearInterval(this.syncInterval)
      this.syncInterval = null
      console.log('⏹️ ChatSyncService: Auto-sync stopped')
    }
  }

  setSyncInterval(seconds: number) {
    this.syncIntervalSeconds = seconds
    localStorage.setItem('chat_sync_interval', seconds.toString())
    console.log(`⚙️ ChatSyncService: Sync interval set to ${seconds}s`)
  }

  getSyncInterval() {
    return this.syncIntervalSeconds
  }

  async getSyncStats() {
    if (!this.db) await this.init()

    const unsyncedMessages = (await this.getUnsyncedMessages()).length
    const totalChats = (await this.db!.getAll('chats')).length
    const lastSyncLogs = await this.db!.getAll('syncLog')
    const lastSync = lastSyncLogs.length > 0 ? lastSyncLogs[lastSyncLogs.length - 1] : null

    return {
      unsyncedMessages,
      totalChats,
      lastSync
    }
  }

  async deleteLocalMessage(messageId: number) {
    if (!this.db) return
    await this.db!.delete('messages', messageId)
    console.log('🗑️ ChatSyncService: Local message deleted', messageId)
  }

  async deleteLocalMessages(messageIds: number[]) {
    if (!this.db) return
    for (const id of messageIds) {
      await this.db!.delete('messages', id)
    }
    console.log(`🗑️ ChatSyncService: ${messageIds.length} local messages deleted`)
  }

  async clearLocalChatMessages(chatId: string) {
    if (!this.db) return
    const messages = await this.db!.getAllFromIndex('messages', 'by-chat', chatId)
    for (const msg of (messages as ChatMessage[])) {
      await this.db!.delete('messages', msg.id)
    }
    console.log(`🗑️ ChatSyncService: All messages cleared for chat ${chatId}`)
  }

  async clearAll() {
    if (!this.db) return
    await this.db!.clear('chats')
    await this.db!.clear('messages')
    await this.db!.clear('syncLog')
    console.log('🗑️ ChatSyncService: All local data cleared')
  }
}

export const chatSyncService = new ChatSyncService()
export default chatSyncService