import { useNavigate } from 'react-router-dom'
import { useScan } from '../store/scan'
import TopBar from '../components/TopBar'
import PageShell from '../components/PageShell'
import {
  Share2,
  Download,
  Check,
  ScanLine,
  MapPin,
  Calendar,
  Package,
} from 'lucide-react'

export default function ReportPreview() {
  const navigate = useNavigate()
  const session = useScan((s) => s.currentSession)
  const clearSession = useScan((s) => s.clearSession)

  if (!session) {
    navigate('/home', { replace: true })
    return null
  }

  const { store, foundProducts, missingProducts, coverage } = session
  const date = new Date().toLocaleDateString('it-IT', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  })

  const handleFinalize = () => {
    clearSession()
    navigate('/home', { replace: true })
  }

  return (
    <PageShell bottomNav={false}>
      <TopBar title="Report" back />

      <div className="px-5 pt-4 pb-28">
        {/* Report card */}
        <div className="bg-white border border-gray-200 rounded-2xl overflow-hidden shadow-sm">
          {/* Header */}
          <div className="bg-black text-white p-5">
            <div className="flex items-center gap-2 mb-3">
              <ScanLine size={20} />
              <span className="text-sm font-semibold tracking-wide">SHELFSCAN REPORT</span>
            </div>
            <h2 className="text-xl font-bold">{store.name}</h2>
            <div className="flex items-center gap-3 mt-2 text-white/60 text-sm">
              <span className="flex items-center gap-1">
                <MapPin size={13} />
                {store.city}
              </span>
              <span className="flex items-center gap-1">
                <Calendar size={13} />
                {date}
              </span>
            </div>
          </div>

          {/* Coverage summary */}
          <div className="p-5 border-b border-gray-100">
            <div className="flex items-center justify-between mb-3">
              <span className="text-sm text-gray-500">Coverage complessivo</span>
              <span className="text-2xl font-bold">{coverage}%</span>
            </div>
            <div className="w-full h-2 bg-gray-100 rounded-full overflow-hidden">
              <div
                className="h-full rounded-full transition-all"
                style={{
                  width: `${coverage}%`,
                  backgroundColor: coverage >= 70 ? '#34C759' : coverage >= 40 ? '#FF9500' : '#FF3B30',
                }}
              />
            </div>
            <div className="flex items-center justify-between mt-2 text-xs text-gray-500">
              <span>{foundProducts.length} trovati</span>
              <span>{missingProducts.length} mancanti</span>
            </div>
          </div>

          {/* Found products */}
          <div className="p-5 border-b border-gray-100">
            <h3 className="text-sm font-semibold text-gray-900 mb-3 flex items-center gap-2">
              <div className="w-5 h-5 bg-success rounded-md flex items-center justify-center">
                <Check size={12} className="text-white" />
              </div>
              Prodotti trovati ({foundProducts.length})
            </h3>
            <div className="space-y-2">
              {foundProducts.map((p) => (
                <div key={p.id} className="flex items-center gap-2 text-sm">
                  <div className="w-1.5 h-1.5 bg-success rounded-full shrink-0" />
                  <span className="flex-1 truncate">{p.name}</span>
                  <span className="text-xs text-gray-400 font-mono">{p.sku}</span>
                </div>
              ))}
            </div>
          </div>

          {/* Missing products */}
          <div className="p-5">
            <h3 className="text-sm font-semibold text-gray-900 mb-3 flex items-center gap-2">
              <div className="w-5 h-5 bg-danger rounded-md flex items-center justify-center">
                <Package size={12} className="text-white" />
              </div>
              Prodotti mancanti ({missingProducts.length})
            </h3>
            <div className="space-y-2">
              {missingProducts.map((p) => (
                <div key={p.id} className="flex items-center gap-2 text-sm">
                  <div className="w-1.5 h-1.5 bg-danger rounded-full shrink-0" />
                  <span className="flex-1 truncate">{p.name}</span>
                  <span className="text-xs text-gray-400 font-mono">{p.sku}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Bottom actions */}
      <div className="fixed bottom-0 left-0 right-0 bg-white/80 backdrop-blur-xl border-t border-gray-200 p-4 safe-bottom">
        <div className="max-w-lg mx-auto space-y-2">
          <div className="flex gap-3">
            <button className="flex-1 h-12 bg-gray-100 rounded-xl flex items-center justify-center gap-2 text-[14px] font-semibold active:bg-gray-200 transition-colors">
              <Download size={18} />
              Scarica PDF
            </button>
            <button className="flex-1 h-12 bg-gray-100 rounded-xl flex items-center justify-center gap-2 text-[14px] font-semibold active:bg-gray-200 transition-colors">
              <Share2 size={18} />
              Condividi
            </button>
          </div>
          <button
            onClick={handleFinalize}
            className="w-full h-12 bg-black text-white rounded-xl font-semibold text-[15px] active:scale-[0.98] transition-transform"
          >
            Salva e chiudi
          </button>
        </div>
      </div>
    </PageShell>
  )
}
