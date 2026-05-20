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

  const tabs: { key: Tab; label: string; count: number }[] = [
    { key: 'all', label: 'TUTTI', count: total },
    { key: 'found', label: 'TROVATI', count: foundProducts.length },
    { key: 'missing', label: 'MANCANTI', count: missingProducts.length },
  ]

  return (
    <PageShell>
      <TopBar
        title="RISULTATI ANALISI"
        back
        right={
          <button onClick={() => navigate('/scan/edit')} className="text-black p-1">
            <Pencil size={18} strokeWidth={1.5} />
          </button>
        }
      />

      {/* Coverage Section */}
      <div className="px-8 py-10 bg-gray-50/50 border-b border-gray-100 text-center">
        <p className="text-[10px] font-black uppercase tracking-[0.25em] text-gray-400 mb-4">Coverage Score</p>
        <div className="inline-flex items-baseline gap-1">
          <span className="text-[64px] font-black tracking-tighter leading-none text-black">{coverage}</span>
          <span className="text-[24px] font-black text-black">%</span>
        </div>
        <div className="mt-2">
          <p className="text-[12px] font-bold uppercase tracking-widest text-black">
            {foundProducts.length} DI {total} PRODOTTI IDENTIFICATI
          </p>
        </div>
      </div>

      {/* Tabs */}
      <div className="px-8 mt-10 mb-6">
        <div className="flex border-b border-gray-100">
          {tabs.map(({ key, label, count }) => (
            <button
              key={key}
              onClick={() => setTab(key)}
              className={`flex-1 pb-4 text-[11px] font-black uppercase tracking-[0.2em] transition-all relative ${
                tab === key
                  ? 'text-black'
                  : 'text-gray-300'
              }`}
            >
              {label}
              {tab === key && (
                <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-black" />
              )}
            </button>
          ))}
        </div>
      </div>

      {/* Product list */}
      <div className="px-8 space-y-4 pb-48">
        {displayed.map((product) => {
          const isFound = foundProducts.some((p) => p.id === product.id)
          return (
            <div
              key={product.id}
              className="flex items-center gap-5 py-4 border-b border-gray-50 last:border-0"
            >
              <div className="w-16 h-16 bg-gray-100 flex items-center justify-center shrink-0 grayscale overflow-hidden">
                {product.imageUrl ? (
                  <img src={product.imageUrl} alt={product.name} className="w-full h-full object-cover opacity-90" />
                ) : (
                  <Package size={24} className="text-gray-300" strokeWidth={1} />
                )}
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-[13px] font-black uppercase tracking-wider text-black truncate">{product.name}</p>
                <div className="flex items-center gap-3 mt-1">
                  <p className="text-[10px] text-gray-400 font-bold tracking-widest uppercase">
                    {isFound ? 'CONFIDENCE' : 'SKU'}: {product.sku}
                  </p>
                </div>
              </div>
              <div className="shrink-0">
                {isFound ? (
                  <Check size={18} className="text-black" strokeWidth={3} />
                ) : (
                  <div className="w-4 h-4 border-2 border-gray-200" />
                )}
              </div>
            </div>
          )
        })}
      </div>

      {/* Bottom actions */}
      <div className="fixed bottom-16 left-0 right-0 p-8 z-40 bg-white/95 backdrop-blur-md border-t border-gray-100">
        <div className="max-w-lg mx-auto flex gap-4">
          <button
            onClick={() => navigate('/scan/camera')}
            className="flex-1 border-2 border-black text-black h-14 text-[12px] font-black uppercase tracking-[0.2em] active:bg-gray-100 transition-colors"
          >
            Riprova
          </button>
          <button
            onClick={() => navigate('/scan/report')}
            className="flex-1 bg-black text-white h-14 text-[12px] font-black uppercase tracking-[0.2em] active:opacity-80 transition-opacity"
          >
            Invia Report
          </button>
        </div>
      </div>
    </PageShell>
  )
}
