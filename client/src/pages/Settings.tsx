import { useNavigate } from 'react-router-dom'
import { useAuth } from '../store/auth'
import PageShell from '../components/PageShell'
import BottomNav from '../components/BottomNav'
import {
  User,
  Bell,
  Shield,
  HelpCircle,
  LogOut,
  ChevronRight,
  ScanLine,
} from 'lucide-react'

export default function Settings() {
  const user = useAuth((s) => s.user)
  const logout = useAuth((s) => s.logout)
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
    navigate('/login', { replace: true })
  }

  const menuItems = [
    { icon: User, label: 'Profilo', subtitle: 'Modifica i tuoi dati' },
    { icon: Bell, label: 'Notifiche', subtitle: 'Gestisci le notifiche' },
    { icon: Shield, label: 'Privacy', subtitle: 'Impostazioni privacy' },
    { icon: HelpCircle, label: 'Supporto', subtitle: 'Assistenza e FAQ' },
  ]

  return (
    <PageShell>
      {/* Header */}
      <div className="px-5 pt-14 pb-6 safe-top">
        <h1 className="text-[28px] font-bold tracking-tight">Impostazioni</h1>
      </div>

      <div className="px-5">
        {/* User card */}
        <div className="bg-gray-50 rounded-2xl p-4 flex items-center gap-3 mb-6">
          <div className="w-14 h-14 bg-black rounded-2xl flex items-center justify-center text-white text-lg font-bold shrink-0">
            {user?.firstName?.[0]}{user?.lastName?.[0]}
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-[16px] font-semibold truncate">
              {user?.firstName} {user?.lastName}
            </p>
            <p className="text-sm text-gray-500 truncate">{user?.email}</p>
            <span className="inline-block text-[10px] font-medium text-accent bg-accent-light px-2 py-0.5 rounded-full mt-1 uppercase">
              {user?.role === 'rep' ? 'Rappresentante' : user?.role}
            </span>
          </div>
        </div>

        {/* Menu */}
        <div className="space-y-1">
          {menuItems.map(({ icon: Icon, label, subtitle }) => (
            <button
              key={label}
              className="w-full flex items-center gap-3 p-3.5 rounded-2xl active:bg-gray-50 transition-colors text-left"
            >
              <div className="w-10 h-10 bg-gray-100 rounded-xl flex items-center justify-center shrink-0">
                <Icon size={18} className="text-gray-600" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-[15px] font-medium">{label}</p>
                <p className="text-xs text-gray-500">{subtitle}</p>
              </div>
              <ChevronRight size={16} className="text-gray-300 shrink-0" />
            </button>
          ))}
        </div>

        {/* Logout */}
        <div className="mt-6 pt-6 border-t border-gray-100">
          <button
            onClick={handleLogout}
            className="w-full flex items-center gap-3 p-3.5 rounded-2xl active:bg-danger-light transition-colors text-left"
          >
            <div className="w-10 h-10 bg-danger-light rounded-xl flex items-center justify-center">
              <LogOut size={18} className="text-danger" />
            </div>
            <span className="text-[15px] font-medium text-danger">Esci</span>
          </button>
        </div>

        {/* Version */}
        <div className="text-center py-8">
          <div className="flex items-center justify-center gap-1.5 mb-1">
            <ScanLine size={14} className="text-gray-300" />
            <span className="text-xs text-gray-300 font-medium">ShelfScan</span>
          </div>
          <p className="text-[11px] text-gray-300">Versione 1.0.0 (MVP)</p>
        </div>
      </div>

      <BottomNav />
    </PageShell>
  )
}
