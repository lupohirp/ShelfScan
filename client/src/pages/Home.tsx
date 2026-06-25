import { useNavigate } from 'react-router-dom'
import { useAuth } from '../store/auth'
import PageShell from '../components/PageShell'
import { ScanLine, Palette, LogOut } from 'lucide-react'

export default function Home() {
  const user = useAuth((s) => s.user)
  const logout = useAuth((s) => s.logout)
  const navigate = useNavigate()

  return (
    <PageShell bottomNav={false}>
      {/* Header Section */}
      <div className="px-8 pt-16 pb-8 safe-top flex items-center justify-between bg-white border-b border-gray-100">
        <div>
          <p className="text-gray-400 text-[10px] font-black uppercase tracking-[0.2em] mb-1">
            Area Agente / Rappresentante
          </p>
          <h1 className="text-[28px] font-black tracking-tight text-black leading-none uppercase">
            Ciao, {user?.firstName || 'Agente'}
          </h1>
        </div>
        <button
          onClick={logout}
          className="flex items-center gap-2 border border-gray-200 text-gray-500 hover:text-black hover:border-black px-4 py-2 text-[10px] font-black uppercase tracking-[0.2em] transition-all active:scale-95"
        >
          <LogOut size={12} />
          <span>Esci</span>
        </button>
      </div>

      {/* Main Container */}
      <div className="p-8 space-y-6 max-w-md mx-auto">
        <p className="text-gray-400 text-[11px] font-bold uppercase tracking-[0.15em] text-center mb-2">
          Seleziona l'operazione da eseguire
        </p>

        {/* CTA 1: Scansione */}
        <button
          onClick={() => navigate('/scan')}
          className="w-full bg-black text-white p-8 flex flex-col items-center justify-center gap-4 transition-all active:scale-[0.98] hover:opacity-90 border-2 border-black rounded-2xl shadow-lg shadow-black/5"
        >
          <div className="w-12 h-12 bg-white/10 flex items-center justify-center rounded-xl">
            <ScanLine size={28} className="text-white" strokeWidth={1.5} />
          </div>
          <div className="text-center">
            <span className="text-[16px] font-black uppercase tracking-[0.2em] block mb-2">
              Scansione Vetrina
            </span>
            <span className="text-white/50 text-[11px] font-medium block leading-relaxed max-w-[280px] mx-auto">
              Analizza la vetrina in tempo reale e invia il report di copertura.
            </span>
          </div>
        </button>

        {/* CTA 2: Personalizzazione */}
        <button
          onClick={() => navigate('/customization')}
          className="w-full bg-white text-black p-8 flex flex-col items-center justify-center gap-4 transition-all active:scale-[0.98] hover:bg-gray-50 border-2 border-gray-100 rounded-2xl shadow-sm"
        >
          <div className="w-12 h-12 bg-gray-50 flex items-center justify-center rounded-xl">
            <Palette size={28} className="text-black" strokeWidth={1.5} />
          </div>
          <div className="text-center">
            <span className="text-[16px] font-black uppercase tracking-[0.2em] block mb-2">
              Richiesta Allestimento
            </span>
            <span className="text-gray-400 text-[11px] font-medium block leading-relaxed max-w-[280px] mx-auto">
              Invia richieste di personalizzazione o nuovi materiali per lo store.
            </span>
          </div>
        </button>
      </div>
    </PageShell>
  )
}
