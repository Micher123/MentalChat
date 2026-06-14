import { useState, useCallback, useRef } from 'react'

export interface SpeechRecognitionResult {
  transcript: string
  confidence: number
}

export interface UseSpeechRecognitionOptions {
  onResult?: (result: SpeechRecognitionResult) => void
  onError?: (error: Error) => void
  onStart?: () => void
  onEnd?: () => void
  continuous?: boolean
  interimResults?: boolean
  language?: string
}

export function useSpeechRecognition(options: UseSpeechRecognitionOptions = {}) {
  const {
    onResult,
    onError,
    onStart,
    onEnd,
    continuous = false,
    interimResults = false,
    language = 'ru-RU',
  } = options

  const [isListening, setIsListening] = useState(false)
  const [isSupported, setIsSupported] = useState(true)
  const [transcript, setTranscript] = useState('')
  const recognitionRef = useRef<SpeechRecognition | null>(null)

  // Check if speech recognition is supported
  const checkSupport = useCallback(() => {
    const SpeechRecognition =
      (window as any).SpeechRecognition || (window as any).webkitSpeechRecognition

    if (!SpeechRecognition) {
      setIsSupported(false)
      return false
    }

    return true
  }, [])

  // Request microphone permission
  const requestMicrophonePermission = useCallback(async (): Promise<boolean> => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      stream.getTracks().forEach(track => track.stop())
      return true
    } catch (error) {
      console.error('Microphone permission denied:', error)
      return false
    }
  }, [])

  // Start listening
  const startListening = useCallback(async () => {
    if (!checkSupport()) {
      onError?.(new Error('Speech recognition is not supported in this browser'))
      return
    }

    const hasPermission = await requestMicrophonePermission()
    if (!hasPermission) {
      onError?.(new Error('Microphone permission denied'))
      return
    }

    const SpeechRecognition =
      (window as any).SpeechRecognition || (window as any).webkitSpeechRecognition

    const recognition = new SpeechRecognition()
    recognitionRef.current = recognition

    recognition.continuous = continuous
    recognition.interimResults = interimResults
    recognition.lang = language

    recognition.onstart = () => {
      setIsListening(true)
      onStart?.()
    }

    recognition.onresult = (event: SpeechRecognitionEvent) => {
      let finalTranscript = ''
      let interimTranscript = ''

      for (let i = event.resultIndex; i < event.results.length; i++) {
        const result = event.results[i]
        if (result.isFinal) {
          finalTranscript += result[0].transcript
        } else {
          interimTranscript += result[0].transcript
        }
      }

      const fullTranscript = finalTranscript || interimTranscript
      setTranscript(fullTranscript)

      if (finalTranscript) {
        onResult?.({
          transcript: finalTranscript,
          confidence: event.results[0][0].confidence,
        })
      }
    }

    recognition.onerror = (event: SpeechRecognitionErrorEvent) => {
      console.error('Speech recognition error:', event.error)
      setIsListening(false)

      let errorMessage = 'Unknown error occurred'
      switch (event.error) {
        case 'no-speech':
          errorMessage = 'No speech detected. Please try again.'
          break
        case 'audio-capture':
          errorMessage = 'No microphone found. Please check your microphone.'
          break
        case 'not-allowed':
          errorMessage = 'Microphone permission denied.'
          break
        case 'network':
          errorMessage = 'Network error occurred.'
          break
        default:
          errorMessage = event.error
      }

      onError?.(new Error(errorMessage))
    }

    recognition.onend = () => {
      setIsListening(false)
      onEnd?.()
    }

    recognition.start()
  }, [checkSupport, requestMicrophonePermission, continuous, interimResults, language, onResult, onError, onStart, onEnd])

  // Stop listening
  const stopListening = useCallback(() => {
    if (recognitionRef.current) {
      recognitionRef.current.stop()
      recognitionRef.current = null
    }
    setIsListening(false)
  }, [])

  // Clear transcript
  const clearTranscript = useCallback(() => {
    setTranscript('')
  }, [])

  return {
    isListening,
    isSupported,
    transcript,
    startListening,
    stopListening,
    clearTranscript,
    requestMicrophonePermission,
  }
}

export default useSpeechRecognition
