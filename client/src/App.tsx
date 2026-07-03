import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuth } from './store/auth'
import Login from './pages/Login'
import Home from './pages/Home'
import StoreSelect from './pages/StoreSelect'
import Camera from './pages/Camera'
import ScanResults from './pages/ScanResults'
import ManualEdit from './pages/ManualEdit'
import ReportPreview from './pages/ReportPreview'
import CustomizationForm from './pages/CustomizationForm'
import AdminLayout from './pages/admin/AdminLayout'
import AdminDashboard from './pages/admin/AdminDashboard'
import AdminProducts from './pages/admin/AdminProducts'
import AdminStores from './pages/admin/AdminStores'
import AdminUsers from './pages/admin/AdminUsers'
import AdminChecks from './pages/admin/AdminChecks'

import { useEffect } from 'react'
import { usePwa } from './store/pwa'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuth((s) => s.isAuthenticated)
  if (!isAuthenticated) return <Navigate to="/login" replace />
  return <>{children}</>
}

export default function App() {
  const setDeferredPrompt = usePwa((s) => s.setDeferredPrompt)
  const setShowInstallBanner = usePwa((s) => s.setShowInstallBanner)

  useEffect(() => {
    const handleBeforeInstallPrompt = (e: Event) => {
      e.preventDefault()
      setDeferredPrompt(e)
      const bannerDismissed = sessionStorage.getItem('pwa-android-banner-dismissed')
      if (!bannerDismissed) {
        setShowInstallBanner(true)
      }
    }

    window.addEventListener('beforeinstallprompt', handleBeforeInstallPrompt)

    // Check if already running standalone
    const isStandalone = window.matchMedia('(display-mode: standalone)').matches || (window.navigator as any).standalone
    if (isStandalone) {
      setShowInstallBanner(false)
    }

    return () => {
      window.removeEventListener('beforeinstallprompt', handleBeforeInstallPrompt)
    }
  }, [setDeferredPrompt, setShowInstallBanner])

  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/" element={<Navigate to="/home" replace />} />

      {/* Agent pages (No BottomNav or menus) */}
      <Route path="/home" element={<ProtectedRoute><Home /></ProtectedRoute>} />
      <Route path="/scan" element={<ProtectedRoute><StoreSelect /></ProtectedRoute>} />
      <Route path="/scan/camera" element={<ProtectedRoute><Camera /></ProtectedRoute>} />
      <Route path="/scan/results" element={<ProtectedRoute><ScanResults /></ProtectedRoute>} />
      <Route path="/scan/edit" element={<ProtectedRoute><ManualEdit /></ProtectedRoute>} />
      <Route path="/scan/report" element={<ProtectedRoute><ReportPreview /></ProtectedRoute>} />
      <Route path="/customization" element={<ProtectedRoute><CustomizationForm /></ProtectedRoute>} />

      {/* Admin routes */}
      <Route path="/admin" element={<ProtectedRoute><AdminLayout /></ProtectedRoute>}>
        <Route index element={<AdminDashboard />} />
        <Route path="products" element={<AdminProducts />} />
        <Route path="stores" element={<AdminStores />} />
        <Route path="users" element={<AdminUsers />} />
        <Route path="checks" element={<AdminChecks />} />
      </Route>

      <Route path="*" element={<Navigate to="/home" replace />} />
    </Routes>
  )
}
