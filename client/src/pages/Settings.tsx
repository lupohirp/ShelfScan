import { useNavigate } from 'react-router-dom'
import { useAuth } from '../store/auth'
import PageShell from '../components/PageShell'
import {
  User,
  Bell,
  Shield,
  HelpCircle,
  LogOut,
  ChevronRight,
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
    { icon: User, label: 'Profilo', subtitle: 'Account Information' },
    { icon: Bell, label: 'Notifiche', subtitle: 'Push & Email Settings' },
    { icon: Shield, label: 'Privacy', subtitle: 'Data & Security' },
    { icon: HelpCircle, label: 'Supporto', subtitle: 'Help Center & FAQ' },
  ]

  return (
    <PageShell>
      {/* Header */}
      <div className="px-8 pt-16 pb-10 safe-top border-b border-gray-100 bg-white">
        <h1 className="text-[32px] font-black tracking-tight text-black leading-none">ACCOUNT</h1>
        <p className="text-gray-400 text-[10px] font-black uppercase tracking-[0.2em] mt-2">Personal Settings</p>
      </div>

      <div className="px-8 pt-8">
        {/* User Card */}
        <div className="lj-card p-6 flex items-center gap-5 mb-10">
          <div className="w-16 h-16 bg-black flex items-center justify-center text-white text-xl font-black shrink-0">
            {user?.firstName?.[0]}{user?.lastName?.[0]}
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-[16px] font-black uppercase tracking-wider truncate">
              {user?.firstName} {user?.lastName}
            </p>
            <p className="text-[11px] font-bold text-gray-400 uppercase tracking-widest truncate">{user?.email}</p>
            <div className="mt-2">
              <span className="inline-block text-[9px] font-black text-white bg-black px-2 py-1 uppercase tracking-widest">
                {user?.role === 'rep' ? 'Associate' : user?.role}
              </span>
            </div>
          </div>
        </div>

        {/* Menu */}
        <div className="space-y-0">
          {menuItems.map(({ icon: Icon, label, subtitle }) => (
            <button
              key={label}
              className="w-full flex items-center justify-between py-6 border-b border-gray-50 active:opacity-50 transition-opacity text-left group"
            >
              <div className="flex items-center gap-4">
                <Icon size={18} className="text-black" strokeWidth={1.5} />
                <div>
                  <p className="text-[13px] font-black uppercase tracking-wider">{label}</p>
                  <p className="text-[10px] text-gray-400 font-bold uppercase tracking-tight">{subtitle}</p>
                </div>
              </div>
              <ChevronRight size={16} className="text-black opacity-30 group-hover:opacity-100 transition-opacity" />
            </button>
          ))}
        </div>

        {/* Logout */}
        <div className="mt-12">
          <button
            onClick={handleLogout}
            className="w-full h-14 border-2 border-black flex items-center justify-center gap-3 text-[12px] font-black uppercase tracking-[0.2em] active:bg-black active:text-white transition-all"
          >
            <LogOut size={18} strokeWidth={2} />
            Logout
          </button>
        </div>

        {/* Version */}
        <div className="text-center py-12 opacity-30">
          <p className="text-[9px] font-black uppercase tracking-[0.3em]">ShelfScan v1.0.0 (MVP)</p>
          <p className="text-[8px] font-bold uppercase tracking-[0.1em] mt-1">Official Store Terminal</p>
        </div>
      </div>
    </PageShell>
  )
}
