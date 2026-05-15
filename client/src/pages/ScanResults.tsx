import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useScan } from '../store/scan'
import TopBar from '../components/TopBar'
import PageShell from '../components/PageShell'
import {
  Check,
  FileText,
  Camera,
  Pencil,
  Package,
} from 'lucide-react'

type Tab = 'all' | 'found' | 'missing'

export default function ScanResults() {
  const [tab, setTab] = useState<Tab>('all')
  const navigate = useNavigate()
  const session = useScan((s) => s.currentSession)

  if (!session) {
    navigate('/home', { replace: true })
    return null
  }

  const { foundProducts, missingProducts, coverage } = session
  const total = foundProducts.length + missingProducts.length

  const displayed =
    tab === 'found'
      ? foundProducts
      : tab === 'missing'
      ? missingProducts
      : [...foundProducts, ...missingProducts]

  const coverageColor =
    coverage >= 70 ? 'text-[#30D158]' : coverage >= 40 ? 'text-[#FF9500]' : 'text-[#FF3B30]'

  const tabs: { key: Tab; label: string; count: number }[] = [
    { key: 'all', label: 'Tutti', count: total },
    { key: 'found', label: 'Trovati', count: foundProducts.length },
    { key: 'missing', label: 'Mancanti', count: missingProducts.length },
  ]

  return (
    <PageShell bottomNav={true}>
      <TopBar
        title="Risultati Analisi"
        back
        right={
          <button onClick={() => navigate('/scan/edit')} className="text-[#0071E3] p-1">
            <Pencil size={18} />
          </button>
        }
      />

      {/* Coverage header */}
      <div className="px-5 py-6 text-center">
        <div className="inline-flex items-center gap-4 glass-card rounded-[28px] px-8 py-5 border-white/60">
          <div className="relative w-16 h-16">
            <svg viewBox="0 0 36 36" className="w-full h-full -rotate-90">
              <path
                d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                fill="none"
                stroke="#F1F5F9"
                strokeWidth="3.5"
              />
              <path
                d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                fill="none"
                stroke={coverage >= 70 ? '#30D158' : coverage >= 40 ? '#FF9500' : '#FF3B30'}
                strokeWidth="3.5"
                strokeDasharray={`${coverage}, 100`}
                strokeLinecap="round"
              />
            </svg>
            <span className={`absolute inset-0 flex items-center justify-center text-lg font-bold ${coverageColor}`}>
              {coverage}%
            </span>
          </div>
          <div className="text-left">
            <p className="text-[24px] font-bold tracking-tight text-[#1D1D1F]">
              {foundProducts.length}/{total}
            </p>
            <p className="text-[12px] text-[#86868B] font-bold uppercase tracking-wider">Prodotti Trovati</p>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="px-5 mb-5">
        <div className="flex bg-black/5 rounded-2xl p-1">
          {tabs.map(({ key, label, count }) => (
            <button
              key={key}
              onClick={() => setTab(key)}
              className={`flex-1 py-2.5 rounded-xl text-[13px] font-bold transition-all ${
                tab === key
                  ? 'bg-white text-[#1D1D1F] shadow-sm'
                  : 'text-[#86868B]'
              }`}
            >
              {label} <span className="opacity-50 ml-0.5">{count}</span>
            </button>
          ))}
        </div>
      </div>

      {/* Product list */}
      <div className="px-5 space-y-3 pb-40">
        {displayed.map((product) => {
          const isFound = foundProducts.some((p) => p.id === product.id)
          return (
            <div
              key={product.id}
              className="flex items-center gap-4 p-4 glass-card rounded-[24px] border-white/60"
            >
              <div
                className={`w-12 h-12 rounded-2xl flex items-center justify-center shrink-0 overflow-hidden ${
                  isFound ? 'bg-[#E8FAF0]' : 'bg-black/5'
                }`}
              >
                {product.imageUrl ? (
                  <img src={product.imageUrl} alt={product.name} className="w-full h-full object-cover" />
                ) : isFound ? (
                  <Check size={20} className="text-[#30D158]" />
                ) : (
                  <Package size={20} className="text-[#86868B]" />
                )}
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-[15px] font-bold text-[#1D1D1F] truncate">{product.name}</p>
                <p className="text-[12px] text-[#86868B] font-medium mt-0.5 uppercase tracking-wide">
                  {product.sku} · {product.category}
                </p>
              </div>
              {isFound && (
                <span className="text-[10px] font-bold text-[#30D158] bg-[#E8FAF0] px-2.5 py-1 rounded-full border border-[#30D158]/10 uppercase tracking-wider">
                  Trovato
                </span>
              )}
              {!isFound && (
                <span className="text-[10px] font-bold text-[#FF3B30] bg-[#FFF0EF] px-2.5 py-1 rounded-full border border-[#FF3B30]/10 uppercase tracking-wider">
                  Mancante
                </span>
              )}
            </div>
          )
        })}
      </div>

      {/* Bottom actions */}
      <div className="fixed bottom-0 left-0 right-0 bg-[#E8E4DF]/90 backdrop-blur-xl border-t border-black/5 p-5 pb-8 safe-bottom z-40">
        <div className="max-w-lg mx-auto flex gap-4">
          <button
            onClick={() => navigate('/scan/camera')}
            className="flex-1 h-14 bg-white rounded-2xl flex items-center justify-center gap-2 text-[15px] font-bold text-[#1D1D1F] shadow-sm active:scale-[0.98] transition-all"
          >
            <Camera size={20} />
            Riprova
          </button>
          <button
            onClick={() => navigate('/scan/report')}
            className="flex-1 h-14 bg-[#1D1D1F] text-white rounded-2xl flex items-center justify-center gap-2 text-[15px] font-bold shadow-lg shadow-black/10 active:scale-[0.98] transition-all"
          >
            <FileText size={20} />
            Invia Report
          </button>
        </div>
      </div>
    </PageShell>
  )
}
