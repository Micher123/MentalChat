import { useState, type FC } from 'react'
import { useNavigate } from 'react-router-dom'
import useAppStore from '../store/useAppStore'

interface PlanFeature {
  text: string
  included: boolean
}

interface Plan {
  key: 'free' | 'pro' | 'ultra'
  name: string
  price: string
  priceNote?: string
  trialDays: number | null
  gradient: string
  borderColor: string
  chipBg: string
  chipText: string
  features: PlanFeature[]
  cta: string
}

const plans: Plan[] = [
  {
    key: 'free',
    name: 'FREE',
    price: '0 ₽',
    trialDays: null,
    gradient: 'from-gray-50 to-gray-100',
    borderColor: 'border-gray-300',
    chipBg: 'bg-gray-200',
    chipText: 'text-gray-600',
    features: [
      { text: 'Сексолог', included: true },
      { text: 'Психолог', included: true },
      { text: 'Таролог', included: false },
      { text: 'Гадание', included: false },
      { text: 'Гадание по фото', included: false },
    ],
    cta: 'Текущий тариф',
  },
  {
    key: 'pro',
    name: 'PRO',
    price: '499 ₽',
    priceNote: 'в месяц',
    trialDays: 3,
    gradient: 'from-purple-50 to-purple-100',
    borderColor: 'border-purple-300',
    chipBg: 'bg-purple-200',
    chipText: 'text-purple-700',
    features: [
      { text: 'Сексолог', included: true },
      { text: 'Психолог', included: true },
      { text: 'Таролог', included: true },
      { text: 'Гадание', included: false },
      { text: 'Гадание по фото', included: false },
    ],
    cta: 'Оформить PRO',
  },
  {
    key: 'ultra',
    name: 'ULTRA',
    price: '999 ₽',
    priceNote: 'в месяц',
    trialDays: 1,
    gradient: 'from-amber-50 to-amber-100',
    borderColor: 'border-amber-400',
    chipBg: 'bg-amber-200',
    chipText: 'text-amber-700',
    features: [
      { text: 'Сексолог', included: true },
      { text: 'Психолог', included: true },
      { text: 'Таролог', included: true },
      { text: 'Гадание', included: true },
      { text: 'Гадание по фото', included: true },
    ],
    cta: 'Оформить ULTRA',
  },
]

