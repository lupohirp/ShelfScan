import { useLocation, useNavigate } from 'react-router-dom'
import { Home, ScanLine, ClockArrowUp, Settings } from 'lucide-react'

const tabs = [
  { path: '/home', icon: Home, label: 'Home' },
  { path: '/scan', icon: ScanLine, label: 'Scansiona' },
  { path: '/history', icon: ClockArrowUp, label: 'Storico' },
  { path: '/settings', icon: Settings, label: 'Impostazioni' },
]

export default function BottomNav() {
  const location = useLocation()
  const navigate = useNavigate()

  return (
    <nav className="fixed bottom-0 left-0 right-0 bg-white/80 backdrop-blur-xl border-t border-gray-200 safe-bottom z-50">
      <div className="max-w-lg mx-auto flex items-center justify-around h-14">
        {tabs.map(({ path, icon: Icon, label }) => {
          const active = location.pathname.startsWith(path)
          return (
            <button
              key={path}
              onClick={() => navigate(path)}
              className={`flex flex-col items-center gap-0.5 px-4 py-1 transition-colors ${
                active ? 'text-accent' : 'text-gray-400'
              }`}
            >
              <Icon size={22} strokeWidth={active ? 2.2 : 1.8} />
              <span className="text-[10px] font-medium">{label}</span>
            </button>
          )
        })}
      </div>
    </nav>
  )
}
