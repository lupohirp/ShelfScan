import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useAuth } from '../../store/auth'
import {
  LayoutDashboard,
  Package,
  MapPin,
  Users,
  ClipboardList,
  Settings,
  ScanLine,
  LogOut,
  Menu,
} from 'lucide-react'
import { useState } from 'react'

const navItems = [
  { to: '/admin', icon: LayoutDashboard, label: 'Dashboard', end: true },
  { to: '/admin/products', icon: Package, label: 'Prodotti' },
  { to: '/admin/stores', icon: MapPin, label: 'Negozi' },
  { to: '/admin/users', icon: Users, label: 'Utenti' },
  { to: '/admin/checks', icon: ClipboardList, label: 'Cronologia Check' },
  { to: '/admin/settings', icon: Settings, label: 'Impostazioni' },
]

export default function AdminLayout() {
  const logout = useAuth((s) => s.logout)
  const navigate = useNavigate()
  const [mobileOpen, setMobileOpen] = useState(false)

  const handleLogout = () => {
    logout()
    navigate('/login', { replace: true })
  }

  const sidebar = (
    <div className="flex flex-col h-full">
      {/* Logo */}
      <div className="flex items-center gap-2.5 px-5 h-16 border-b border-gray-100">
        <div className="w-8 h-8 bg-black rounded-lg flex items-center justify-center">
          <ScanLine size={16} className="text-white" />
        </div>
        <span className="text-[15px] font-bold tracking-tight">ShelfScan</span>
        <span className="text-[10px] font-medium text-gray-400 bg-gray-100 px-1.5 py-0.5 rounded ml-1">Admin</span>
      </div>

      {/* Nav */}
      <nav className="flex-1 py-3 px-3 space-y-0.5 overflow-y-auto">
        {navItems.map(({ to, icon: Icon, label, end }) => (
          <NavLink
            key={to}
            to={to}
            end={end}
            onClick={() => setMobileOpen(false)}
            className={({ isActive }) =>
              `flex items-center gap-2.5 px-3 py-2.5 rounded-xl text-[14px] font-medium transition-colors ${
                isActive
                  ? 'bg-black text-white'
                  : 'text-gray-600 hover:bg-gray-50'
              }`
            }
          >
            <Icon size={18} />
            {label}
          </NavLink>
        ))}
      </nav>

      {/* Logout */}
      <div className="p-3 border-t border-gray-100">
        <button
          onClick={handleLogout}
          className="flex items-center gap-2.5 px-3 py-2.5 rounded-xl text-[14px] font-medium text-gray-500 hover:text-danger hover:bg-danger-light transition-colors w-full"
        >
          <LogOut size={18} />
          Esci
        </button>
      </div>
    </div>
  )

  return (
    <div className="min-h-svh bg-gray-50 flex">
      {/* Desktop sidebar */}
      <aside className="hidden lg:flex w-60 bg-white border-r border-gray-200 flex-col shrink-0 fixed inset-y-0 left-0 z-30">
        {sidebar}
      </aside>

      {/* Mobile sidebar */}
      {mobileOpen && (
        <div className="fixed inset-0 z-50 lg:hidden">
          <div className="absolute inset-0 bg-black/30" onClick={() => setMobileOpen(false)} />
          <aside className="absolute left-0 top-0 bottom-0 w-64 bg-white shadow-xl">
            {sidebar}
          </aside>
        </div>
      )}

      {/* Main content */}
      <div className="flex-1 lg:ml-60">
        {/* Mobile header */}
        <header className="lg:hidden sticky top-0 z-20 bg-white/80 backdrop-blur-xl border-b border-gray-200 h-12 flex items-center px-4 gap-3">
          <button onClick={() => setMobileOpen(true)} className="text-gray-600">
            <Menu size={22} />
          </button>
          <span className="text-[15px] font-semibold">ShelfScan Admin</span>
        </header>

        <main className="p-4 lg:p-8">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
