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
    <header className="sticky top-0 z-40 bg-[#E8E4DF]/80 backdrop-blur-2xl safe-top border-b border-black/5">
      <div className="flex items-center justify-between h-14 px-4">
        <div className="w-10">
          {back && (
            <button 
              onClick={() => navigate(-1)} 
              className="w-10 h-10 flex items-center justify-center -ml-2 text-[#1D1D1F] active:bg-black/5 rounded-full transition-colors"
            >
              <ArrowLeft size={22} />
            </button>
          )}
        </div>
        <h1 className="text-[17px] font-bold tracking-tight text-[#1D1D1F]">{title}</h1>
        <div className="w-10 flex justify-end">{right}</div>
      </div>
    </header>
  )
}
