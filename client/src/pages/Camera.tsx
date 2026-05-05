import { useState, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { useScan } from '../store/scan'
import { mockProducts } from '../lib/mock-data'
import { X, Zap, ImageIcon, ScanLine } from 'lucide-react'
import type { CheckSession } from '../types'

export default function Camera() {
  const navigate = useNavigate()
  const selectedStore = useScan((s) => s.selectedStore)
  const setSession = useScan((s) => s.setSession)
  const [analyzing, setAnalyzing] = useState(false)
  const fileRef = useRef<HTMLInputElement>(null)

  const handleCapture = () => {
    setAnalyzing(true)
    // Simulate AI analysis
    setTimeout(() => {
      const foundCount = 5 + Math.floor(Math.random() * 4)
      const shuffled = [...mockProducts].sort(() => Math.random() - 0.5)
      const found = shuffled.slice(0, foundCount)
      const missing = shuffled.slice(foundCount)
      const total = found.length + missing.length

      const session: CheckSession = {
        id: Date.now().toString(),
        store: selectedStore || { id: '1', name: 'Negozio', city: '', address: '' },
        status: 'draft',
        scans: [],
        foundProducts: found,
        missingProducts: missing,
        coverage: Math.round((found.length / total) * 100),
        createdAt: new Date().toISOString(),
      }
      setSession(session)
      navigate('/scan/results', { replace: true })
    }, 2500)
  }

  if (analyzing) {
    return (
      <div className="min-h-svh bg-black flex flex-col items-center justify-center text-white">
        <div className="relative mb-8">
          <div className="w-24 h-24 border-4 border-white/20 rounded-3xl flex items-center justify-center">
            <ScanLine size={40} className="animate-pulse-scan" />
          </div>
          <div className="absolute inset-0 rounded-3xl border-2 border-accent animate-ping opacity-30" />
        </div>
        <h2 className="text-xl font-semibold mb-2">Analisi in corso</h2>
        <p className="text-white/50 text-sm">Riconoscimento prodotti tramite AI...</p>
        <div className="mt-8 flex gap-1">
          {[0, 1, 2].map((i) => (
            <div
              key={i}
              className="w-2 h-2 bg-white rounded-full animate-bounce"
              style={{ animationDelay: `${i * 0.15}s` }}
            />
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-svh bg-black flex flex-col relative">
      {/* Simulated camera viewfinder */}
      <div className="flex-1 relative flex items-center justify-center">
        {/* Close button */}
        <button
          onClick={() => navigate(-1)}
          className="absolute top-12 left-5 w-10 h-10 bg-black/40 backdrop-blur-md rounded-full flex items-center justify-center text-white z-10"
        >
          <X size={20} />
        </button>

        {/* Flash toggle */}
        <button className="absolute top-12 right-5 w-10 h-10 bg-black/40 backdrop-blur-md rounded-full flex items-center justify-center text-white z-10">
          <Zap size={18} />
        </button>

        {/* Viewfinder frame */}
        <div className="w-[85%] aspect-[3/4] relative">
          {/* Corner markers */}
          <div className="absolute top-0 left-0 w-8 h-8 border-t-2 border-l-2 border-white/80 rounded-tl-lg" />
          <div className="absolute top-0 right-0 w-8 h-8 border-t-2 border-r-2 border-white/80 rounded-tr-lg" />
          <div className="absolute bottom-0 left-0 w-8 h-8 border-b-2 border-l-2 border-white/80 rounded-bl-lg" />
          <div className="absolute bottom-0 right-0 w-8 h-8 border-b-2 border-r-2 border-white/80 rounded-br-lg" />

          {/* Scan line */}
          <div className="absolute inset-x-4 top-1/2 h-px bg-accent/50" />

          <div className="absolute inset-0 flex items-center justify-center">
            <p className="text-white/40 text-sm text-center px-6">
              Inquadra la vetrina e scatta una foto
            </p>
          </div>
        </div>

        {/* Store name badge */}
        {selectedStore && (
          <div className="absolute top-12 left-1/2 -translate-x-1/2 bg-black/40 backdrop-blur-md text-white text-xs font-medium px-3 py-1.5 rounded-full">
            {selectedStore.name}
          </div>
        )}
      </div>

      {/* Bottom controls */}
      <div className="pb-10 pt-6 px-6 safe-bottom flex items-center justify-center gap-8">
        {/* Gallery */}
        <button
          onClick={() => fileRef.current?.click()}
          className="w-12 h-12 bg-white/15 rounded-xl flex items-center justify-center text-white"
        >
          <ImageIcon size={22} />
        </button>
        <input ref={fileRef} type="file" accept="image/*" className="hidden" onChange={handleCapture} />

        {/* Shutter */}
        <button
          onClick={handleCapture}
          className="w-[72px] h-[72px] rounded-full border-4 border-white flex items-center justify-center active:scale-95 transition-transform"
        >
          <div className="w-[60px] h-[60px] rounded-full bg-white" />
        </button>

        {/* Placeholder for symmetry */}
        <div className="w-12 h-12" />
      </div>
    </div>
  )
}
