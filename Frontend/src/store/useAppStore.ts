import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface User {
  id: number
  email: string
  displayName: string
  mentalState?: string
  tier?: 'free' | 'pro' | 'ultra'
  verified?: boolean
  marketingEmail?: boolean
  microphonePerm?: boolean
  avatarURL?: string
}

interface ChatMessage {
  id: number
  userId: number
  chatType: 'psychologist' | 'tarot' | 'sexologist' | 'fortune_teller'
  content: string
  isFromAI: boolean
  role: 'user' | 'ai'
  timestamp: string
}

interface Theme {
  name: string
  // Градиент фона страницы
  bgFrom: string
  bgTo: string
  // Основные цвета
  primary: string
  primaryLight: string
  primaryDark: string
  // Поверхности
  surface: string
  surfaceHover: string
  sidebarBg: string
  // Текст
  textPrimary: string
  textSecondary: string
  textOnPrimary: string
  textMuted: string
  // Границы и инпуты
  border: string
  borderLight: string
  inputBg: string
  inputBorder: string
  // Акцент и хедер
  accent: string
  headerBg: string
  // Специальные
  cardBg: string
  dangerText: string
  dangerHover: string
  // Ореол вокруг сообщений в чате
  messageBubble: string
}

interface AppState {
  user: User | null
  setUser: (user: User) => void
  clearUser: () => void
  
  isAuthenticated: boolean
  setAuthenticated: (authenticated: boolean) => void
  token: string | null
  setToken: (token: string) => void
  clearAuth: () => void
  
  currentChatType: 'psychologist' | 'tarot' | 'sexologist' | 'fortune_teller'
  setCurrentChatType: (chatType: 'psychologist' | 'tarot' | 'sexologist' | 'fortune_teller') => void
  chatHistories: Record<string, ChatMessage[]>
  addMessage: (message: ChatMessage) => void
  setChatHistory: (chatType: string, messages: ChatMessage[]) => void
  clearChat: () => void
  
  theme: Theme
  setTheme: (theme: Theme) => void
  microphonePerm: boolean
  setMicrophonePerm: (perm: boolean) => void
  marketingEmail: boolean
  setMarketingEmail: (email: boolean) => void
  privacyPolicy: boolean
  setPrivacyPolicy: (policy: boolean) => void
  
  sidebarOpen: boolean
  toggleSidebar: () => void
  setSidebarOpen: (open: boolean) => void
  welcomeModalOpen: boolean
  setWelcomeModalOpen: (open: boolean) => void
}

// ======== НЕЖНЫЕ СВЕТЛЫЕ ПАЛИТРЫ ДЛЯ ЖЕНСКОГО СЕРВИСА ========
// Пастельные, тёплые, женственные. messageBubble = полупрозрачный ореол темы

