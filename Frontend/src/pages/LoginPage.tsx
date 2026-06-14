import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import useAppStore from '../store/useAppStore'
import { authApi } from '../services/api'

const LoginPage = () => {
  const navigate = useNavigate()
  const { setUser, setAuthenticated, setToken, theme } = useAppStore()

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const handleLogin = async () => {
    setLoading(true)
    setError('')

    try {
      const response = await authApi.login({ email, password })
      
      // Сохраняем токены
      setToken(response.data.token)
      localStorage.setItem('token', response.data.token)
      localStorage.setItem('refresh_token', response.data.refresh_token)
      
      // Сохраняем пользователя
      setUser(response.data.user)
      setAuthenticated(true)
      console.log('Установлены: токен, пользователь, authenticated=true');
      
      navigate('/dashboard')
      console.log('navigate вызван, переход на /dashboard');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Ошибка входа')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div 
      className="min-h-screen flex items-center justify-center p-4"
      style={{ background: `linear-gradient(135deg, ${theme.bgFrom}, ${theme.bgTo})` }}
    >
      <div 
        className="rounded-3xl p-8 max-w-md w-full shadow-2xl backdrop-blur-xl"
        style={{ backgroundColor: theme.cardBg }}
      >
        <h1 
          className="text-4xl font-bold text-center mb-8"
          style={{ color: theme.textPrimary }}
        >
          MentalChat
        </h1>
        
        <div className="space-y-6">
          <div className="space-y-4">
            <div>
              <label 
                className="block text-sm font-medium mb-1"
                style={{ color: theme.textSecondary }}
              >
                Почта
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full px-4 py-3 rounded-xl outline-none transition-all"
                style={{ 
                  backgroundColor: theme.inputBg,
                  border: `1px solid ${theme.inputBorder}`,
                  color: theme.textPrimary,
                }}
                placeholder="your@email.com"
              />
            </div>
            
            <div>
              <label 
                className="block text-sm font-medium mb-1"
                style={{ color: theme.textSecondary }}
              >
                Пароль
              </label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full px-4 py-3 rounded-xl outline-none transition-all"
                style={{ 
                  backgroundColor: theme.inputBg,
                  border: `1px solid ${theme.inputBorder}`,
                  color: theme.textPrimary,
                }}
                placeholder="••••••••"
              />
            </div>
          </div>
          
          {error && (
            <div className="text-sm text-center" style={{ color: theme.dangerText }}>
              {error}
            </div>
          )}
          
          <button
            onClick={handleLogin}
            disabled={loading}
            className={`w-full py-3 rounded-xl font-semibold transition-all transform hover:scale-105 ${
              loading
                ? 'cursor-not-allowed'
                : 'shadow-lg hover:shadow-xl'
            }`}
            style={loading 
              ? { backgroundColor: theme.border, color: theme.textMuted } 
              : { 
                  background: `linear-gradient(135deg, ${theme.primary}, ${theme.accent})`, 
                  color: theme.textOnPrimary 
                }
            }
          >
            {loading ? 'Вход...' : 'Войти'}
          </button>
          
          <div className="text-center space-y-2">
            <p className="text-sm" style={{ color: theme.textSecondary }}>
              Нет аккаунта?{' '}
              <a 
                href="/register" 
                className="hover:underline"
                style={{ color: theme.primary }}
              >
                Зарегистрироваться
              </a>
            </p>
            <p className="text-sm" style={{ color: theme.textSecondary }}>
              <a 
                href="/forgot-password" 
                className="hover:underline"
                style={{ color: theme.primary }}
              >
                Забыли пароль?
              </a>
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}

export default LoginPage