const SubscriptionPage: FC = () => {
  const navigate = useNavigate()
  const { user, theme, setUser } = useAppStore()
  const [loadingPlan, setLoadingPlan] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const currentTier = user?.tier || 'free'

  const handleUpgrade = async (planKey: string) => {
    if (planKey === 'free') return
    if (planKey === currentTier) {
      setError('У вас уже активирован этот тариф.')
      return
    }

    setLoadingPlan(planKey)
    setError(null)

    try {
      const token = localStorage.getItem('token')
      const response = await fetch('/api/v1/payment/initiate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ tier: planKey, payment_type: 'monthly' }),
      })

      if (!response.ok) {
        const errData = await response.json().catch(() => ({}))
        throw new Error(errData.error || 'Ошибка при создании платежа')
      }

      const data = await response.json()

      // Если есть URL для оплаты — редиректим пользователя
      if (data.payment_url) {
        window.location.href = data.payment_url
      } else {
        // fallback: показываем сообщение об успехе
        alert(`Тариф ${planKey.toUpperCase()} активирован! (Тестовый режим)`)
        setUser({ ...user!, tier: planKey as 'free' | 'pro' | 'ultra' })
        navigate('/dashboard')
      }
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Не удалось создать платёж')
    } finally {
      setLoadingPlan(null)
    }
  }

  return (
    <div
      className="min-h-screen flex flex-col items-center justify-center p-6"
      style={{
        background: `linear-gradient(135deg, ${theme.bgFrom}, ${theme.bgTo})`,
      }}
    >
      {/* Кнопка назад */}
      <button
        onClick={() => navigate('/dashboard')}
        className="absolute top-4 left-4 p-2 rounded-full hover:opacity-80 transition-opacity"
        style={{ color: theme.textSecondary }}
      >
        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
        </svg>
      </button>

      <div className="max-w-5xl w-full">
        <h1
          className="text-3xl font-bold text-center mb-2"
          style={{ color: theme.textPrimary }}
        >
          Выберите тариф
        </h1>
        <p
          className="text-center mb-8"
          style={{ color: theme.textSecondary }}
        >
          Текущий тариф:{' '}
          <span
            className={`inline-block px-3 py-0.5 rounded-full text-xs font-bold ${
              currentTier === 'pro'
                ? 'bg-purple-100 text-purple-700'
                : currentTier === 'ultra'
                ? 'bg-amber-100 text-amber-700'
                : 'bg-gray-100 text-gray-600'
            }`}
          >
            {currentTier.toUpperCase()}
          </span>
        </p>

        {error && (
          <div
            className="max-w-md mx-auto mb-6 p-3 rounded-xl text-sm text-center"
            style={{
              backgroundColor: 'rgba(220, 38, 38, 0.1)',
              color: '#dc2626',
            }}
          >
            {error}
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {plans.map((plan) => {
            const isCurrent = plan.key === currentTier
            const isLoading = loadingPlan === plan.key

            return (
              <div
                key={plan.key}
                className={`relative rounded-2xl p-6 bg-gradient-to-b ${plan.gradient} border-2 ${plan.borderColor} shadow-lg flex flex-col ${
                  plan.key === 'ultra' ? 'md:scale-105' : ''
                }`}
              >
                {/* Бейдж пробного периода */}
                {plan.trialDays !== null && (
                  <div
                    className={`absolute -top-3 left-1/2 -translate-x-1/2 px-4 py-1 rounded-full text-xs font-bold ${plan.chipBg} ${plan.chipText}`}
                  >
                    {plan.trialDays} дн. пробного периода
                  </div>
                )}

                <div className="text-center mb-4 mt-2">
                  <h2
                    className={`text-xl font-bold ${plan.chipText}`}
                  >
                    {plan.name}
                  </h2>
                </div>

                <div className="text-center mb-4">
                  <span className="text-3xl font-bold" style={{ color: theme.textPrimary }}>
                    {plan.price}
                  </span>
                  {plan.priceNote && (
                    <span className="text-sm ml-1" style={{ color: theme.textMuted }}>
                      {plan.priceNote}
                    </span>
                  )}
                </div>

                <ul className="space-y-2 mb-6 flex-1">
                  {plan.features.map((feat, i) => (
                    <li key={i} className="flex items-center gap-2 text-sm">
                      {feat.included ? (
                        <svg className="w-5 h-5 flex-shrink-0 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                        </svg>
                      ) : (
                        <svg className="w-5 h-5 flex-shrink-0 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                      )}
                      <span style={{ color: feat.included ? theme.textPrimary : theme.textMuted }}>
                        {feat.text}
                      </span>
                    </li>
                  ))}
                </ul>

                <button
                  onClick={() => handleUpgrade(plan.key)}
                  disabled={isCurrent || isLoading}
                  className={`w-full py-3 rounded-xl font-semibold text-sm transition-all ${
                    isCurrent
                      ? 'cursor-default opacity-60'
                      : 'hover:opacity-90 active:scale-95'
                  }`}
                  style={{
                    ...(plan.key === 'free'
                      ? {
                          backgroundColor: theme.border,
                          color: theme.textMuted,
                        }
                      : {
                          backgroundColor: theme.primary,
                          color: theme.textOnPrimary,
                        }),
                  }}
                >
                  {isLoading ? (
                    <span className="flex items-center justify-center gap-2">
                      <span className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                      Обработка...
                    </span>
                  ) : isCurrent ? (
                    'Текущий тариф'
                  ) : (
                    plan.cta
                  )}
                </button>
              </div>
            )
          })}
        </div>

        <p
          className="text-center mt-6 text-xs"
          style={{ color: theme.textMuted }}
        >
          Пробный период активируется один раз для каждого платного тарифа. Оплата спишется автоматически по окончании пробного периода.
        </p>
      </div>
    </div>
  )
}

export default SubscriptionPage