const themes: Record<string, Theme> = {
  // ---- ДЕФОЛТ: Песочная (Oat) - кремово-карамельная нежность ----
  'Oat': {
    name: 'Oat',
    bgFrom: '#FEF9F0',
    bgTo: '#F5ECD7',
    primary: '#C8A87C',
    primaryLight: '#DEC9A8',
    primaryDark: '#A68452',
    surface: 'rgba(255,255,255,0.55)',
    surfaceHover: 'rgba(200,168,124,0.10)',
    sidebarBg: 'rgba(255,252,245,0.50)',
    textPrimary: '#3E2E1F',
    textSecondary: '#5C4530',
    textOnPrimary: '#FFFFFF',
    textMuted: '#8C7A63',
    border: '#D9C9A8',
    borderLight: 'rgba(200,168,124,0.25)',
    inputBg: 'rgba(255,255,255,0.65)',
    inputBorder: '#E0CEA8',
    accent: '#E8C878',
    headerBg: 'rgba(255,252,245,0.40)',
    cardBg: 'rgba(255,255,255,0.60)',
    dangerText: '#D48484',
    dangerHover: 'rgba(212,132,132,0.08)',
    messageBubble: 'rgba(180,144,100,0.28)',
  },
  // ---- Розовый лепесток (Rose Petal) ----
  'Rose Petal': {
    name: 'Rose Petal',
    bgFrom: '#FFF0F3',
    bgTo: '#FADDE2',
    primary: '#D4919A',
    primaryLight: '#E8B8BE',
    primaryDark: '#B06D76',
    surface: 'rgba(255,255,255,0.55)',
    surfaceHover: 'rgba(212,145,154,0.10)',
    sidebarBg: 'rgba(255,245,246,0.50)',
    textPrimary: '#3D1E24',
    textSecondary: '#5C3238',
    textOnPrimary: '#FFFFFF',
    textMuted: '#9B7A80',
    border: '#E4C0C6',
    borderLight: 'rgba(212,145,154,0.22)',
    inputBg: 'rgba(255,255,255,0.65)',
    inputBorder: '#EDD0D6',
    accent: '#E8A0AA',
    headerBg: 'rgba(255,245,246,0.40)',
    cardBg: 'rgba(255,255,255,0.60)',
    dangerText: '#D48484',
    dangerHover: 'rgba(212,132,132,0.08)',
    messageBubble: 'rgba(190,120,135,0.28)',
  },
  // ---- Лавандовый туман (Lavender Mist) ----
  'Lavender Mist': {
    name: 'Lavender Mist',
    bgFrom: '#F5F0FF',
    bgTo: '#E8DEF5',
    primary: '#B09AC4',
    primaryLight: '#D0C0E0',
    primaryDark: '#8B72A5',
    surface: 'rgba(255,255,255,0.55)',
    surfaceHover: 'rgba(176,154,196,0.10)',
    sidebarBg: 'rgba(250,248,255,0.50)',
    textPrimary: '#2D1A3E',
    textSecondary: '#4A3060',
    textOnPrimary: '#FFFFFF',
    textMuted: '#8A7A9E',
    border: '#CFC0E0',
    borderLight: 'rgba(176,154,196,0.22)',
    inputBg: 'rgba(255,255,255,0.65)',
    inputBorder: '#DDD0E8',
    accent: '#C4B0D8',
    headerBg: 'rgba(250,248,255,0.40)',
    cardBg: 'rgba(255,255,255,0.60)',
    dangerText: '#D48484',
    dangerHover: 'rgba(212,132,132,0.08)',
    messageBubble: 'rgba(150,128,175,0.28)',
  },
  // ---- Персиковый сад (Peach Blossom) ----
  'Peach Blossom': {
    name: 'Peach Blossom',
    bgFrom: '#FFF4ED',
    bgTo: '#FDE5D6',
    primary: '#D49874',
    primaryLight: '#E8BEAA',
    primaryDark: '#B07654',
    surface: 'rgba(255,255,255,0.55)',
    surfaceHover: 'rgba(212,152,116,0.10)',
    sidebarBg: 'rgba(255,248,242,0.50)',
    textPrimary: '#3D2014',
    textSecondary: '#5C3422',
    textOnPrimary: '#FFFFFF',
    textMuted: '#9B7A68',
    border: '#E4C8B4',
    borderLight: 'rgba(212,152,116,0.22)',
    inputBg: 'rgba(255,255,255,0.65)',
    inputBorder: '#EDD6C8',
    accent: '#E8B098',
    headerBg: 'rgba(255,248,242,0.40)',
    cardBg: 'rgba(255,255,255,0.60)',
    dangerText: '#D48484',
    dangerHover: 'rgba(212,132,132,0.08)',
    messageBubble: 'rgba(190,130,95,0.28)',
  },
  // ---- Мятный шёлк (Mint Silk) ----
  'Mint Silk': {
    name: 'Mint Silk',
    bgFrom: '#F2FBF6',
    bgTo: '#DAF0E2',
    primary: '#84B898',
    primaryLight: '#AAD0B8',
    primaryDark: '#62987A',
    surface: 'rgba(255,255,255,0.55)',
    surfaceHover: 'rgba(132,184,152,0.10)',
    sidebarBg: 'rgba(245,252,248,0.50)',
    textPrimary: '#1C3A28',
    textSecondary: '#305840',
    textOnPrimary: '#FFFFFF',
    textMuted: '#6E8E78',
    border: '#B8D8C4',
    borderLight: 'rgba(132,184,152,0.22)',
    inputBg: 'rgba(255,255,255,0.65)',
    inputBorder: '#CAE4D2',
    accent: '#A0D0B0',
    headerBg: 'rgba(245,252,248,0.40)',
    cardBg: 'rgba(255,255,255,0.60)',
    dangerText: '#D48484',
    dangerHover: 'rgba(212,132,132,0.08)',
    messageBubble: 'rgba(105,160,128,0.28)',
  },
  // ---- Небесная лазурь (Sky Azure) ----
  'Sky Azure': {
    name: 'Sky Azure',
    bgFrom: '#F0F7FE',
    bgTo: '#D8E8FA',
    primary: '#8AAAce',
    primaryLight: '#B0C8E4',
    primaryDark: '#6488B0',
    surface: 'rgba(255,255,255,0.55)',
    surfaceHover: 'rgba(138,170,206,0.10)',
    sidebarBg: 'rgba(245,250,255,0.50)',
    textPrimary: '#1A2840',
    textSecondary: '#2E3E5C',
    textOnPrimary: '#FFFFFF',
    textMuted: '#6E7E9A',
    border: '#BCD0E8',
    borderLight: 'rgba(138,170,206,0.22)',
    inputBg: 'rgba(255,255,255,0.65)',
    inputBorder: '#CCDAEE',
    accent: '#A4C0E0',
    headerBg: 'rgba(245,250,255,0.40)',
    cardBg: 'rgba(255,255,255,0.60)',
    dangerText: '#D48484',
    dangerHover: 'rgba(212,132,132,0.08)',
    messageBubble: 'rgba(110,145,185,0.28)',
  },
  // ---- Золотистый песок (Golden Sand) ----
  'Golden Sand': {
    name: 'Golden Sand',
    bgFrom: '#FFFBF0',
    bgTo: '#F8F0D8',
    primary: '#D4B878',
    primaryLight: '#E8D4A0',
    primaryDark: '#B09450',
    surface: 'rgba(255,255,255,0.55)',
    surfaceHover: 'rgba(212,184,120,0.10)',
    sidebarBg: 'rgba(255,252,245,0.50)',
    textPrimary: '#3D2E14',
    textSecondary: '#5C4822',
    textOnPrimary: '#FFFFFF',
    textMuted: '#9B8A60',
    border: '#E0CEA0',
    borderLight: 'rgba(212,184,120,0.22)',
    inputBg: 'rgba(255,255,255,0.65)',
    inputBorder: '#E8D8B0',
    accent: '#E8D080',
    headerBg: 'rgba(255,252,245,0.40)',
    cardBg: 'rgba(255,255,255,0.60)',
    dangerText: '#D48484',
    dangerHover: 'rgba(212,132,132,0.08)',
    messageBubble: 'rgba(190,160,95,0.28)',
  },
}

