import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useScan } from '../store/scan'
import TopBar from '../components/TopBar'
import PageShell from '../components/PageShell'
import { getApiUrl } from '../lib/api'
import {
  Check,
  ScanLine,
  MapPin,
  Calendar,
  Package,
  Loader2,
} from 'lucide-react'

export default function ReportPreview() {
  const navigate = useNavigate()
  const session = useScan((s) => s.currentSession)
  const clearSession = useScan((s) => s.clearSession)
  const [saving, setSaving] = useState(false)

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

  const handleFinalize = async () => {
    setSaving(true)
    try {
      const apiBase = getApiUrl()
      const finalizedSession = {
        ...session,
        status: 'finalized' as const,
        finalizedAt: new Date().toISOString(),
      }
      
      const res = await fetch(`${apiBase}/visits`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(finalizedSession),
      })
      
      if (!res.ok) {
        throw new Error(await res.text())
      }
      
      clearSession()
      navigate('/home', { replace: true })
    } catch (err) {
      console.error('Error saving visit:', err)
      alert('Impossibile salvare la visita: ' + (err instanceof Error ? err.message : String(err)))
    } finally {
      setSaving(false)
    }
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
                  <span className="text-xs text-gray-400 font-mono">
                    SKU: {p.sku || 'N/A'}{p.confidence && ` (${p.confidence}%)`}
                  </span>
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
                  <span className="text-xs text-gray-400 font-mono">SKU: {p.sku || 'N/A'}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Bottom actions */}
      <div className="fixed bottom-0 left-0 right-0 bg-white/95 backdrop-blur-xl border-t border-gray-200 p-4 safe-bottom">
        <div className="max-w-lg mx-auto">
          <button
            onClick={handleFinalize}
            disabled={saving}
            className="w-full h-12 bg-black text-white rounded-xl font-semibold text-[15px] active:scale-[0.98] transition-transform flex items-center justify-center gap-2 disabled:opacity-50 disabled:pointer-events-none"
          >
            {saving ? (
              <>
                <Loader2 className="animate-spin" size={18} />
                <span>Salvataggio...</span>
              </>
            ) : (
              <span>Salva e chiudi</span>
            )}
          </button>
        </div>
      </div>
    </PageShell>
  )
}
