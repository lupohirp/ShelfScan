import { useState, useRef, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useScan } from '../store/scan'
import { X, Zap, ImageIcon, ScanLine, Camera as CameraIcon } from 'lucide-react'
import type { CheckSession } from '../types'
import { GoogleGenerativeAI } from '@google/generative-ai'

export default function Camera() {
  const navigate = useNavigate()
  const selectedStore = useScan((s) => s.selectedStore)
  const setSession = useScan((s) => s.setSession)
  
  const [analyzing, setAnalyzing] = useState(false)
  const [pitch, setPitch] = useState(0)
  const [roll, setRoll] = useState(0)
  const [isStable, setIsStable] = useState(false)
  const [stream, setStream] = useState<MediaStream | null>(null)
  
  const videoRef = useRef<HTMLVideoElement>(null)
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const fileRef = useRef<HTMLInputElement>(null)
  const stabilityTimerRef = useRef<NodeJS.Timeout | null>(null)

  // Request 4K Camera & Lifecycle Management
  useEffect(() => {
    let currentStream: MediaStream | null = null
    async function setupCamera() {
      try {
        currentStream = await navigator.mediaDevices.getUserMedia({
          video: {
            facingMode: 'environment',
            width: { ideal: 3840 },
            height: { ideal: 2160 },
          },
          audio: false,
        })
        setStream(currentStream)
        if (videoRef.current) {
          videoRef.current.srcObject = currentStream
        }
      } catch (err) {
        console.error('Error accessing camera:', err)
      }
    }
    setupCamera()
    
    return () => {
      if (currentStream) {
        currentStream.getTracks().forEach(track => track.stop())
        console.log('Camera stopped')
      }
    }
  }, [])

  const handleBack = () => {
    if (stream) {
      stream.getTracks().forEach(track => track.stop())
    }
    navigate(-1)
  }

  // Device Orientation for Leveling
  useEffect(() => {
    const handleOrientation = (e: DeviceOrientationEvent) => {
      if (e.beta !== null && e.gamma !== null) {
        setPitch(e.beta)
        setRoll(e.gamma)
      }
    }
    window.addEventListener('deviceorientation', handleOrientation)
    return () => window.removeEventListener('deviceorientation', handleOrientation)
  }, [])

  const captureImage = useCallback(async () => {
    console.log('Capture triggered')
    if (!videoRef.current || !canvasRef.current) {
      console.error('Video or Canvas ref missing')
      return
    }
    
    const video = videoRef.current
    const canvas = canvasRef.current
    
    // Use actual video dimensions for 4K capture
    canvas.width = video.videoWidth
    canvas.height = video.videoHeight
    const ctx = canvas.getContext('2d')
    if (!ctx) return
    
    ctx.drawImage(video, 0, 0)
    const imageData = canvas.toDataURL('image/jpeg', 0.8)
    
    // Visual flash effect
    const flash = document.createElement('div')
    flash.className = 'fixed inset-0 bg-white z-[200] animate-out fade-out duration-300'
    document.body.appendChild(flash)
    setTimeout(() => flash.remove(), 300)

    handleAnalysis(imageData)
  }, [videoRef, canvasRef])

  // Stability Detection & Auto-capture
  useEffect(() => {
    // For shelf scanning, "level" means the phone is vertical (beta approx 90)
    // and not tilted sideways (gamma approx 0)
    const isLevel = Math.abs(pitch - 90) < 7 && Math.abs(roll) < 7
    setIsStable(isLevel)

    if (isLevel && !analyzing) {
      if (!stabilityTimerRef.current) {
        stabilityTimerRef.current = setTimeout(() => {
          captureImage()
        }, 1500)
      }
    } else {
      if (stabilityTimerRef.current) {
        clearTimeout(stabilityTimerRef.current)
        stabilityTimerRef.current = null
      }
    }
  }, [pitch, roll, captureImage, analyzing])

  const handleAnalysis = async (imageData: string) => {
    if (analyzing) return
    setAnalyzing(true)
    
    try {
      // Do not stop the stream here, keep it alive for better UX and retries
      
      const res = await fetch(imageData)
      const blob = await res.blob()
      
      if (blob.size === 0) {
        throw new Error('Captured image is empty')
      }
      
      const formData = new FormData()
      formData.append('image', blob, 'shelf_scan.jpg')
      
      const apiHost = window.location.hostname
      const analysisResponse = await fetch(`http://${apiHost}:8080/analyze`, {
        method: 'POST',
        body: formData
      })
      
      if (!analysisResponse.ok) {
        const errText = await analysisResponse.text()
        throw new Error(`Analyze API failed: ${errText}`)
      }
      
      const analysisData = await analysisResponse.json() as {
        found: {desc: string, match: string, score: number}[],
        missing: string[]
      }
      console.log('AI Analysis data:', analysisData)
      
      const identifiedItems = analysisData.found.filter(r => r.score > 0.45).map(r => r.match)
      const uniqueIdentified = [...new Set(identifiedItems)]

      const foundProducts: any[] = uniqueIdentified.map(name => ({
        id: 'found-' + name,
        sku: 'IDENTIFIED',
        name: name,
        category: 'identified',
        imageUrl: '',
        status: 'active'
      }))

      const missingProducts: any[] = analysisData.missing.map(name => ({
        id: 'missing-' + name,
        sku: 'MISSING',
        name: name,
        category: 'missing',
        imageUrl: '',
        status: 'active'
      }))

      const session: CheckSession = {
        id: Date.now().toString(),
        store: selectedStore || { id: '1', name: 'Vetrina Centrale', city: '', address: 'Indirizzo mock' },
        status: 'draft',
        scans: [],
        foundProducts: foundProducts,
        missingProducts: missingProducts,
        coverage: (foundProducts.length + missingProducts.length) > 0 
          ? Math.round((foundProducts.length / (foundProducts.length + missingProducts.length)) * 100) 
          : 0,
        createdAt: new Date().toISOString(),
      }
      
      setSession(session)
      navigate('/scan/results', { replace: true })
    } catch (err) {
      console.error('Shelf Analysis failed:', err)
      const errorMsg = err instanceof Error ? err.message : String(err)
      alert('Errore durante l\'analisi della vetrina: ' + errorMsg)
      setAnalyzing(false)
    }
  }

  const handleFileCapture = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      const reader = new FileReader()
      reader.onload = (ev) => {
        if (ev.target?.result) {
          handleAnalysis(ev.target.result as string)
        }
      }
      reader.readAsDataURL(file)
    }
  }

  if (analyzing) {
    return (
      <div className="fixed inset-0 bg-black flex flex-col items-center justify-center text-white z-[100]">
        <div className="relative mb-8">
          <div className="w-24 h-24 border-4 border-white/20 rounded-3xl flex items-center justify-center">
            <ScanLine size={40} className="animate-pulse-scan text-accent" />
          </div>
          <div className="absolute inset-0 rounded-3xl border-2 border-accent animate-ping opacity-30" />
        </div>
        <h2 className="text-xl font-semibold mb-2">Analisi in corso</h2>
        <p className="text-white/50 text-sm">Confronto con inventario Qdrant...</p>
      </div>
    )
  }

  return (
    <div className="fixed inset-0 bg-black overflow-hidden select-none">
      {/* 1. Video Layer */}
      <video
        ref={videoRef}
        autoPlay
        playsInline
        muted
        className="absolute inset-0 w-full h-full object-cover z-0"
      />
      
      <canvas ref={canvasRef} className="hidden" />

      {/* 2. Top Bar UI */}
      <div className="absolute top-0 left-0 right-0 p-6 pt-12 flex justify-between items-center z-50 pointer-events-none">
        <button
          onClick={handleBack}
          className="w-12 h-12 bg-black/50 backdrop-blur-xl rounded-full flex items-center justify-center text-white pointer-events-auto active:bg-white/20 transition-colors"
        >
          <X size={24} />
        </button>

        {selectedStore && (
          <div className="bg-black/50 backdrop-blur-md text-white text-xs font-bold tracking-wider uppercase px-4 py-2 rounded-full border border-white/10">
            {selectedStore.name}
          </div>
        )}

        <button className="w-12 h-12 bg-black/50 backdrop-blur-xl rounded-full flex items-center justify-center text-white pointer-events-auto">
          <Zap size={20} />
        </button>
      </div>

      {/* 3. Viewfinder / Level UI */}
      <div className="absolute inset-0 flex items-center justify-center z-10 pointer-events-none">
        {/* Level Bubble */}
        <div className={`w-36 h-36 border-2 rounded-full flex items-center justify-center transition-all duration-300 ${isStable ? 'border-green-500 bg-green-500/20 scale-110' : 'border-white/20'}`}>
          <div 
            className={`w-5 h-5 rounded-full transition-colors duration-300 ${isStable ? 'bg-green-500 shadow-[0_0_15px_rgba(34,197,94,0.6)]' : 'bg-red-500'}`}
            style={{ transform: `translate(${roll * 2.5}px, ${(pitch - 90) * 2.5}px)` }}
          />
        </div>

        {/* Frame */}
        <div className={`absolute w-[85%] aspect-[3/4] border-2 rounded-3xl transition-colors duration-500 ${isStable ? 'border-green-500/60' : 'border-white/10'}`}>
          <div className={`absolute top-0 left-0 w-12 h-12 border-t-4 border-l-4 rounded-tl-3xl transition-colors ${isStable ? 'border-green-500' : 'border-white/40'}`} />
          <div className={`absolute top-0 right-0 w-12 h-12 border-t-4 border-r-4 rounded-tr-3xl transition-colors ${isStable ? 'border-green-500' : 'border-white/40'}`} />
          <div className={`absolute bottom-0 left-0 w-12 h-12 border-b-4 border-l-4 rounded-bl-3xl transition-colors ${isStable ? 'border-green-500' : 'border-white/40'}`} />
          <div className={`absolute bottom-0 right-0 w-12 h-12 border-b-4 border-r-4 rounded-br-3xl transition-colors ${isStable ? 'border-green-500' : 'border-white/40'}`} />
        </div>
      </div>

      {/* 4. Bottom Controls UI */}
      <div className="absolute bottom-0 left-0 right-0 p-8 pb-16 flex justify-around items-center z-50 pointer-events-none bg-gradient-to-t from-black/90 to-transparent">
        <button
          onClick={() => fileRef.current?.click()}
          className="w-14 h-14 bg-white/10 backdrop-blur-2xl rounded-2xl flex items-center justify-center text-white pointer-events-auto active:bg-white/30 transition-all"
        >
          <ImageIcon size={28} />
        </button>
        <input ref={fileRef} type="file" accept="image/*" className="hidden" onChange={handleFileCapture} />

        {/* Shutter Button - FORCED VISIBILITY */}
        <button
          onClick={captureImage}
          className={`w-24 h-24 rounded-full border-[6px] flex items-center justify-center transition-all pointer-events-auto active:scale-90 ${isStable ? 'border-green-500 scale-110 shadow-[0_0_30px_rgba(34,197,94,0.4)]' : 'border-white shadow-lg'}`}
          style={{ backgroundColor: 'transparent' }}
        >
          <div className={`w-18 h-18 rounded-full transition-colors ${isStable ? 'bg-green-500' : 'bg-white'}`} />
        </button>

        <div className="w-14 h-14 flex items-center justify-center text-white/20">
          <CameraIcon size={28} />
        </div>
      </div>
    </div>
  )
}