const defaultTheme = themes['Oat']

const applyThemeToDocument = (theme: Theme) => {
  const root = document.documentElement
  root.style.setProperty('--bg-from', theme.bgFrom)
  root.style.setProperty('--bg-to', theme.bgTo)
  root.style.setProperty('--primary', theme.primary)
  root.style.setProperty('--primary-light', theme.primaryLight)
  root.style.setProperty('--primary-dark', theme.primaryDark)
  root.style.setProperty('--surface', theme.surface)
  root.style.setProperty('--surface-hover', theme.surfaceHover)
  root.style.setProperty('--sidebar-bg', theme.sidebarBg)
  root.style.setProperty('--text-primary', theme.textPrimary)
  root.style.setProperty('--text-secondary', theme.textSecondary)
  root.style.setProperty('--text-on-primary', theme.textOnPrimary)
  root.style.setProperty('--text-muted', theme.textMuted)
  root.style.setProperty('--border', theme.border)
  root.style.setProperty('--border-light', theme.borderLight)
  root.style.setProperty('--input-bg', theme.inputBg)
  root.style.setProperty('--input-border', theme.inputBorder)
  root.style.setProperty('--accent', theme.accent)
  root.style.setProperty('--header-bg', theme.headerBg)
  root.style.setProperty('--card-bg', theme.cardBg)
  root.style.setProperty('--danger-text', theme.dangerText)
  root.style.setProperty('--danger-hover', theme.dangerHover)
  root.style.setProperty('--message-bubble', theme.messageBubble)
}

const useAppStore = create<AppState>()(
  persist(
    (set, get) => ({
      user: null,
      setUser: (user) => set({ user }),
      clearUser: () => set({ user: null }),
      
      isAuthenticated: false,
      setAuthenticated: (authenticated) => set({ isAuthenticated: authenticated }),
      token: null,
      setToken: (token) => set({ token }),
      clearAuth: () => set({ isAuthenticated: false, token: null }),
      
      currentChatType: 'psychologist',
      setCurrentChatType: (chatType) => set({ currentChatType: chatType }),
      chatHistories: {},
      addMessage: (message) => {
        const histories = { ...get().chatHistories }
        const chatHistory = histories[message.chatType] || []
        histories[message.chatType] = [...chatHistory, message]
        set({ chatHistories: histories })
      },
      setChatHistory: (chatType, messages) => {
        const histories = { ...get().chatHistories }
        histories[chatType] = messages
        set({ chatHistories: histories })
      },
      clearChat: () => {
        const histories = { ...get().chatHistories }
        const currentChat = get().currentChatType
        histories[currentChat] = []
        set({ chatHistories: histories })
      },
      
      theme: defaultTheme,
      setTheme: (theme) => {
        applyThemeToDocument(theme)
        set({ theme })
      },
      microphonePerm: false,
      setMicrophonePerm: (perm) => set({ microphonePerm: perm }),
      marketingEmail: false,
      setMarketingEmail: (email) => set({ marketingEmail: email }),
      privacyPolicy: false,
      setPrivacyPolicy: (policy) => set({ privacyPolicy: policy }),
      
      sidebarOpen: true,
      toggleSidebar: () => set({ sidebarOpen: !get().sidebarOpen }),
      setSidebarOpen: (open) => set({ sidebarOpen: open }),
      welcomeModalOpen: true,
      setWelcomeModalOpen: (open) => set({ welcomeModalOpen: open }),
    }),
    {
      name: 'mentalchat-storage',
      partialize: (state) => ({
        token: state.token,
        isAuthenticated: state.isAuthenticated,
        user: state.user,
        theme: state.theme,
        microphonePerm: state.microphonePerm,
        marketingEmail: state.marketingEmail,
        privacyPolicy: state.privacyPolicy,
        sidebarOpen: state.sidebarOpen,
      }),
    }
  )
)

const initTheme = () => {
  const state = useAppStore.getState()
  if (state.theme) {
    applyThemeToDocument(state.theme)
  }
}

initTheme()

export default useAppStore
export { themes }