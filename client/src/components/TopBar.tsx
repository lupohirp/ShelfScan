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
    <header className="sticky top-0 z-40 bg-white/95 backdrop-blur-md safe-top border-b border-gray-100">
      <div className="flex items-center justify-between h-14 px-6">
        <div className="w-10">
          {back && (
            <button 
              onClick={() => navigate(-1)} 
              className="w-10 h-10 flex items-center justify-center -ml-3 text-black active:opacity-50 transition-opacity"
            >
              <ArrowLeft size={22} strokeWidth={1.5} />
            </button>
          )}
        </div>
        <h1 className="text-[13px] font-black uppercase tracking-[0.25em] text-black text-center flex-1 truncate px-2">
          {title}
        </h1>
        <div className="w-10 flex justify-end">
          {right}
        </div>
      </div>
    </header>
  )
}
