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
    <div className="fixed bottom-6 left-0 right-0 px-4 z-50 flex justify-center pointer-events-none">
      <nav className="glass-pill rounded-full px-1.5 py-1.5 flex items-center pointer-events-auto relative border-white/60 shadow-[0_8px_32px_rgba(62,42,26,0.12)]">
        {tabs.map(({ path, icon: Icon, label }) => {
          // Precise matching for the sliding pill
          const active = location.pathname.startsWith(path)
          
          return (
            <button
              key={path}
              onClick={() => navigate(path)}
              className={`relative flex flex-col items-center justify-center w-[85px] h-12 rounded-2xl transition-colors duration-300 z-10 shrink-0 ${
                active ? 'text-black' : 'text-[#86868B] active:text-[#1D1D1F]'
              }`}
            >
              {active && (
                <motion.div
                  layoutId="activeTabPill"
                  className="absolute inset-0 bg-[#B4894D] rounded-2xl z-[-1] shadow-[0_4px_12px_rgba(180,137,77,0.25)]"
                  transition={{ 
                    type: 'spring', 
                    stiffness: 380, 
                    damping: 30,
                    mass: 1
                  }}
                />
              )}
              <motion.div
                animate={{ scale: active ? 1.05 : 1 }}
                transition={{ type: 'spring', stiffness: 400, damping: 17 }}
              >
                <Icon size={19} strokeWidth={active ? 2.5 : 2} />
              </motion.div>
              <span className="text-[9px] font-bold mt-1 tracking-tight leading-none">
                {label}
              </span>
            </button>
          )
        })}
      </nav>
    </div>
  )
}
