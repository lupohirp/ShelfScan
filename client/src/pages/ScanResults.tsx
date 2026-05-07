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
    coverage >= 70 ? 'text-success' : coverage >= 40 ? 'text-warning' : 'text-danger'

  const tabs: { key: Tab; label: string; count: number }[] = [
    { key: 'all', label: 'Tutti', count: total },
    { key: 'found', label: 'Trovati', count: foundProducts.length },
    { key: 'missing', label: 'Mancanti', count: missingProducts.length },
  ]

  return (
    <PageShell bottomNav={false}>
      <TopBar
        title="Risultati"
        back
        right={
          <button onClick={() => navigate('/scan/edit')} className="text-accent p-1">
            <Pencil size={18} />
          </button>
        }
      />

      {/* Coverage header */}
      <div className="px-5 py-5 text-center">
        <div className="inline-flex items-center gap-3 bg-gray-50 rounded-2xl px-6 py-4">
          <div className="relative w-16 h-16">
            <svg viewBox="0 0 36 36" className="w-full h-full -rotate-90">
              <path
                d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                fill="none"
                stroke="#E8E8ED"
                strokeWidth="3"
              />
              <path
                d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                fill="none"
                stroke={coverage >= 70 ? '#34C759' : coverage >= 40 ? '#FF9500' : '#FF3B30'}
                strokeWidth="3"
                strokeDasharray={`${coverage}, 100`}
                strokeLinecap="round"
              />
            </svg>
            <span className={`absolute inset-0 flex items-center justify-center text-lg font-bold ${coverageColor}`}>
              {coverage}%
            </span>
          </div>
          <div className="text-left">
            <p className="text-[22px] font-bold tracking-tight">
              {foundProducts.length}/{total}
            </p>
            <p className="text-xs text-gray-500">prodotti trovati</p>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="px-5 mb-4">
        <div className="flex bg-gray-100 rounded-xl p-1">
          {tabs.map(({ key, label, count }) => (
            <button
              key={key}
              onClick={() => setTab(key)}
              className={`flex-1 py-2 rounded-lg text-[13px] font-medium transition-all ${
                tab === key
                  ? 'bg-white text-gray-900 shadow-sm'
                  : 'text-gray-500'
              }`}
            >
              {label} ({count})
            </button>
          ))}
        </div>
      </div>

      {/* Product list */}
      <div className="px-5 space-y-2 pb-32">
        {displayed.map((product) => {
          const isFound = foundProducts.some((p) => p.id === product.id)
          return (
            <div
              key={product.id}
              className="flex items-center gap-3 p-3 bg-gray-50 rounded-2xl"
            >
              <div
                className={`w-11 h-11 rounded-xl flex items-center justify-center shrink-0 overflow-hidden ${
                  isFound ? 'bg-success-light' : 'bg-gray-200'
                }`}
              >
                {product.imageUrl ? (
                  <img src={product.imageUrl} alt={product.name} className="w-full h-full object-cover" />
                ) : isFound ? (
                  <Check size={18} className="text-success" />
                ) : (
                  <Package size={18} className="text-gray-400" />
                )}
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-[14px] font-semibold truncate">{product.name}</p>
                <p className="text-xs text-gray-500">
                  {product.sku} · {product.category}
                </p>
              </div>
              {isFound && (
                <span className="text-[11px] font-medium text-success bg-success-light px-2 py-0.5 rounded-full shrink-0">
                  Trovato
                </span>
              )}
              {!isFound && (
                <span className="text-[11px] font-medium text-danger bg-danger-light px-2 py-0.5 rounded-full shrink-0">
                  Mancante
                </span>
              )}
            </div>
          )
        })}
      </div>

      {/* Bottom actions */}
      <div className="fixed bottom-0 left-0 right-0 bg-white/80 backdrop-blur-xl border-t border-gray-200 p-4 safe-bottom">
        <div className="max-w-lg mx-auto flex gap-3">
          <button
            onClick={() => navigate('/scan/camera')}
            className="flex-1 h-12 bg-gray-100 rounded-xl flex items-center justify-center gap-2 text-[14px] font-semibold active:bg-gray-200 transition-colors"
          >
            <Camera size={18} />
            Altra vetrina
          </button>
          <button
            onClick={() => navigate('/scan/report')}
            className="flex-1 h-12 bg-black text-white rounded-xl flex items-center justify-center gap-2 text-[14px] font-semibold active:scale-[0.98] transition-transform"
          >
            <FileText size={18} />
            Genera Report
          </button>
        </div>
      </div>
    </PageShell>
  )
}
