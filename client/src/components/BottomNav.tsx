import { useLocation, useNavigate } from 'react-router-dom'
import { Home, ScanLine, ClockArrowUp, Settings } from 'lucide-react'
import { motion } from 'framer-motion'

const tabs = [
  { path: '/home', icon: Home, label: 'Home' },
  { path: '/scan', icon: ScanLine, label: 'Check' },
  { path: '/history', icon: ClockArrowUp, label: 'Storico' },
  { path: '/settings', icon: Settings, label: 'Account' },
]

export default function BottomNav() {
  const location = useLocation()
  const navigate = useNavigate()

  return (
    <div className="fixed bottom-0 left-0 right-0 z-50 bg-white border-t border-gray-100 safe-bottom">
      <nav className="flex items-center justify-around h-16 max-w-lg mx-auto relative">
        {tabs.map(({ path, icon: Icon, label }) => {
          const active = location.pathname.startsWith(path)
          
          return (
            <button
              key={path}
              onClick={() => navigate(path)}
              className={`relative flex flex-col items-center justify-center w-full h-full transition-colors duration-300 ${
                active ? 'text-black' : 'text-gray-300'
              }`}
            >
              {active && (
                <motion.div
                  layoutId="activeTabUnderline"
                  className="absolute top-0 left-1/2 -translate-x-1/2 w-10 h-0.5 bg-black"
                  transition={{ type: 'spring', stiffness: 500, damping: 30 }}
                />
              )}
              <Icon size={20} strokeWidth={active ? 2.5 : 2} />
              <span className="text-[9px] font-black uppercase tracking-[0.15em] mt-1.5 leading-none">
                {label}
              </span>
            </button>
          )
        })}
      </nav>
    </div>
  )
}
