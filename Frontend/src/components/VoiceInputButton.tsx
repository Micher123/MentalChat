import React, { useState } from 'react'
import { useSpeechRecognition } from '../hooks/useSpeechRecognition'
import { userApi } from '../services/api'
import useAppStore from '../store/useAppStore'

interface VoiceInputButtonProps {
  onTranscript: (text: string) => void
  disabled?: boolean
  className?: string
}

const VoiceInputButton: React.FC<VoiceInputButtonProps> = ({
  onTranscript,
  disabled = false,
  className = '',
}) => {
  const [showPermissionModal, setShowPermissionModal] = useState(false)
  const [permissionGranted, setPermissionGranted] = useState(false)
  const [rememberChoice, setRememberChoice] = useState(true)
  const { setMicrophonePerm, theme } = useAppStore()

  const {
    isListening,
    isSupported,
    startListening,
    stopListening,
    requestMicrophonePermission,
  } = useSpeechRecognition({
    continuous: false,
    interimResults: true,
    language: 'ru-RU',
    onResult: (result) => {
      onTranscript(result.transcript)
      stopListening()
    },
    onError: (error) => {
      console.error('Voice recognition error:', error)
      stopListening()
    },
  })

  const handleButtonClick = async () => {
    if (!isSupported) {
      alert('Ваш браузер не поддерживает голосовой ввод')
      return
    }

    if (!permissionGranted) {
      setShowPermissionModal(true)
      return
    }

    if (isListening) {
      stopListening()
    } else {
      startListening()
    }
  }

  const handleGrantPermission = async () => {
    const granted = await requestMicrophonePermission()
    setPermissionGranted(granted)
    setShowPermissionModal(false)

    if (granted) {
      // Save to store
      setMicrophonePerm(true)
      
      // Save to database if remember choice is checked
      if (rememberChoice) {
        try {
          await userApi.saveMicrophonePermission(true)
        } catch (error) {
          console.error('Failed to save microphone permission:', error)
        }
      }
      
      startListening()
    }
  }

  const handleDenyPermission = () => {
    setShowPermissionModal(false)
  }

  return (
    <>
      <button
        onClick={handleButtonClick}
        disabled={disabled || !isSupported}
        className={`relative p-3 rounded-full transition-all transform hover:scale-105 ${className}`}
        style={
          disabled || !isSupported
            ? { backgroundColor: '#D1D5DB', color: '#9CA3AF' }
            : isListening
            ? { background: 'linear-gradient(135deg, #EF4444, #EC4899)', color: '#FFFFFF' }
            : { backgroundColor: theme.primary + '20', color: theme.primary }
        }
        title={isListening ? 'Остановить запись' : 'Голосовой ввод'}
      >
        {/* Microphone Icon */}
        <svg
          className="w-6 h-6"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z"
          />
        </svg>

        {/* Recording indicator */}
        {isListening && (
          <span className="absolute -top-1 -right-1 flex h-3 w-3">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-red-400 opacity-75"></span>
            <span className="relative inline-flex rounded-full h-3 w-3 bg-red-500"></span>
          </span>
        )}
      </button>

      {/* Permission Modal */}
      {showPermissionModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
          <div
            className="rounded-3xl p-8 max-w-md w-full mx-4 shadow-2xl transform transition-all"
            style={{ backgroundColor: theme.cardBg }}
          >
            <div className="text-center space-y-6">
              <div className="text-6xl">🎙️</div>
              
              <h3 className="text-2xl font-bold" style={{ color: theme.textPrimary }}>
                Доступ к микрофону
              </h3>
              
              <p style={{ color: theme.textSecondary }}>
                Разрешите доступ к микрофону для использования голосового ввода.
                Это позволит вам отправлять сообщения голосом.
              </p>

              <div className="rounded-xl p-4 text-left" style={{ backgroundColor: theme.surface }}>
                <p className="text-sm font-semibold mb-2" style={{ color: theme.textPrimary }}>
                  Что это даст:
                </p>
                <ul className="text-sm space-y-1" style={{ color: theme.textSecondary }}>
                  <li>✓ Быстрый ввод сообщений голосом</li>
                  <li>✓ Удобство использования</li>
                  <li>✓ Возможность диктовать длинные тексты</li>
                </ul>
              </div>

              <label className="flex items-center space-x-3 cursor-pointer">
                <input
                  type="checkbox"
                  checked={rememberChoice}
                  onChange={(e) => setRememberChoice(e.target.checked)}
                  className="w-5 h-5 rounded"
                  style={{ accentColor: theme.primary }}
                />
                <span className="text-sm" style={{ color: theme.textSecondary }}>
                  Запомнить выбор и не спрашивать снова
                </span>
              </label>

              <div className="flex space-x-4">
                <button
                  onClick={handleDenyPermission}
                  className="flex-1 py-3 rounded-xl font-semibold transition-all"
                  style={{ backgroundColor: theme.border, color: theme.textMuted }}
                >
                  Нет
                </button>
                <button
                  onClick={handleGrantPermission}
                  className="flex-1 py-3 rounded-xl font-semibold transition-all shadow-lg hover:shadow-xl"
                  style={{ background: `linear-gradient(135deg, ${theme.primary}, ${theme.accent})`, color: theme.textOnPrimary }}
                >
                  Да, разрешить
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Not supported message */}
      {!isSupported && (
        <div className="text-xs mt-2 text-center" style={{ color: theme.textMuted }}>
          Голосовой ввод не поддерживается вашим браузером
        </div>
      )}
    </>
  )
}

export default VoiceInputButton
