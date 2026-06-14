import useAppStore from '../store/useAppStore'

const WelcomeModal = () => {
  const { welcomeModalOpen, setWelcomeModalOpen, user, theme } = useAppStore()

  // Show only if not logged in
  if (user || !welcomeModalOpen) return null

  const messages = [
    "Ты прекрасна!",
    "Будь собой!",
    "Все лучшее внутри!",
    "Ты сильная и независимая!",
    "Твоя улыбка освещает мир!",
    "Ты заслуживаешь только лучшего!",
    "Ты уникальна и неповторима!",
    "Твоя красота исходит изнутри!",
  ]

  const randomMessage = messages[Math.floor(Math.random() * messages.length)]

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm transition-all">
      <div 
        className="rounded-3xl p-12 max-w-md w-full shadow-2xl transform transition-all animate-fade-in backdrop-blur-xl"
        style={{ backgroundColor: theme.cardBg }}
      >
        <div className="text-center space-y-6">
          <div className="text-7xl mb-4 animate-lotus-bloom">🌸</div>
          <h2 className="text-3xl font-bold" style={{ color: theme.textPrimary }}>{randomMessage}</h2>
          <p className="text-lg" style={{ color: theme.textSecondary }}>Добро пожаловать в MentalChat</p>
          <button
            onClick={() => setWelcomeModalOpen(false)}
            className="px-8 py-3 rounded-full font-semibold shadow-lg hover:shadow-xl transition-all transform hover:scale-105"
            style={{ 
              background: `linear-gradient(135deg, ${theme.primary}, ${theme.accent})`, 
              color: theme.textOnPrimary 
            }}
          >
            Начать
          </button>
        </div>
      </div>
    </div>
  )
}

export default WelcomeModal
