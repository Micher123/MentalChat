import React, { useState, useRef, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import useAppStore, { themes } from '../store/useAppStore'
import VoiceInputButton from '../components/VoiceInputButton'
import { StrawberryIcon, BrainIcon, CardsIcon, MysticOrbIcon, ClosedBookIcon, OpenBookIcon, UserIcon } from '../components/DeveloperIcons'
import { chatApi } from '../services/api'
import { chatSyncService } from '../services/chatSyncService'
import { useChatSync } from '../hooks/useChatSync'

const DashboardPage = () => {
  const navigate = useNavigate()
  const { user, currentChatType, setCurrentChatType, chatHistories, addMessage, setChatHistory, setWelcomeModalOpen, theme, setTheme } = useAppStore()

  const [inputMessage, setInputMessage] = useState('')
  const [loading, setLoading] = useState(false)
  const [typingState, setTypingState] = useState<'idle' | 'entering' | 'visible' | 'exiting'>('idle')
  const typingMinTimeRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const [chatSessions, setChatSessions] = useState<any[]>([])
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const [searchTerm, setSearchTerm] = useState('')
  const [showSearch, setShowSearch] = useState(false)
  const [showLocalHistory, setShowLocalHistory] = useState(false)
  const [localChats, setLocalChats] = useState<any[]>([])
  const [localMessages, setLocalMessages] = useState<any[]>([])
  const [showSettings, setShowSettings] = useState(false)
  const [showProfile, setShowProfile] = useState(false)
  const [deleteMode, setDeleteMode] = useState(false)
  const [selectedMessages, setSelectedMessages] = useState<Set<number>>(new Set())
  const [showClearConfirm, setShowClearConfirm] = useState(false)
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const settingsPanelRef = useRef<HTMLDivElement>(null)
  const profilePanelRef = useRef<HTMLDivElement>(null)

  const currentChatHistory = chatHistories[currentChatType] || []

  // Инициализация синхронизации (только статус, без кнопки)
  const { unsyncedCount, lastSyncTime } = useChatSync({
    enabled: !!user,
    userId: user?.id,
    syncIntervalSeconds: 300
  })

  // Welcome message
  React.useEffect(() => {
    if (user && Object.keys(chatHistories).length === 0) {
      setWelcomeModalOpen(true)
    }
  }, [user, chatHistories, setWelcomeModalOpen])

  // Загрузка истории текущего чата из IndexedDB — только если в сторе ещё пусто
  const loadChatHistory = useCallback(async (chatType: string) => {
    if (!user) return
    
    // Не перезаписываем сообщения, которые уже есть в Zustand (текущая сессия)
    const existing = useAppStore.getState().chatHistories[chatType]
    if (existing && existing.length > 0) {
      console.log(`📋 Chat history for ${chatType} already in store, skipping IndexedDB load`)
      return
    }
    
    try {
      const chats = await chatSyncService.getLocalChats(user.id)
      const matchingChats = chats.filter((c: any) => c.chatType === chatType)
      
      if (matchingChats.length > 0) {
        const latestChat = matchingChats.sort((a: any, b: any) => 
          new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
        )[0]
        const msgs = await chatSyncService.getLocalMessages(latestChat.id)
        
        const formattedMessages = msgs.map((m: any) => ({
          id: m.id,
          userId: m.userId,
          chatType: m.chatType || chatType,
          content: m.content,
          isFromAI: m.role === 'ai',
          role: m.role,
          timestamp: m.timestamp,
        }))
        
        setChatHistory(chatType, formattedMessages)
        console.log(`📋 Loaded ${formattedMessages.length} messages for ${chatType} from IndexedDB`)
      }
    } catch (err) {
      console.error('Failed to load chat history:', err)
    }
  }, [user, setChatHistory])

  // Загружаем историю когда user становится доступен (Zustand persist восстанавливается асинхронно)
  useEffect(() => {
    if (user) {
      loadChatHistory(currentChatType)
    }
  }, [user]) // eslint-disable-line react-hooks/exhaustive-deps

  // Загрузка локальных чатов для сайдбара
  React.useEffect(() => {
    const loadLocalChats = async () => {
      if (user) {
        try {
          const chats = await chatSyncService.getLocalChats(user.id)
          setLocalChats(chats)
        } catch (err) {
          console.error('Failed to load local chats:', err)
        }
      }
    }
    loadLocalChats()
  }, [user, showLocalHistory])

  // Загрузка локальных сообщений для просмотра истории
  React.useEffect(() => {
    const loadLocalMessages = async () => {
      if (user && showLocalHistory) {
        try {
          const msgs = await chatSyncService.getLocalMessages(currentChatType)
          setLocalMessages(msgs)
        } catch (err) {
          console.error('Failed to load local messages:', err)
        }
      }
    }
    loadLocalMessages()
  }, [user, currentChatType, showLocalHistory])

  // Закрытие панели настроек при клике вне её
  useEffect(() => {
    if (!showSettings) return
    const handleClickOutside = (e: MouseEvent) => {
      if (settingsPanelRef.current && !settingsPanelRef.current.contains(e.target as Node)) {
        setShowSettings(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [showSettings])

  // Закрытие панели профиля при клике вне её
  useEffect(() => {
    if (!showProfile) return
    const handleClickOutside = (e: MouseEvent) => {
      if (profilePanelRef.current && !profilePanelRef.current.contains(e.target as Node)) {
        setShowProfile(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [showProfile])

  // Слушатель 401 → мягкий редирект через React Router
  useEffect(() => {
    const handleUnauthorized = () => {
      useAppStore.getState().setAuthenticated(false)
      navigate('/login')
    }
    window.addEventListener('auth:unauthorized', handleUnauthorized)
    return () => window.removeEventListener('auth:unauthorized', handleUnauthorized)
  }, [navigate])

  // Закрытие локальной истории при сворачивании сайдбара
  useEffect(() => {
    if (!sidebarOpen) setShowLocalHistory(false)
  }, [sidebarOpen])

  // Auto-scroll to bottom
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [currentChatHistory, localMessages])

  // Load chat sessions
  React.useEffect(() => {
    if (user) {
      setChatSessions([
        { id: 1, title: 'Психолог', lastMessage: 'Как я могу помочь?', chatType: 'psychologist', icon: 'brain' },
        { id: 2, title: 'Таролог', lastMessage: 'Сконцентрируйтесь на вопросе', chatType: 'tarot', icon: 'cards' },
        { id: 3, title: 'Сексолог', lastMessage: 'Расскажите подробнее', chatType: 'sexologist', icon: 'strawberry' },
        { id: 4, title: 'Гадалка', lastMessage: 'Карты говорят...', chatType: 'fortune_teller', icon: 'crystal-ball' },
      ])
    }
  }, [user])

  const handleSendMessage = async () => {
    if (!inputMessage.trim()) return

    const currentInput = inputMessage
    setInputMessage('')
    setLoading(true)

    // Засекаем время для минимальной длительности индикатора "Печатает..."
    const startTime = Date.now()
    const TYPING_MIN_DURATION = 800 // ms

    // Плавное появление индикатора
    setTypingState('entering')
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        setTypingState('visible')
      })
    })

    try {
      // Стабильный chatId для пары userId + chatType
      const chatId = await chatSyncService.getOrCreateChat(user?.id || 0, currentChatType)
      
      const userMessage = {
        id: Date.now(),
        userId: user?.id || 0,
        chatType: currentChatType,
        content: currentInput,
        isFromAI: false,
        role: 'user' as const,
        timestamp: new Date().toISOString(),
      }
      addMessage(userMessage)
      
      if (user) {
        await chatSyncService.addMessage(chatId, {
          ...userMessage,
          chatId
        })
      }

      const response = await chatApi.sendMessage({
        chatType: currentChatType,
        content: currentInput,
      })

      // If queued, AI is processing another request and the reply will arrive via sync/next message
      if (response.data && !response.data.queued) {
        const data = response.data
        
        const aiMessage = {
          id: Date.now() + 1,
          userId: user?.id || 0,
          chatType: currentChatType,
          content: data.message,
          isFromAI: true,
          role: 'ai' as const,
          timestamp: new Date().toISOString(),
        }
        addMessage(aiMessage)
        
        if (user) {
          await chatSyncService.addMessage(chatId, {
            ...aiMessage,
            chatId
          })
        }
      }
    } catch (error) {
      console.error('Error sending message:', error)
    } finally {
      setLoading(false)

      // Плавное исчезновение индикатора с минимальной задержкой
      const elapsed = Date.now() - startTime
      const remaining = Math.max(0, TYPING_MIN_DURATION - elapsed)

      const finishTyping = () => {
        setTypingState('exiting')
        typingMinTimeRef.current = setTimeout(() => {
          setTypingState('idle')
        }, 300) // совпадает с длительностью typingFadeOut
      }

      if (remaining > 0) {
        typingMinTimeRef.current = setTimeout(finishTyping, remaining)
      } else {
        finishTyping()
      }
    }
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSendMessage()
    }
  }

  const handleVoiceTranscript = (text: string) => {
    setInputMessage(text)
  }

  const handleChatTypeChange = async (chatType: string) => {
    setCurrentChatType(chatType as any)
    // Загружаем локальную историю из IndexedDB и дожидаемся
    await loadChatHistory(chatType)

    // Загружаем историю с сервера и мержим
    try {
      const response = await chatApi.getHistory({ chatType, limit: 100, offset: 0 })
      const serverMessages: any[] = response.data?.messages || []

      if (serverMessages.length > 0) {
        const store = useAppStore.getState()
        const existing = store.chatHistories[chatType] || []

        // Дедупликация: пропускаем сообщения, которые уже есть (по id, content_hash или содержимому)
        const existingIds = new Set(existing.map((m: any) => m.id))
        const existingHashes = new Set(existing.map((m: any) => m.contentHash).filter(Boolean))
        // Простая сигнатура для дедупликации по содержимому: content|role|timestamp (округляем до минуты)
        const existingSignatures = new Set(
          existing.map((m: any) => `${m.content}|${m.role}|${m.timestamp?.slice(0, 16)}`)
        )

        const newMessages = serverMessages
          .filter((m: any) => {
            if (existingIds.has(m.id)) return false
            if (m.content_hash && existingHashes.has(m.content_hash)) return false
            const sig = `${m.content}|${m.role}|${(m.created_at || m.timestamp || '').slice(0, 16)}`
            if (existingSignatures.has(sig)) return false
            return true
          })
          .map((m: any) => ({
            id: m.id || Date.now() + Math.random(),
            userId: m.user_id || user?.id || 0,
            chatType: m.chat_type || chatType,
            content: m.content,
            isFromAI: m.role === 'ai' || m.is_from_ai,
            role: m.role || (m.is_from_ai ? 'ai' : 'user'),
            timestamp: m.created_at || m.timestamp || new Date().toISOString(),
          }))

        if (newMessages.length > 0) {
          // Сортируем по timestamp и добавляем
          const merged = [...existing, ...newMessages].sort(
            (a: any, b: any) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
          )
          setChatHistory(chatType, merged)
          console.log(`🔄 Merged ${newMessages.length} server messages for ${chatType}`)
        }
      }
    } catch (err) {
      console.error('Failed to load server chat history:', err)
    }
  }

  const loadLocalChat = async (chatId: string) => {
    try {
      const msgs = await chatSyncService.getLocalMessages(chatId)
      setLocalMessages(msgs)
      setShowLocalHistory(true)
      const parts = chatId.split('_')
      if (parts.length >= 2) {
        const chatType = parts[1]
        setCurrentChatType(chatType as any)
        loadChatHistory(chatType)
      }
    } catch (err) {
      console.error('Failed to load local chat:', err)
    }
  }

  // Применение темы из палитры store
  const applyTheme = (themeName: string) => {
    const selectedTheme = themes[themeName]
    if (selectedTheme) {
      setTheme(selectedTheme)
    }
  }

  // Список названий тем для отображения в сетке
  const themeNames = Object.keys(themes)

  // --- Удаление сообщений ---
  const toggleMessageSelection = (messageId: number) => {
    setSelectedMessages(prev => {
      const next = new Set(prev)
      if (next.has(messageId)) {
        next.delete(messageId)
      } else {
        next.add(messageId)
      }
      return next
    })
  }

  const handleDeleteSelected = async () => {
    if (selectedMessages.size === 0) return
    try {
      const messageIds = Array.from(selectedMessages)
      await chatApi.deleteMessages({ message_ids: messageIds })
      // Удаляем из локального стора
      const store = useAppStore.getState()
      const updatedHistory = currentChatHistory.filter(m => !selectedMessages.has(m.id))
      store.setChatHistory(currentChatType, updatedHistory)
      // Удаляем из IndexedDB
      await chatSyncService.deleteLocalMessages(messageIds)
      setSelectedMessages(new Set())
      setDeleteMode(false)
      setShowDeleteConfirm(false)
    } catch (err) {
      console.error('Failed to delete messages:', err)
      alert('Не удалось удалить сообщения. Попробуйте позже.')
    }
  }

  const handleClearHistory = async () => {
    try {
      await chatApi.clearHistory({ chat_type: currentChatType })
      const store = useAppStore.getState()
      store.setChatHistory(currentChatType, [])
      // Удаляем из IndexedDB
      if (user) {
        const chatId = `${user.id}_${currentChatType}`
        await chatSyncService.clearLocalChatMessages(chatId)
      }
      setDeleteMode(false)
      setShowClearConfirm(false)
    } catch (err) {
      console.error('Failed to clear history:', err)
      alert('Не удалось очистить историю. Попробуйте позже.')
    }
  }

  return (
    <div 
      className="flex h-screen"
      style={{ background: `linear-gradient(135deg, ${theme.bgFrom}, ${theme.bgTo})` }}
    >
      {/* Sidebar */}
      <div className={`${
        sidebarOpen ? 'w-64' : 'w-16'
      } transition-all duration-300 flex flex-col shadow-xl`}
        style={{ backgroundColor: theme.sidebarBg, backdropFilter: 'blur(16px)' }}
      >
        <div className="p-4" style={{ borderBottom: `1px solid ${theme.borderLight}` }}>
          <h1 
            className={`font-bold text-2xl ${sidebarOpen ? 'block' : 'hidden'}`}
            style={{ color: theme.textPrimary }}
          >
            MentalChat
          </h1>
          <button 
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className="mt-2 p-2 rounded-full hover-bg-theme-light transition-colors"
            style={{ color: theme.textSecondary }}
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </button>
        </div>

        <div className="flex-1 overflow-y-auto p-2">
          <div className="mb-4">
            <h2 
              className={`text-xs font-semibold uppercase tracking-wider mb-2 ${sidebarOpen ? 'block' : 'hidden'}`}
              style={{ color: theme.textMuted }}
            >
              Чаты
            </h2>
            {chatSessions.map((session) => {
              const iconEl = (className: string) => {
                switch (session.icon) {
                  case 'brain': return <BrainIcon className={className} />
                  case 'cards': return <CardsIcon className={className} />
                  case 'strawberry': return <StrawberryIcon className={className} />
                  case 'crystal-ball': return <MysticOrbIcon className={className} />
                  default: return null
                }
              }

              return (
                <button
                  key={session.id}
                  onClick={() => handleChatTypeChange(session.chatType)}
                  title={sidebarOpen ? undefined : session.title}
                  className={`w-full rounded-xl mb-2 transition-all flex items-center gap-3 ${
                    sidebarOpen ? 'text-left p-3' : 'justify-center p-2'
                  } ${
                    currentChatType === session.chatType
                      ? 'sidebar-active shadow-lg'
                      : 'hover-bg-theme-light'
                  }`}
                >
                  <span className="flex-shrink-0 w-6 h-6" style={{ color: theme.primary }}>
                    {iconEl('w-6 h-6')}
                  </span>
                  {sidebarOpen && (
                    <div className="min-w-0 flex-1">
                      <div className="font-medium truncate" style={{ color: theme.textPrimary }}>{session.title}</div>
                      <div className="text-xs opacity-75 truncate" style={{ color: theme.textSecondary }}>{session.lastMessage}</div>
                    </div>
                  )}
                </button>
              )
            })}
          </div>

          {/* Локальная история */}
          {user && localChats.length > 0 && (
            <div className="mb-4">
              <button
                onClick={() => { if (sidebarOpen) setShowLocalHistory(!showLocalHistory) }}
                title={sidebarOpen ? undefined : 'Локальная история'}
                className={`w-full mb-2 transition-all flex items-center gap-2 ${
                  sidebarOpen ? 'text-left p-2' : 'justify-center p-2'
                }`}
                style={{ color: theme.textSecondary }}
              >
                <span className="flex-shrink-0 w-6 h-6">
                  {sidebarOpen ? <OpenBookIcon className="w-6 h-6" /> : <ClosedBookIcon className="w-6 h-6" />}
                </span>
                {sidebarOpen && (
                  <span className="text-xs font-semibold uppercase tracking-wider">
                    Локальная история ({localChats.length})
                  </span>
                )}
              </button>
              
              {showLocalHistory && (
                <div className="max-h-64 overflow-y-auto">
                  {localChats.map((chat) => (
                    <button
                      key={chat.id}
                      onClick={() => loadLocalChat(chat.id)}
                      className={`w-full text-left p-3 rounded-xl mb-2 transition-all ${
                        chat.id === currentChatType
                          ? 'hover-bg-theme border-2'
                          : 'hover-bg-theme-light'
                      }`}
                      style={chat.id === currentChatType ? { borderColor: theme.primary } : {}}
                    >
                      <div className="font-medium text-sm" style={{ color: theme.textPrimary }}>
                        {chat.chatType === 'psychologist' && 'Психолог'}
                        {chat.chatType === 'tarot' && 'Таролог'}
                        {chat.chatType === 'sexologist' && 'Сексолог'}
                        {chat.chatType === 'fortune_teller' && 'Гадалка'}
                      </div>
                      <div className="text-xs" style={{ color: theme.textSecondary }}>
                        {new Date(chat.lastMessageTime).toLocaleDateString()}
                      </div>
                      {!chat.synced && (
                        <span className="text-xs text-orange-500">
                          {' '}● Не синхронизировано
                        </span>
                      )}
                    </button>
                  ))}
                </div>
              )}
            </div>
          )}

        </div>

        {/* Настройки и профиль — прижаты к низу сайдбара, центрированы */}
        <div className="relative p-2 border-t" style={{ borderColor: theme.borderLight }}>
          <div className="flex flex-col items-center">
            {/* Кнопка профиля */}
            <button
              onClick={() => { if (sidebarOpen) setShowProfile(!showProfile) }}
              className="flex items-center justify-center gap-2 p-2 rounded-xl hover-bg-theme-light transition-colors mb-1"
              title={sidebarOpen ? undefined : 'Профиль'}
              style={{ color: theme.textSecondary }}
            >
              <span className="w-6 h-6 flex-shrink-0">
                <UserIcon className="w-6 h-6" />
              </span>
              {sidebarOpen && <span className="text-xs font-semibold uppercase tracking-wider">Профиль</span>}
            </button>

            {/* Кнопка настроек */}
            <button
              onClick={() => setShowSettings(!showSettings)}
              className="flex items-center justify-center gap-2 p-2 rounded-xl hover-bg-theme-light transition-colors"
              title="Настройки"
              style={{ color: theme.textSecondary }}
            >
              <svg className="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
              {sidebarOpen && <span className="text-xs font-semibold uppercase tracking-wider">Настройки</span>}
            </button>

            {/* Панель профиля */}
            {sidebarOpen && showProfile && user && (
              <div 
                ref={profilePanelRef}
                className="absolute bottom-full left-2 right-2 mb-2 p-4 backdrop-blur-lg rounded-2xl space-y-3 shadow-xl"
                style={{ backgroundColor: theme.surface }}
              >
                <h3 className="text-sm font-semibold" style={{ color: theme.textPrimary }}>
                  {user.displayName || 'Пользователь'}
                </h3>
                <div className="text-xs space-y-1" style={{ color: theme.textSecondary }}>
                  <p>Почта: {user.email}</p>
                  <p>
                    Тариф:{' '}
                    <span className={`inline-block px-2 py-0.5 rounded-full text-[10px] font-semibold ${
                      user.tier === 'pro' ? 'bg-purple-100 text-purple-700' :
                      user.tier === 'ultra' ? 'bg-amber-100 text-amber-700' :
                      'bg-gray-100 text-gray-600'
                    }`}>
                      {user.tier?.toUpperCase() || 'FREE'}
                    </span>
                  </p>
                </div>
                <div className="space-y-2 pt-1">
                  <button
                    onClick={() => {
                      setShowProfile(false)
                      navigate('/subscription')
                    }}
                    className="w-full py-2.5 rounded-xl font-semibold text-sm transition-colors"
                    style={{ backgroundColor: theme.primary, color: theme.textOnPrimary }}
                  >
                    Upgrade profile
                  </button>
                  <button
                    onClick={() => {
                      if (confirm('Вы уверены, что хотите удалить профиль? Это действие необратимо.')) {
                        // TODO: вызвать API удаления
                        localStorage.removeItem('token')
                        navigate('/login')
                      }
                    }}
                    className="w-full py-2.5 rounded-xl font-semibold text-sm transition-colors"
                    style={{ color: theme.dangerText }}
                    onMouseEnter={(e) => e.currentTarget.style.backgroundColor = theme.dangerHover}
                    onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'transparent'}
                  >
                    Удалить профиль
                  </button>
                </div>
              </div>
            )}

            {sidebarOpen && showSettings && (
              <div 
                ref={settingsPanelRef}
                className="absolute bottom-full left-2 right-2 mb-2 p-3 backdrop-blur-lg rounded-2xl space-y-3 shadow-xl"
                style={{ backgroundColor: theme.surface }}
              >
                <div>
                  <h3 className="text-xs font-semibold mb-2" style={{ color: theme.textSecondary }}>Тема оформления</h3>
                  <div className="grid grid-cols-4 gap-1.5">
                    {themeNames.map((name) => {
                      const t = themes[name]
                      return (
                        <div key={name} className="relative group">
                          <button
                            onClick={() => applyTheme(name)}
                            className={`w-full aspect-square rounded-xl border-2 transition-all hover:scale-110 active:scale-95 ${
                              theme.name === name
                                ? 'ring-2 ring-offset-1'
                                : 'border-transparent hover:border-gray-400'
                            }`}
                            style={{ 
                              backgroundColor: t.primary,
                              borderColor: theme.name === name ? theme.textPrimary : undefined,
                              '--tw-ring-color': theme.primary,
                            } as React.CSSProperties}
                            title={name}
                          />
                          {/* Tooltip с названием цвета */}
                          <div className="absolute -top-9 left-1/2 -translate-x-1/2 px-2 py-1 rounded-lg bg-gray-800 text-white text-[10px] whitespace-nowrap opacity-0 group-hover:opacity-100 active:opacity-100 transition-opacity pointer-events-none z-10">
                            {name}
                            <div className="absolute top-full left-1/2 -translate-x-1/2 border-4 border-transparent border-t-gray-800" />
                          </div>
                        </div>
                      )
                    })}
                  </div>
                </div>

                <button
                  onClick={() => {
                    localStorage.removeItem('token')
                    navigate('/login')
                  }}
                  className="w-full py-2.5 rounded-xl font-semibold text-sm transition-colors"
                  style={{ color: theme.dangerText }}
                  onMouseEnter={(e) => e.currentTarget.style.backgroundColor = theme.dangerHover}
                  onMouseLeave={(e) => e.currentTarget.style.backgroundColor = 'transparent'}
                >
                  Выйти
                </button>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Main content */}
      <div className="flex-1 flex flex-col relative">
      {/* Chat header */}
      <div 
        className="p-4 flex justify-between items-center"
        style={{ 
          backgroundColor: theme.headerBg, 
          backdropFilter: 'blur(16px)',
          borderBottom: `1px solid ${theme.borderLight}` 
        }}
      >
        <div>
          <h2 className="text-2xl font-bold capitalize" style={{ color: theme.textPrimary }}>
            {currentChatType === 'psychologist' && 'Психолог'}
            {currentChatType === 'tarot' && 'Таролог'}
            {currentChatType === 'sexologist' && 'Сексолог'}
            {currentChatType === 'fortune_teller' && 'Гадалка'}
          </h2>
          <p className="text-sm" style={{ color: theme.textSecondary }}>
            {deleteMode ? `Выбрано: ${selectedMessages.size}` : 'Специалист в режиме онлайн'}
          </p>
        </div>

        <div className="flex items-center gap-2">
          {/* Кнопка режима удаления */}
          <button
            onClick={() => {
              setDeleteMode(!deleteMode)
              setSelectedMessages(new Set())
              setShowDeleteConfirm(false)
              setShowClearConfirm(false)
            }}
            className={`p-2 rounded-full transition-colors ${
              deleteMode ? '' : 'hover-bg-theme-light'
            }`}
            title={deleteMode ? 'Отменить' : 'Удалить сообщения'}
            style={{
              color: deleteMode ? theme.dangerText : theme.textSecondary,
              backgroundColor: deleteMode ? theme.dangerHover : 'transparent',
            }}
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          </button>
          <button
            onClick={() => setShowSearch(!showSearch)}
            className="p-2 rounded-full hover-bg-theme-light transition-colors"
            title="Поиск по сообщениям"
            style={{ color: theme.primary }}
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </button>
        </div>
      </div>

      {/* Search bar */}
      {showSearch && (
        <div 
          className="p-4 backdrop-blur-lg"
          style={{ 
            backgroundColor: theme.surface,
            borderBottom: `1px solid ${theme.borderLight}` 
          }}
        >
          <div className="flex space-x-2">
            <input
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="Поиск по сообщениям..."
              className="flex-1 px-4 py-2 rounded-full focus-theme outline-none transition-all"
              style={{ 
                backgroundColor: theme.inputBg, 
                border: `1px solid ${theme.inputBorder}`,
                color: theme.textPrimary,
              }}
            />
            <button
              onClick={() => {
                setShowSearch(false)
              }}
              className="px-4 py-2 rounded-full btn-theme transition-colors"
            >
              Найти
            </button>
          </div>
        </div>
      )}

        {/* Chat area */}
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {showLocalHistory && localMessages.length > 0 ? (
            <>
              <div className="text-center mb-4">
                <span className="text-xs px-3 py-1 rounded-full" style={{ backgroundColor: theme.surface, color: theme.textMuted }}>
                  📜 Локальная история
                </span>
              </div>
              {localMessages.map((message) => (
                <div
                  key={message.id}
                  className={`flex ${message.role === 'user' ? 'justify-end' : 'justify-start'}`}
                >
                  <div
                    className={`max-w-md px-6 py-3 rounded-3xl ${
                      message.role === 'user'
                        ? 'chat-user-msg'
                        : ''
                    }`}
                    style={{
                      ...(message.role === 'user'
                        ? { color: theme.textPrimary }
                        : { backgroundColor: theme.surface, color: theme.textPrimary }),
                      boxShadow: `0 0 18px 8px ${theme.messageBubble}`,
                    }}
                  >
                    <div>{message.content}</div>
                    <div className="text-xs opacity-75 mt-1">
                      {new Date(message.timestamp).toLocaleTimeString()}
                      {!message.synced && ' ⚠️'}
                    </div>
                  </div>
                </div>
              ))}
            </>
          ) : (
            currentChatHistory.map((message) => (
              <div
                key={message.id}
                className={`flex ${message.role === 'user' ? 'justify-end' : 'justify-start'} ${deleteMode ? 'cursor-pointer' : ''}`}
                onClick={() => {
                  if (deleteMode) toggleMessageSelection(message.id)
                }}
              >
                <div className={`flex items-center gap-2 ${message.role === 'user' ? 'flex-row-reverse' : 'flex-row'}`}>
                  {deleteMode && (
                    <div
                      className={`w-5 h-5 rounded-full border-2 flex items-center justify-center flex-shrink-0 transition-all ${
                        selectedMessages.has(message.id)
                          ? ''
                          : 'opacity-40'
                      }`}
                      style={{
                        borderColor: selectedMessages.has(message.id) ? theme.dangerText : theme.textMuted,
                        backgroundColor: selectedMessages.has(message.id) ? theme.dangerText : 'transparent',
                      }}
                    >
                      {selectedMessages.has(message.id) && (
                        <svg className="w-3 h-3 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
                        </svg>
                      )}
                    </div>
                  )}
                  <div
                    className={`max-w-md px-6 py-3 rounded-3xl transition-all ${
                      message.role === 'user'
                        ? 'chat-user-msg'
                        : ''
                    } ${
                      deleteMode && selectedMessages.has(message.id)
                        ? 'ring-2'
                        : ''
                    }`}
                    style={{
                      ...(message.role === 'user'
                        ? { color: theme.textPrimary }
                        : { backgroundColor: theme.surface, color: theme.textPrimary }),
                      boxShadow: `0 0 18px 8px ${theme.messageBubble}`,
                      ...(deleteMode && selectedMessages.has(message.id) ? {
                        boxShadow: `0 0 0 2px ${theme.dangerText}`,
                      } : {}),
                    }}
                  >
                    <div className="text-xs opacity-60 mb-1">
                      {new Date(message.timestamp).toLocaleTimeString()}
                    </div>
                    {message.content}
                  </div>
                </div>
              </div>
            ))
          )}
          {typingState !== 'idle' && (
            <div className="flex justify-start">
              <div 
                className={`flex items-center gap-3 px-5 py-3 rounded-2xl ${
                  typingState === 'entering' ? 'typing-enter' :
                  typingState === 'exiting' ? 'typing-exit' : ''
                }`}
                style={{ 
                  backgroundColor: theme.surface,
                  color: theme.textSecondary,
                  boxShadow: `0 0 18px 8px ${theme.messageBubble}`,
                }}
              >
                <span className="text-sm font-medium">Печатает</span>
                <div className="flex items-center gap-1">
                  <span className="typing-dot w-1.5 h-1.5 rounded-full" style={{ backgroundColor: theme.primary, animationDelay: '0ms' }} />
                  <span className="typing-dot w-1.5 h-1.5 rounded-full" style={{ backgroundColor: theme.primary, animationDelay: '200ms' }} />
                  <span className="typing-dot w-1.5 h-1.5 rounded-full" style={{ backgroundColor: theme.primary, animationDelay: '400ms' }} />
                </div>
              </div>
            </div>
          )}
          <div ref={messagesEndRef} />
        </div>

        {/* Панель действий удаления */}
        {deleteMode && (
          <div
            className="px-4 py-3 flex items-center justify-between backdrop-blur-lg"
            style={{
              backgroundColor: theme.headerBg,
              borderTop: `1px solid ${theme.borderLight}`,
            }}
          >
            <div className="flex items-center gap-2">
              <button
                onClick={() => {
                  setDeleteMode(false)
                  setSelectedMessages(new Set())
                  setShowDeleteConfirm(false)
                  setShowClearConfirm(false)
                }}
                className="px-4 py-2 rounded-full text-sm font-medium transition-colors hover-bg-theme-light"
                style={{ color: theme.textSecondary }}
              >
                Отмена
              </button>
              {selectedMessages.size > 0 && (
                <button
                  onClick={() => setShowDeleteConfirm(true)}
                  className="px-4 py-2 rounded-full text-sm font-medium transition-colors"
                  style={{
                    backgroundColor: theme.dangerText,
                    color: '#ffffff',
                  }}
                >
                  Удалить выбранные ({selectedMessages.size})
                </button>
              )}
            </div>
            <button
              onClick={() => {
                if (currentChatHistory.length === 0) {
                  alert('Нет сообщений для удаления')
                  return
                }
                setShowClearConfirm(true)
              }}
              className="px-4 py-2 rounded-full text-sm font-medium transition-colors"
              style={{
                backgroundColor: theme.dangerHover,
                color: theme.dangerText,
              }}
            >
              Очистить всю историю
            </button>
          </div>
        )}

        {/* Модальное окно подтверждения очистки истории */}
        {showClearConfirm && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
            <div
              className="mx-4 p-6 rounded-2xl shadow-2xl max-w-sm w-full"
              style={{ backgroundColor: theme.surface }}
            >
              <h3 className="text-lg font-bold mb-2" style={{ color: theme.textPrimary }}>
                Очистить историю?
              </h3>
              <p className="text-sm mb-6" style={{ color: theme.textSecondary }}>
                Все сообщения в чате «{currentChatType === 'psychologist' ? 'Психолог' : currentChatType === 'tarot' ? 'Таролог' : currentChatType === 'sexologist' ? 'Сексолог' : 'Гадалка'}» будут безвозвратно удалены.
              </p>
              <div className="flex gap-3">
                <button
                  onClick={() => setShowClearConfirm(false)}
                  className="flex-1 px-4 py-2.5 rounded-full text-sm font-medium transition-colors hover-bg-theme-light"
                  style={{ color: theme.textSecondary }}
                >
                  Отмена
                </button>
                <button
                  onClick={handleClearHistory}
                  className="flex-1 px-4 py-2.5 rounded-full text-sm font-medium transition-colors"
                  style={{
                    backgroundColor: theme.dangerText,
                    color: '#ffffff',
                  }}
                >
                  Очистить
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Модальное окно подтверждения удаления выбранных */}
        {showDeleteConfirm && (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
            <div
              className="mx-4 p-6 rounded-2xl shadow-2xl max-w-sm w-full"
              style={{ backgroundColor: theme.surface }}
            >
              <h3 className="text-lg font-bold mb-2" style={{ color: theme.textPrimary }}>
                Удалить сообщения?
              </h3>
              <p className="text-sm mb-6" style={{ color: theme.textSecondary }}>
                Выбрано сообщений: {selectedMessages.size}. Это действие нельзя отменить.
              </p>
              <div className="flex gap-3">
                <button
                  onClick={() => setShowDeleteConfirm(false)}
                  className="flex-1 px-4 py-2.5 rounded-full text-sm font-medium transition-colors hover-bg-theme-light"
                  style={{ color: theme.textSecondary }}
                >
                  Отмена
                </button>
                <button
                  onClick={handleDeleteSelected}
                  className="flex-1 px-4 py-2.5 rounded-full text-sm font-medium transition-colors"
                  style={{
                    backgroundColor: theme.dangerText,
                    color: '#ffffff',
                  }}
                >
                  Удалить
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Input area */}
        <div 
          className="p-4 backdrop-blur-lg"
          style={{ backgroundColor: theme.headerBg }}
        >
          <div className="flex items-center space-x-4">
            <button 
              className="p-3 rounded-full hover-bg-theme transition-colors"
              style={{ color: theme.primary }}
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13" />
              </svg>
            </button>
            
            <div className="flex-1 relative">
              <input
                type="text"
                value={inputMessage}
                onChange={(e) => setInputMessage(e.target.value)}
                onKeyPress={handleKeyPress}
                placeholder="Введите сообщение или используйте голосовой ввод..."
                className="w-full px-6 py-4 rounded-full focus-theme outline-none transition-all"
                style={{ 
                  backgroundColor: theme.inputBg, 
                  border: `1px solid ${theme.inputBorder}`,
                  color: theme.textPrimary,
                }}
              />
            </div>
            
            {/* Voice Input Button */}
            <VoiceInputButton
              onTranscript={handleVoiceTranscript}
              className="shadow-md"
            />
            
            <button
              onClick={handleSendMessage}
              disabled={!inputMessage.trim() || loading}
              className={`p-3 rounded-full transition-all transform hover:scale-105 ${
                inputMessage.trim() && !loading
                  ? 'btn-theme shadow-lg'
                  : ''
              }`}
              style={!inputMessage.trim() || loading ? { backgroundColor: theme.border, color: theme.textMuted, cursor: 'not-allowed' } : {}}
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
              </svg>
            </button>
          </div>
        </div>

        {/* Незаметный индикатор синхронизации в правом нижнем углу */}
        {user && (
          <div 
            className="absolute bottom-20 right-4 flex items-center gap-1.5 px-2 py-1 rounded-lg backdrop-blur-sm text-[10px] pointer-events-none select-none"
            style={{ backgroundColor: theme.sidebarBg, color: theme.textMuted }}
          >
            <span 
              className={`w-1.5 h-1.5 rounded-full ${unsyncedCount > 0 ? 'bg-orange-400' : 'bg-green-400'}`}
            />
            {lastSyncTime 
              ? `Синхр. ${lastSyncTime.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`
              : 'Авто-синхронизация'}
            {unsyncedCount > 0 && ` · ${unsyncedCount}`}
          </div>
        )}
      </div>
    </div>
  )
}

export default DashboardPage