import { useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import WelcomeModal from './components/WelcomeModal'
import RegistrationPage from './pages/RegistrationPage'
import LoginPage from './pages/LoginPage'
import DashboardPage from './pages/DashboardPage'
import SubscriptionPage from './pages/SubscriptionPage'
import useAppStore from './store/useAppStore'

const App = () => {
  const token = useAppStore((state) => state.token)
  const isAuthenticated = !!token
  const theme = useAppStore((state) => state.theme)

  // Apply theme globally via CSS custom properties on <html>
  useEffect(() => {
    const root = document.documentElement
    root.style.setProperty('--color-bg-from', theme.bgFrom)
    root.style.setProperty('--color-bg-to', theme.bgTo)
    root.style.setProperty('--color-primary', theme.primary)
    root.style.setProperty('--color-accent', theme.accent)
    root.style.setProperty('--color-card-bg', theme.cardBg)
    root.style.setProperty('--color-surface', theme.surface)
    root.style.setProperty('--color-text-primary', theme.textPrimary)
    root.style.setProperty('--color-text-secondary', theme.textSecondary)
    root.style.setProperty('--color-text-on-primary', theme.textOnPrimary)
    root.style.setProperty('--color-text-muted', theme.textMuted)
    root.style.setProperty('--color-danger-text', theme.dangerText)
    root.style.setProperty('--color-input-bg', theme.inputBg)
    root.style.setProperty('--color-input-border', theme.inputBorder)
    root.style.setProperty('--color-border', theme.border)
    root.style.setProperty('--color-message-bubble', theme.messageBubble)

    // Set body background
    document.body.style.background = `linear-gradient(135deg, ${theme.bgFrom}, ${theme.bgTo})`
    document.body.style.minHeight = '100vh'
  }, [theme])

  return (
    <BrowserRouter>
      <WelcomeModal />
      <Routes>
        <Route
          path="/"
          element={<Navigate to={isAuthenticated ? '/dashboard' : '/login'} />}
        />
        <Route
          path="/register"
          element={<RegistrationPage />}
        />
        <Route
          path="/login"
          element={<LoginPage />}
        />
        <Route
          path="/dashboard"
          element={isAuthenticated ? <DashboardPage /> : <Navigate to="/login" />}
        />
        <Route
          path="/subscription"
          element={isAuthenticated ? <SubscriptionPage /> : <Navigate to="/login" />}
        />
      </Routes>
    </BrowserRouter>
  )
}

export default App
