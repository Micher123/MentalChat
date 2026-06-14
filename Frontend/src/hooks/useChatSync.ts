import { useState, useEffect, useCallback } from 'react'
import { chatSyncService } from '../services/chatSyncService'
import { chatApi } from '../services/api'

interface UseChatSyncOptions {
  enabled?: boolean
  userId?: number
  syncIntervalSeconds?: number
}

interface UseChatSyncReturn {
  isSyncing: boolean
  lastSyncTime: Date | null
  unsyncedCount: number
  syncError: string | null
  syncNow: () => Promise<void>
  setSyncInterval: (seconds: number) => void
  syncInterval: number
}

export function useChatSync(options: UseChatSyncOptions = {}): UseChatSyncReturn {
  const {
    enabled = true,
    userId,
    syncIntervalSeconds = 300
  } = options

  const [isSyncing, setIsSyncing] = useState(false)
  const [lastSyncTime, setLastSyncTime] = useState<Date | null>(null)
  const [unsyncedCount, setUnsyncedCount] = useState(0)
  const [syncError, setSyncError] = useState<string | null>(null)
  const [syncInterval, setSyncIntervalState] = useState(syncIntervalSeconds)

  // Инициализация
  useEffect(() => {
    if (!enabled) return

    chatSyncService.init()
    chatSyncService.setSyncInterval(syncIntervalSeconds)

    // Загрузка статистики
    const loadStats = async () => {
      const stats = await chatSyncService.getSyncStats()
      setUnsyncedCount(stats.unsyncedMessages)
    }
    loadStats()

    // Авто-синхронизация
    const syncCallback = async (messages: any[]) => {
      try {
        await chatApi.syncMessages({
          messages: messages.map(m => ({
            chatId: m.chatId,
            content: m.content,
            isFromAI: m.isFromAI,
            role: m.role,
            timestamp: m.timestamp,
            localId: m.id
          }))
        })

        // Помечаем сообщения как синхронизированные
        const syncedIds = messages.map(m => m.id)
        await chatSyncService.markMessagesSynced(syncedIds)
        
        setLastSyncTime(new Date())
        setSyncError(null)
      } catch (err: any) {
        setSyncError(err.response?.data?.error || 'Sync failed')
        throw err
      }
    }

    chatSyncService.startAutoSync(userId || 0, syncCallback)

    // Обновление статистики каждые 30 секунд
    const statsInterval = setInterval(async () => {
      const stats = await chatSyncService.getSyncStats()
      setUnsyncedCount(stats.unsyncedMessages)
    }, 30000)

    return () => {
      chatSyncService.stopAutoSync()
      clearInterval(statsInterval)
    }
  }, [enabled, userId, syncIntervalSeconds])

  // Ручная синхронизация
  const syncNow = useCallback(async () => {
    if (isSyncing) return

    setIsSyncing(true)
    setSyncError(null)

    try {
      const unsynced = await chatSyncService.getUnsyncedMessages()
      
      if (unsynced.length === 0) {
        setLastSyncTime(new Date())
        return
      }

      await chatApi.syncMessages({
        messages: unsynced.map(m => ({
          chatId: m.chatId,
          content: m.content,
          isFromAI: m.isFromAI,
          role: m.role,
          timestamp: m.timestamp,
          localId: m.id
        }))
      })

      const syncedIds = unsynced.map(m => m.id)
      await chatSyncService.markMessagesSynced(syncedIds)
      
      setLastSyncTime(new Date())
      setUnsyncedCount(0)
    } catch (err: any) {
      setSyncError(err.response?.data?.error || 'Sync failed')
      console.error('Manual sync failed:', err)
    } finally {
      setIsSyncing(false)
    }
  }, [isSyncing])

  // Обновление интервала
  const handleSetSyncInterval = useCallback((seconds: number) => {
    chatSyncService.setSyncInterval(seconds)
    setSyncIntervalState(seconds)
  }, [])

  return {
    isSyncing,
    lastSyncTime,
    unsyncedCount,
    syncError,
    syncNow,
    setSyncInterval: handleSetSyncInterval,
    syncInterval
  }
}

export default useChatSync