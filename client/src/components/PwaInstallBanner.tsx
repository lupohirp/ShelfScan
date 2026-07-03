import { useState, useEffect } from 'react'
import { usePwa } from '../store/pwa'
import { Share, Download, X, HelpCircle } from 'lucide-react'

export default function PwaInstallBanner() {
  const deferredPrompt = usePwa((s) => s.deferredPrompt)
  const showInstallBanner = usePwa((s) => s.showInstallBanner)
  const setShowInstallBanner = usePwa((s) => s.setShowInstallBanner)

  const [showIOSHint, setShowIOSHint] = useState(false)

  useEffect(() => {
    // Detect iOS
    const ios = /iPad|iPhone|iPod/.test(navigator.userAgent) && !(window as any).MSStream

    // Check if not standalone
    const isStandalone = window.matchMedia('(display-mode: standalone)').matches || (window.navigator as any).standalone
    
    if (ios && !isStandalone) {
      // Check if user dismissed iOS hint in this session
      const hintDismissed = sessionStorage.getItem('pwa-ios-hint-dismissed')
      if (!hintDismissed) {
        setShowIOSHint(true)
      }
    }
  }, [])

  const handleInstallClick = async () => {
    if (!deferredPrompt) return
    
    deferredPrompt.prompt()
    const { outcome } = await deferredPrompt.userChoice
    console.log(`User response to install prompt: ${outcome}`)
    
    // Clear prompt regardless of choice
    usePwa.getState().setDeferredPrompt(null)
    setShowInstallBanner(false)
  }

  const handleDismissIOS = () => {
    setShowIOSHint(false)
    sessionStorage.setItem('pwa-ios-hint-dismissed', 'true')
  }

  const handleDismissAndroid = () => {
    setShowInstallBanner(false)
    sessionStorage.setItem('pwa-android-banner-dismissed', 'true')
  }

  // iOS Instruction Card
  if (showIOSHint) {
    return (
      <div className="fixed bottom-6 left-6 right-6 md:left-auto md:right-6 md:w-[360px] z-50 bg-gradient-to-br from-black to-gray-900 text-white rounded-2xl p-5 shadow-2xl overflow-hidden border border-white/10 animate-in fade-in slide-in-from-bottom-4 duration-300">
        <div className="absolute top-0 right-0 p-3">
          <button 
            onClick={handleDismissIOS} 
            className="text-white/40 hover:text-white transition-colors p-1 rounded-full hover:bg-white/5"
          >
            <X size={16} />
          </button>
        </div>
        
        <div className="flex gap-4 items-start pr-6">
          <div className="w-10 h-10 rounded-xl bg-amber-500/10 border border-amber-500/30 flex items-center justify-center text-amber-400 shrink-0">
            <HelpCircle size={20} />
          </div>
          <div>
            <h3 className="text-sm font-bold uppercase tracking-wider text-amber-400 mb-1">
              Installa ShelfScan
            </h3>
            <p className="text-[11px] text-white/70 leading-relaxed mb-4">
              Installa l'applicazione sul tuo iPhone per l'avvio rapido e il supporto offline.
            </p>
            
            <div className="space-y-3 bg-white/5 rounded-xl p-3.5 border border-white/5 text-[11px]">
              <div className="flex items-center gap-2.5">
                <span className="w-5 h-5 rounded-full bg-white/10 flex items-center justify-center font-bold text-[9px]">1</span>
                <span>Premi il pulsante <strong>Condividi</strong> (<Share size={12} className="inline text-sky-400 mx-0.5" />) in basso nel browser Safari.</span>
              </div>
              <div className="flex items-center gap-2.5">
                <span className="w-5 h-5 rounded-full bg-white/10 flex items-center justify-center font-bold text-[9px]">2</span>
                <span>Seleziona l'opzione <strong>"Aggiungi alla schermata Home"</strong>.</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    )
  }

  // Android / Chrome Custom Install Banner
  if (showInstallBanner && deferredPrompt) {
    return (
      <div className="fixed bottom-6 left-6 right-6 md:left-auto md:right-6 md:w-[360px] z-50 bg-gradient-to-br from-black to-gray-900 text-white rounded-2xl p-5 shadow-2xl overflow-hidden border border-white/10 animate-in fade-in slide-in-from-bottom-4 duration-300">
        <div className="absolute top-0 right-0 p-3">
          <button 
            onClick={handleDismissAndroid} 
            className="text-white/40 hover:text-white transition-colors p-1 rounded-full hover:bg-white/5"
          >
            <X size={16} />
          </button>
        </div>

        <div className="flex gap-4 items-start pr-6">
          <div className="w-10 h-10 rounded-xl bg-amber-500/10 border border-amber-500/30 flex items-center justify-center text-amber-400 shrink-0">
            <Download size={20} />
          </div>
          <div className="flex-1">
            <h3 className="text-sm font-bold uppercase tracking-wider text-amber-400 mb-1">
              Installa ShelfScan
            </h3>
            <p className="text-[11px] text-white/70 leading-relaxed mb-4">
              Aggiungi l'app alla tua home screen per ricevere aggiornamenti, supporto offline e un avvio fulmineo.
            </p>
            
            <button
              onClick={handleInstallClick}
              className="px-5 h-9 bg-white text-black font-bold text-xs uppercase tracking-wider rounded-xl hover:bg-gray-100 active:scale-95 transition-all flex items-center justify-center gap-1.5 shadow-md shadow-white/5"
            >
              <Download size={14} />
              Installa Ora
            </button>
          </div>
        </div>
      </div>
    )
  }

  return null
}
