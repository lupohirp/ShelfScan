import { Outlet } from 'react-router-dom'
import BottomNav from './BottomNav'

export default function MobileLayout() {
  return (
    <div className="relative min-h-svh bg-white">
      <main className="relative z-10">
        <Outlet />
      </main>
      <BottomNav />
    </div>
  )
}
