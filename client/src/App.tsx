import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuth } from './store/auth'
import Login from './pages/Login'
import Home from './pages/Home'
import StoreSelect from './pages/StoreSelect'
import Camera from './pages/Camera'
import ScanResults from './pages/ScanResults'
import ManualEdit from './pages/ManualEdit'
import ReportPreview from './pages/ReportPreview'
import History from './pages/History'
import HistoryDetail from './pages/HistoryDetail'
import Settings from './pages/Settings'
import AdminLayout from './pages/admin/AdminLayout'
import AdminDashboard from './pages/admin/AdminDashboard'
import AdminProducts from './pages/admin/AdminProducts'
import AdminStores from './pages/admin/AdminStores'
import AdminUsers from './pages/admin/AdminUsers'
import AdminChecks from './pages/admin/AdminChecks'

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuth((s) => s.isAuthenticated)
  if (!isAuthenticated) return <Navigate to="/login" replace />
  return <>{children}</>
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/" element={<Navigate to="/home" replace />} />

      {/* Mobile routes */}
      <Route path="/home" element={<ProtectedRoute><Home /></ProtectedRoute>} />
      <Route path="/scan" element={<ProtectedRoute><StoreSelect /></ProtectedRoute>} />
      <Route path="/scan/camera" element={<ProtectedRoute><Camera /></ProtectedRoute>} />
      <Route path="/scan/results" element={<ProtectedRoute><ScanResults /></ProtectedRoute>} />
      <Route path="/scan/edit" element={<ProtectedRoute><ManualEdit /></ProtectedRoute>} />
      <Route path="/scan/report" element={<ProtectedRoute><ReportPreview /></ProtectedRoute>} />
      <Route path="/history" element={<ProtectedRoute><History /></ProtectedRoute>} />
      <Route path="/history/:id" element={<ProtectedRoute><HistoryDetail /></ProtectedRoute>} />
      <Route path="/settings" element={<ProtectedRoute><Settings /></ProtectedRoute>} />

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
