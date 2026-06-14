import axios from 'axios'

const API_BASE_URL = '/api/v1'

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor for adding auth token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor for handling errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      const token = localStorage.getItem('token')
      if (token) {
        localStorage.removeItem('token')
        // Мягкое событие вместо жёсткого редиректа — страницы сами решат как реагировать
        window.dispatchEvent(new CustomEvent('auth:unauthorized'))
      }
    }
    return Promise.reject(error)
  }
)

export const authApi = {
  register: (data: {
    email: string
    password: string
    displayName: string
    mentalState: string
    marketingEmail: boolean
    privacyPolicy: boolean
    fingerprint?: string
    fingerprint_data?: any
  }) => {
    // Преобразуем camelCase в snake_case для backend
    const requestData = {
      email: data.email,
      password: data.password,
      display_name: data.displayName,
      mental_state: data.mentalState,
      marketing_email: data.marketingEmail,
      privacy_policy: data.privacyPolicy,
      fingerprint: data.fingerprint,
      fingerprint_data: data.fingerprint_data,
    }
    return api.post('/auth/register', requestData)
  },
  
  login: (data: { email: string; password: string }) => api.post('/auth/login', data),
  
  verifyEmail: (data: { email: string; token: string }) => api.post('/auth/verify-email', data),
  
  requestPasswordReset: (data: { email: string }) => api.post('/auth/request-password-reset', data),
  
  resetPassword: (data: { token: string; newPassword: string }) => api.post('/auth/reset-password', data),
  
  getTrialInfo: () => api.get('/config/trial'),
}

export const userApi = {
  getProfile: () => api.get('/user/profile'),
  
  updateProfile: (data: {
    displayName?: string
    mentalState?: string
    marketingEmail?: boolean
    microphonePerm?: boolean
  }) => api.put('/user/profile', data),
  
  deleteProfile: () => api.delete('/user/profile'),
  
  getChatSessions: () => api.get('/user/chat-sessions'),
  
  archiveChatSession: (sessionId: number) => api.post('/user/chat-sessions/archive', { session_id: sessionId }),
  
  saveMicrophonePermission: (granted: boolean) => api.post('/user/microphone-permission', { granted }),
}

export const chatApi = {
  sendMessage: (data: { chatType: string; content: string }) => api.post('/chat', {
    chat_type: data.chatType,
    content: data.content,
  }),
  
  getHistory: (data: { chatType: string; limit?: number; offset?: number }) => api.post('/chat/history', {
    chat_type: data.chatType,
    limit: data.limit,
    offset: data.offset,
  }),
  
  searchMessages: (data: { chatType: string; searchTerm: string }) => api.post('/chat/search', {
    chat_type: data.chatType,
    search_term: data.searchTerm,
  }),
  
  syncMessages: (data: { messages: Array<{
    chatId: string
    content: string
    isFromAI: boolean
    role: 'user' | 'ai'
    timestamp: string
    localId: number
  }> }) => api.post('/chat/sync', data),

  deleteMessages: (data: { message_ids: number[] }) => api.post('/chat/messages/delete', data),

  clearHistory: (data: { chat_type: string }) => api.post('/chat/history/clear', data),
}

export const paymentApi = {
  initiatePayment: (data: { tier: string; paymentType: string }) => api.post('/payment/initiate', data),
}

export const voiceApi = {
  transcribe: (formData: FormData) => api.post('/voice/transcribe', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  }),
}

export default api
