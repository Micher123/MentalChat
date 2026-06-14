import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import useAppStore from '../store/useAppStore'
import { authApi } from '../services/api'
import useFingerprint from '../hooks/useFingerprint'

const RegistrationPage = () => {
  const navigate = useNavigate()
  const { setUser, setAuthenticated, setMarketingEmail, setPrivacyPolicy, setMicrophonePerm, setToken, theme } = useAppStore()
  const { fingerprint, fingerprintData, loading: fpLoading } = useFingerprint()

  const [step, setStep] = useState(1)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  
  // Step 1 form state
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [passwordStrength, setPasswordStrength] = useState(0)
  const [passwordError, setPasswordError] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [mentalState, setMentalState] = useState('harmony')
  const [marketingEmail, setMarketingEmailLocal] = useState(true)
  const [privacyPolicy, setPrivacyPolicyLocal] = useState(false)
  
  // Step 2 form state
  const [selectedTier, setSelectedTier] = useState<string>('free')
  const [trialInfo, setTrialInfo] = useState<any>(null)

  // Load trial info on mount
  useEffect(() => {
    authApi.getTrialInfo().then(response => {
      setTrialInfo(response.data.tiers)
    })
  }, [])

  const handleRegister = async () => {
    if (!privacyPolicy) {
      setError('Необходимо принять политику конфиденциальности')
      return
    }

    if (password !== confirmPassword) {
      setError('Пароли не совпадают')
      return
    }

    if (passwordStrength < 3) {
      setError('Пароль слишком слабый. Минимум 8 символов, заглавная буква, цифра и спецсимвол')
      return
    }

    setLoading(true)
    setError('')
    setPasswordError('')

    try {
      const response = await authApi.register({
        email,
        password,
        displayName,
        mentalState,
        marketingEmail,
        privacyPolicy,
        fingerprint,
        fingerprint_data: fingerprintData || undefined,
      })

      // Сохраняем токены
      setToken(response.data.token)
      localStorage.setItem('refresh_token', response.data.refresh_token || '')
      
      // Сохраняем пользователя
      setUser(response.data.user)
      setAuthenticated(true)
      setMarketingEmail(marketingEmail)
      setPrivacyPolicy(privacyPolicy)
      setMicrophonePerm(true) // Default to true for simplicity

      // Navigate to login or dashboard
      navigate('/dashboard')
    } catch (err: any) {
      setError(err.response?.data?.error || 'Ошибка регистрации')
    } finally {
      setLoading(false)
    }
  }

  const checkPasswordStrength = (pwd: string) => {
    let strength = 0
    const errors: string[] = []

    if (pwd.length >= 8) strength++
    else errors.push('минимум 8 символов')

    if (/[A-Z]/.test(pwd)) strength++
    else errors.push('заглавная буква')

    if (/[a-z]/.test(pwd)) strength++
    else errors.push('строчная буква')

    if (/[0-9]/.test(pwd)) strength++
    else errors.push('цифра')

    if (/[^A-Za-z0-9]/.test(pwd)) strength++
    else errors.push('спецсимвол')

    setPasswordStrength(strength)
    
    if (pwd.length > 0 && strength < 5) {
      setPasswordError(`Требуется: ${errors.join(', ')}`)
    } else {
      setPasswordError('')
    }

    return strength
  }

  const handleTierSelect = (tier: string) => {
    setSelectedTier(tier)
  }

  // Базовый стиль инпута
  const inputStyle = {
    backgroundColor: theme.inputBg,
    border: `1px solid ${theme.inputBorder}`,
    color: theme.textPrimary,
  }

  // Стиль активной кнопки ментального состояния
  const activeStateStyle = {
    backgroundColor: theme.primary,
    color: theme.textOnPrimary,
  }

  return (
    <div 
      className="min-h-screen flex items-center justify-center p-4"
      style={{ background: `linear-gradient(135deg, ${theme.bgFrom}, ${theme.bgTo})` }}
    >
      <div 
        className="rounded-3xl p-8 max-w-2xl w-full shadow-2xl backdrop-blur-xl"
        style={{ backgroundColor: theme.cardBg }}
      >
        <h1 
          className="text-4xl font-bold text-center mb-8"
          style={{ color: theme.textPrimary }}
        >
          MentalChat
        </h1>
        
        {step === 1 && (
          <div className="space-y-6">
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1" style={{ color: theme.textSecondary }}>Почта</label>
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full px-4 py-3 rounded-xl outline-none transition-all"
                  style={inputStyle}
                  placeholder="your@email.com"
                />
              </div>
              
              <div>
                <label className="block text-sm font-medium mb-1" style={{ color: theme.textSecondary }}>Пароль</label>
                <input
                  type="password"
                  value={password}
                  onChange={(e) => {
                    setPassword(e.target.value)
                    checkPasswordStrength(e.target.value)
                  }}
                  className="w-full px-4 py-3 rounded-xl outline-none transition-all"
                  style={inputStyle}
                  placeholder="••••••••"
                  minLength={8}
                />
                {password.length > 0 && (
                  <div className="mt-2">
                    <div className="flex gap-1 mb-1">
                      {[1, 2, 3, 4, 5].map((level) => (
                        <div
                          key={level}
                          className={`h-1 flex-1 rounded-full transition-all ${
                            passwordStrength >= level
                              ? passwordStrength <= 2
                                ? 'bg-red-500'
                                : passwordStrength <= 3
                                ? 'bg-yellow-500'
                                : 'bg-green-500'
                              : 'bg-gray-300'
                          }`}
                        />
                      ))}
                    </div>
                    {passwordError && (
                      <p className="text-xs" style={{ color: theme.dangerText }}>{passwordError}</p>
                    )}
                  </div>
                )}
              </div>
              
              <div>
                <label className="block text-sm font-medium mb-1" style={{ color: theme.textSecondary }}>Подтвердите пароль</label>
                <input
                  type="password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  className="w-full px-4 py-3 rounded-xl outline-none transition-all"
                  style={{
                    ...inputStyle,
                    borderColor: confirmPassword && password !== confirmPassword
                      ? '#EF4444'
                      : confirmPassword && password === confirmPassword
                      ? '#22C55E'
                      : theme.inputBorder,
                  }}
                  placeholder="••••••••"
                  minLength={8}
                />
                {confirmPassword && password === confirmPassword && (
                  <p className="text-xs text-green-500 mt-1">✓ Пароли совпадают</p>
                )}
                {confirmPassword && password !== confirmPassword && (
                  <p className="text-xs mt-1" style={{ color: theme.dangerText }}>✗ Пароли не совпадают</p>
                )}
              </div>
              
              <div>
                <label className="block text-sm font-medium mb-1" style={{ color: theme.textSecondary }}>Имя/Ник</label>
                <input
                  type="text"
                  value={displayName}
                  onChange={(e) => setDisplayName(e.target.value)}
                  className="w-full px-4 py-3 rounded-xl outline-none transition-all"
                  style={inputStyle}
                  placeholder="Ваше имя"
                />
              </div>
              
              <div>
                <label className="block text-sm font-medium mb-1" style={{ color: theme.textSecondary }}>Текущее ментальное состояние</label>
                <div className="grid grid-cols-2 gap-3">
                  {[
                    { value: 'harmony', label: 'В гармонии' },
                    { value: 'satisfied', label: 'Удовлетворена' },
                    { value: 'anxiety', label: 'Тревога' },
                    { value: 'stress', label: 'Стресс' },
                  ].map((state) => (
                    <button
                      key={state.value}
                      type="button"
                      onClick={() => setMentalState(state.value)}
                      className={`px-4 py-3 rounded-xl transition-all`}
                      style={
                        mentalState === state.value
                          ? activeStateStyle
                          : { backgroundColor: theme.surface, color: theme.textSecondary }
                      }
                    >
                      {state.label}
                    </button>
                  ))}
                </div>
              </div>
            </div>
            
            <div className="space-y-3">
              <label className="flex items-start space-x-3 cursor-pointer">
                <input
                  type="checkbox"
                  checked={marketingEmail}
                  onChange={(e) => setMarketingEmailLocal(e.target.checked)}
                  className="mt-1 w-5 h-5 rounded"
                  style={{ accentColor: theme.primary }}
                />
                <span className="text-sm" style={{ color: theme.textSecondary }}>Я хочу получать рассылку о новых возможностях и скидках</span>
              </label>
              
              <label className="flex items-start space-x-3 cursor-pointer">
                <input
                  type="checkbox"
                  checked={privacyPolicy}
                  onChange={(e) => setPrivacyPolicyLocal(e.target.checked)}
                  className="mt-1 w-5 h-5 rounded"
                  style={{ accentColor: theme.primary }}
                />
                <span className="text-sm" style={{ color: theme.textSecondary }}>
                  Я согласен с <a href="#" className="hover:underline" style={{ color: theme.primary }}>Пользовательским соглашением</a> и <a href="#" className="hover:underline" style={{ color: theme.primary }}>Политикой конфиденциальности</a>
                </span>
              </label>
            </div>
            
            {error && (
              <div className="text-sm text-center" style={{ color: theme.dangerText }}>{error}</div>
            )}
            
            <button
              onClick={handleRegister}
              disabled={loading || !privacyPolicy || fpLoading}
              className={`w-full py-4 rounded-xl font-semibold transition-all transform hover:scale-105 ${
                loading || !privacyPolicy || fpLoading
                  ? 'cursor-not-allowed'
                  : 'shadow-lg hover:shadow-xl'
              }`}
              style={loading || !privacyPolicy || fpLoading
                ? { backgroundColor: theme.border, color: theme.textMuted }
                : { background: `linear-gradient(135deg, ${theme.primary}, ${theme.accent})`, color: theme.textOnPrimary }
              }
            >
              {fpLoading ? 'Сбор данных устройства...' : loading ? 'Регистрация...' : 'Регистрация'}
            </button>
            
            <p className="text-center text-sm" style={{ color: theme.textSecondary }}>
              Уже есть аккаунт? <a href="/login" className="hover:underline" style={{ color: theme.primary }}>Войти</a>
            </p>
          </div>
        )}
        
        {step === 2 && trialInfo && (
          <div className="space-y-6">
            <h2 className="text-2xl font-bold text-center" style={{ color: theme.textPrimary }}>Выберите тариф</h2>
            
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {trialInfo.map((tier: any) => (
                <div
                  key={tier.tier}
                  onClick={() => handleTierSelect(tier.tier)}
                  className={`cursor-pointer rounded-2xl p-6 transition-all transform hover:scale-105`}
                  style={
                    selectedTier === tier.tier
                      ? { background: `linear-gradient(135deg, ${theme.primary}, ${theme.accent})`, color: theme.textOnPrimary }
                      : { backgroundColor: theme.surface, color: theme.textPrimary }
                  }
                >
                  <h3 className="text-xl font-bold mb-2">{tier.name}</h3>
                  <p className="mb-4 opacity-90">{tier.description}</p>
                  <div className="mb-4">
                    <span className="text-3xl font-bold">{tier.price}</span>
                  </div>
                  {tier.trial_days && (
                    <p className="text-sm mb-4 opacity-75">
                      {tier.trial_days}-дневный бесплатный пробный период
                    </p>
                  )}
                  <ul className="space-y-2 text-sm">
                    {tier.features.map((feature: string, index: number) => (
                      <li key={index} className="flex items-center">
                        <span className="mr-2">✓</span>
                        {feature}
                      </li>
                    ))}
                  </ul>
                </div>
              ))}
            </div>
            
            <button
              onClick={() => setStep(1)}
              className="w-full py-3 rounded-xl font-semibold transition-all"
              style={{ backgroundColor: theme.surface, color: theme.textSecondary }}
            >
              Назад
            </button>
          </div>
        )}
      </div>
    </div>
  )
}

export default RegistrationPage