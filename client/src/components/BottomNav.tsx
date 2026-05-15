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
    <div className="fixed bottom-6 left-0 right-0 px-6 z-50 flex justify-center pointer-events-none">
      <nav className="glass-pill rounded-full px-4 py-2 flex items-center gap-1 pointer-events-auto">
        {tabs.map(({ path, icon: Icon, label }) => {
          const active = location.pathname.startsWith(path)
          return (
            <button
              key={path}
              onClick={() => navigate(path)}
              className={`flex flex-col items-center justify-center w-16 h-12 rounded-2xl transition-all duration-300 ${
                active 
                  ? 'text-accent bg-accent/10' 
                  : 'text-gray-400 active:bg-gray-100'
              }`}
            >
              <Icon size={20} strokeWidth={active ? 2.5 : 2} />
              <span className={`text-[10px] font-bold mt-0.5 ${active ? 'opacity-100' : 'opacity-70'}`}>
                {label}
              </span>
            </button>
          )
        })}
      </nav>
    </div>
  )
}
