import { ArrowLeft } from 'lucide-react'
import { useNavigate } from 'react-router-dom'

interface TopBarProps {
  title: string
  back?: boolean
  right?: React.ReactNode
}

export default function TopBar({ title, back, right }: TopBarProps) {
  const navigate = useNavigate()

  return (
    <header className="sticky top-0 z-40 bg-white/80 backdrop-blur-xl safe-top">
      <div className="flex items-center justify-between h-12 px-4">
        <div className="w-10">
          {back && (
            <button onClick={() => navigate(-1)} className="p-1 -ml-1 text-accent">
              <ArrowLeft size={22} />
            </button>
          )}
        </div>
        <h1 className="text-[17px] font-semibold tracking-tight">{title}</h1>
        <div className="w-10 flex justify-end">{right}</div>
      </div>
      <div className="h-px bg-gray-200" />
    </header>
  )
}
