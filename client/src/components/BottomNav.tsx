import { useLocation, useNavigate } from 'react-router-dom'
import { Home, ScanLine, ClockArrowUp, Settings } from 'lucide-react'
import { motion } from 'framer-motion'

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
      <nav className="glass-pill rounded-full px-2 py-2 flex items-center pointer-events-auto relative">
        {tabs.map(({ path, icon: Icon, label }) => {
          const active = location.pathname.startsWith(path)
          return (
            <button
              key={path}
              onClick={() => navigate(path)}
              className={`relative flex flex-col items-center justify-center w-18 h-12 rounded-2xl transition-colors duration-300 z-10 ${
                active ? 'text-accent' : 'text-gray-400 active:text-gray-600'
              }`}
            >
              {active && (
                <motion.div
                  layoutId="activeTab"
                  className="absolute inset-0 bg-accent/10 rounded-2xl z-[-1]"
                  transition={{ type: 'spring', bounce: 0.2, duration: 0.6 }}
                />
              )}
              <motion.div
                animate={{ scale: active ? 1.1 : 1 }}
                transition={{ type: 'spring', stiffness: 400, damping: 17 }}
              >
                <Icon size={20} strokeWidth={active ? 2.5 : 2} />
              </motion.div>
              <span className={`text-[10px] font-bold mt-0.5 transition-opacity ${active ? 'opacity-100' : 'opacity-70'}`}>
                {label}
              </span>
            </button>
          )
        })}
      </nav>
    </div>
  )
}
