import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useScan } from '../store/scan'
import TopBar from '../components/TopBar'
import PageShell from '../components/PageShell'
import {
  Check,
  Pencil,
  Package,
  Image as ImageIcon,
  X,
  ChevronLeft,
  ChevronRight
} from 'lucide-react'

type Tab = 'all' | 'found' | 'missing'

export default function ScanResults() {
  const [tab, setTab] = useState<Tab>('all')
  const [showPhoto, setShowPhoto] = useState(false)
  const [currentImageIdx, setCurrentImageIdx] = useState(0)
  const navigate = useNavigate()
  const session = useScan((s) => s.currentSession)

  if (!session) {
    navigate('/home', { replace: true })
    return null
  }

  const { foundProducts, missingProducts, coverage, analyzedImages } = session
  const total = foundProducts.length + missingProducts.length

  const displayed =
    tab === 'found'
      ? foundProducts
      : tab === 'missing'
      ? missingProducts
      : [...foundProducts, ...missingProducts]

  const tabs: { key: Tab; label: string }[] = [
    { key: 'all', label: 'TUTTI' },
    { key: 'found', label: 'TROVATI' },
    { key: 'missing', label: 'MANCANTI' },
  ]

  const currentAnalyzed = analyzedImages?.[currentImageIdx]

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
      <div className="px-8 py-10 bg-gray-50/50 border-b border-gray-100 text-center relative">
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
        {analyzedImages && analyzedImages.length > 0 && (
          <button 
            onClick={() => setShowPhoto(true)}
            className="absolute top-4 right-4 flex items-center justify-center gap-2 px-3 py-1.5 bg-black text-white rounded-full text-[10px] font-bold uppercase tracking-widest active:scale-95 transition-transform shadow-lg"
          >
            <ImageIcon size={14} /> Foto ({analyzedImages.length})
          </button>
        )}
      </div>

      {/* Tabs */}
      <div className="px-8 mt-6 mb-6">
        <div className="flex border-b border-gray-100">
          {tabs.map(({ key, label }) => (
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
          const isFound = foundProducts.some((p) => p.name === product.name)
          return (
            <div
              key={product.id || product.name}
              className="flex items-center gap-5 py-4 border-b border-gray-50 last:border-0"
            >
              <div className="w-16 h-16 bg-gray-100 flex items-center justify-center shrink-0 grayscale overflow-hidden rounded-lg">
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
                    {isFound ? 'CONFIDENCE' : 'SKU'}: {product.sku || 'MISSING'}
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

      {/* Photo Overlay Modal */}
      {showPhoto && analyzedImages && currentAnalyzed && (
        <div className="fixed inset-0 z-[100] bg-black flex flex-col">
          <div className="flex justify-between items-center p-6 pt-12">
            <div className="text-white text-[10px] font-black uppercase tracking-widest bg-white/10 px-3 py-1.5 rounded-full">
              Foto {currentImageIdx + 1} di {analyzedImages.length}
            </div>
            <button 
              onClick={() => setShowPhoto(false)}
              className="w-12 h-12 bg-white/20 backdrop-blur-md rounded-full flex items-center justify-center text-white active:bg-white/40"
            >
              <X size={24} />
            </button>
          </div>
          
          <div className="flex-1 relative overflow-hidden flex items-center justify-center p-4">
            {/* Navigation Arrows */}
            {analyzedImages.length > 1 && (
              <>
                <button 
                  onClick={() => setCurrentImageIdx(prev => (prev > 0 ? prev - 1 : analyzedImages.length - 1))}
                  className="absolute left-4 z-[110] w-10 h-10 bg-black/50 text-white rounded-full flex items-center justify-center active:scale-90"
                >
                  <ChevronLeft size={24} />
                </button>
                <button 
                  onClick={() => setCurrentImageIdx(prev => (prev < analyzedImages.length - 1 ? prev + 1 : 0))}
                  className="absolute right-4 z-[110] w-10 h-10 bg-black/50 text-white rounded-full flex items-center justify-center active:scale-90"
                >
                  <ChevronRight size={24} />
                </button>
              </>
            )}

            <div className="relative max-w-full max-h-full transition-all duration-300">
              <img 
                src={currentAnalyzed.capturedImage} 
                alt={`Analyzed ${currentImageIdx}`} 
                className="max-w-full max-h-full object-contain animate-in fade-in zoom-in-95 duration-300" 
              />
              {currentAnalyzed.detections?.map((det, i) => {
                const box = det.box || det.box_2d;
                if (!box || box.length !== 4) return null;
                const top = box[0] / 10;
                const left = box[1] / 10;
                const height = (box[2] - box[0]) / 10;
                const width = (box[3] - box[1]) / 10;
                
                return (
                  <div 
                    key={i} 
                    className="absolute border-2 border-green-400 bg-green-400/10 transition-all duration-500"
                    style={{
                      top: `${top}%`,
                      left: `${left}%`,
                      width: `${width}%`,
                      height: `${height}%`
                    }}
                  >
                    <div className="absolute -top-5 left-0 right-0 text-center">
                      <span className="bg-green-400 text-black text-[7px] font-black uppercase tracking-tighter px-1 py-0.5 whitespace-nowrap rounded-sm">
                        {det.desc}
                      </span>
                    </div>
                  </div>
                )
              })}
            </div>
          </div>

          {/* Thumbnails at bottom for quick navigation */}
          {analyzedImages.length > 1 && (
            <div className="h-24 bg-black/50 backdrop-blur-xl border-t border-white/10 flex items-center px-6 gap-3 overflow-x-auto no-scrollbar">
              {analyzedImages.map((ir, idx) => (
                <button 
                  key={idx}
                  onClick={() => setCurrentImageIdx(idx)}
                  className={`relative w-14 h-16 shrink-0 rounded-md overflow-hidden border-2 transition-all ${currentImageIdx === idx ? 'border-accent scale-110' : 'border-transparent opacity-50'}`}
                >
                  <img src={ir.capturedImage} className="w-full h-full object-cover" />
                </button>
              ))}
            </div>
          )}
        </div>
      )}
    </PageShell>
  )
}
