import { Outlet } from 'react-router-dom'
import BottomNav from './BottomNav'

export default function MobileLayout() {
  return (
    <div className="relative min-h-svh overflow-hidden bg-[#E8E4DF]">
      {/* Animated Background Elements - Softer/Warm (same as login) */}
      <div className="fixed top-[-10%] left-[-10%] w-[60%] h-[60%] bg-[#B4894D]/10 rounded-full blur-[120px] animate-pulse pointer-events-none" />
      <div className="fixed bottom-[-10%] right-[-10%] w-[60%] h-[60%] bg-[#B4894D]/5 rounded-full blur-[120px] animate-pulse delay-1000 pointer-events-none" />
      
      <div className="relative z-10">
        <Outlet />
      </div>
      <BottomNav />
    </div>
  )
}
