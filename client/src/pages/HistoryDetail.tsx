import { useParams, useNavigate } from 'react-router-dom'
import { mockHistory } from '../lib/mock-data'
import TopBar from '../components/TopBar'
import PageShell from '../components/PageShell'
import {
  Check,
  Share2,
  Download,
  MapPin,
  Calendar,
  Package,
} from 'lucide-react'

export default function HistoryDetail() {
  const { id } = useParams()
  const navigate = useNavigate()
  const check = mockHistory.find((c) => c.id === id)

  if (!check) {
    navigate('/history', { replace: true })
    return null
  }

  const { store, foundProducts, missingProducts, coverage } = check

  return (
    <PageShell bottomNav={false}>
      <TopBar
        title="Dettaglio Check"
        back
        right={
          <button className="text-accent p-1">
            <Share2 size={18} />
          </button>
        }
      />

      <div className="px-5 pt-4 pb-24">
        {/* Store info */}
        <div className="mb-5">
          <h2 className="text-xl font-bold">{store.name}</h2>
          <div className="flex items-center gap-3 mt-1.5 text-sm text-gray-500">
            <span className="flex items-center gap-1">
              <MapPin size={14} />
              {store.address}, {store.city}
            </span>
          </div>
          <div className="flex items-center gap-1 mt-1 text-sm text-gray-400">
            <Calendar size={14} />
            {new Date(check.createdAt).toLocaleDateString('it-IT', {
              day: 'numeric',
              month: 'long',
              year: 'numeric',
              hour: '2-digit',
              minute: '2-digit',
            })}
          </div>
        </div>

        {/* Coverage */}
        <div className="bg-gray-50 rounded-2xl p-5 mb-5">
          <div className="flex items-center justify-between mb-3">
            <span className="text-sm text-gray-500 font-medium">Copertura</span>
            <span className="text-2xl font-bold">{coverage}%</span>
          </div>
          <div className="w-full h-2.5 bg-gray-200 rounded-full overflow-hidden">
            <div
              className="h-full rounded-full"
              style={{
                width: `${coverage}%`,
                backgroundColor: coverage >= 70 ? '#34C759' : coverage >= 40 ? '#FF9500' : '#FF3B30',
              }}
            />
          </div>
          <div className="flex gap-4 mt-3">
            <div className="flex items-center gap-1.5">
              <div className="w-2 h-2 bg-success rounded-full" />
              <span className="text-xs text-gray-600">{foundProducts.length} trovati</span>
            </div>
            <div className="flex items-center gap-1.5">
              <div className="w-2 h-2 bg-danger rounded-full" />
              <span className="text-xs text-gray-600">{missingProducts.length} mancanti</span>
            </div>
          </div>
        </div>

        {/* Found */}
        <div className="mb-5">
          <h3 className="text-sm font-semibold mb-3 flex items-center gap-2">
            <div className="w-5 h-5 bg-success rounded-md flex items-center justify-center">
              <Check size={12} className="text-white" />
            </div>
            Trovati ({foundProducts.length})
          </h3>
          <div className="space-y-1.5">
            {foundProducts.map((p) => (
              <div key={p.id} className="flex items-center gap-3 p-3 bg-success-light/30 rounded-xl">
                <div className="w-9 h-9 bg-success-light rounded-lg flex items-center justify-center shrink-0">
                  <Check size={16} className="text-success" />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-[13px] font-semibold truncate">{p.name}</p>
                  <p className="text-[11px] text-gray-500">
                    SKU: {p.sku || 'N/A'}{p.confidence && ` · CONFIDENZA: ${p.confidence}%`}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Missing */}
        <div>
          <h3 className="text-sm font-semibold mb-3 flex items-center gap-2">
            <div className="w-5 h-5 bg-danger rounded-md flex items-center justify-center">
              <Package size={12} className="text-white" />
            </div>
            Mancanti ({missingProducts.length})
          </h3>
          <div className="space-y-1.5">
            {missingProducts.map((p) => (
              <div key={p.id} className="flex items-center gap-3 p-3 bg-danger-light/30 rounded-xl">
                <div className="w-9 h-9 bg-danger-light rounded-lg flex items-center justify-center shrink-0">
                  <Package size={16} className="text-danger" />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-[13px] font-semibold truncate">{p.name}</p>
                  <p className="text-[11px] text-gray-500">SKU: {p.sku || 'N/A'}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Bottom action */}
      <div className="fixed bottom-0 left-0 right-0 bg-white/80 backdrop-blur-xl border-t border-gray-200 p-4 safe-bottom">
        <div className="max-w-lg mx-auto">
          <button className="w-full h-12 bg-black text-white rounded-xl flex items-center justify-center gap-2 font-semibold text-[15px] active:scale-[0.98] transition-transform">
            <Download size={18} />
            Esporta PDF
          </button>
        </div>
      </div>
    </PageShell>
  )
}
