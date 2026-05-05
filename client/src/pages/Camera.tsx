import { useState, useRef, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useScan } from '../store/scan'
import { mockProducts } from '../lib/mock-data'
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

  // Request 4K Camera
  useEffect(() => {
    async function setupCamera() {
      try {
        const s = await navigator.mediaDevices.getUserMedia({
          video: {
            facingMode: 'environment',
            width: { ideal: 3840 },
            height: { ideal: 2160 },
          },
          audio: false,
        })
        setStream(s)
        if (videoRef.current) {
          videoRef.current.srcObject = s
        }
      } catch (err) {
        console.error('Error accessing camera:', err)
      }
    }
    setupCamera()
    return () => {
      stream?.getTracks().forEach(track => track.stop())
    }
  }, [])

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
    if (!videoRef.current || !canvasRef.current) return
    
    const video = videoRef.current
    const canvas = canvasRef.current
    canvas.width = video.videoWidth
    canvas.height = video.videoHeight
    const ctx = canvas.getContext('2d')
    if (!ctx) return
    
    ctx.drawImage(video, 0, 0)
    const imageData = canvas.toDataURL('image/jpeg', 0.9)
    handleAnalysis(imageData)
  }, [videoRef, canvasRef])

  // Stability Detection & Auto-capture
  useEffect(() => {
    const isLevel = Math.abs(pitch) < 5 && Math.abs(roll) < 5
    setIsStable(isLevel)

    if (isLevel) {
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
  }, [pitch, roll, captureImage])

  const handleAnalysis = async (imageData: string) => {
    setAnalyzing(true)
    
    try {
      // Integration with Google AI SDK
      const apiKey = import.meta.env.VITE_GEMINI_API_KEY
      if (apiKey) {
        const genAI = new GoogleGenerativeAI(apiKey)
        const model = genAI.getGenerativeModel({ model: 'gemini-1.5-flash' })
        
        const base64Data = imageData.split(',')[1]
        const prompt = "Sei un esperto di inventario per gioiellerie. Identifica i gioielli in questa vetrina e confrontali con il database. (Mock: restituisci un report delle discrepanze)"
        
        const result = await model.generateContent([
          prompt,
          { inlineData: { data: base64Data, mimeType: 'image/jpeg' } }
        ])
        console.log(result.response.text())
      }

      // Simulate session creation with mock results for now
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
      }, 1000)
    } catch (err) {
      console.error('AI Analysis failed:', err)
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
    <div className="min-h-svh bg-black flex flex-col relative overflow-hidden">
      {/* Video Stream */}
      <video
        ref={videoRef}
        autoPlay
        playsInline
        muted
        className="absolute inset-0 w-full h-full object-cover"
      />
      
      <canvas ref={canvasRef} className="hidden" />

      {/* Overlay UI */}
      <div className="absolute inset-0 flex flex-col">
        {/* Top Bar */}
        <div className="flex justify-between items-center p-5 pt-12 z-10 bg-gradient-to-b from-black/60 to-transparent">
          <button
            onClick={() => navigate(-1)}
            className="w-10 h-10 bg-black/40 backdrop-blur-md rounded-full flex items-center justify-center text-white"
          >
            <X size={20} />
          </button>

          {selectedStore && (
            <div className="bg-black/40 backdrop-blur-md text-white text-xs font-medium px-3 py-1.5 rounded-full">
              {selectedStore.name}
            </div>
          )}

          <button className="w-10 h-10 bg-black/40 backdrop-blur-md rounded-full flex items-center justify-center text-white">
            <Zap size={18} />
          </button>
        </div>

        {/* Viewfinder & Level */}
        <div className="flex-1 relative flex items-center justify-center">
          {/* Level Bubble */}
          <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
            <div className={`w-32 h-32 border-2 rounded-full flex items-center justify-center transition-colors duration-300 ${isStable ? 'border-green-500 bg-green-500/10' : 'border-white/30'}`}>
              <div 
                className={`w-4 h-4 rounded-full transition-colors duration-300 ${isStable ? 'bg-green-500' : 'bg-red-500'}`}
                style={{
                  transform: `translate(${roll * 2}px, ${pitch * 2}px)`
                }}
              />
            </div>
          </div>

          {/* Central Frame */}
          <div className={`w-[85%] aspect-[3/4] border-2 rounded-lg transition-colors duration-300 pointer-events-none ${isStable ? 'border-green-500/50' : 'border-white/20'}`}>
            <div className="absolute top-0 left-0 w-8 h-8 border-t-4 border-l-4 border-inherit rounded-tl-lg" />
            <div className="absolute top-0 right-0 w-8 h-8 border-t-4 border-r-4 border-inherit rounded-tr-lg" />
            <div className="absolute bottom-0 left-0 w-8 h-8 border-b-4 border-l-4 border-inherit rounded-bl-lg" />
            <div className="absolute bottom-0 right-0 w-8 h-8 border-b-4 border-r-4 border-inherit rounded-br-lg" />
          </div>
        </div>

        {/* Bottom Controls */}
        <div className="pb-10 pt-6 px-10 bg-gradient-to-t from-black/80 to-transparent flex items-center justify-between z-10">
          <button
            onClick={() => fileRef.current?.click()}
            className="w-12 h-12 bg-white/10 backdrop-blur-md rounded-xl flex items-center justify-center text-white"
          >
            <ImageIcon size={22} />
          </button>
          <input ref={fileRef} type="file" accept="image/*" className="hidden" onChange={handleFileCapture} />

          <button
            onClick={captureImage}
            className={`w-20 h-20 rounded-full border-4 flex items-center justify-center transition-all ${isStable ? 'border-green-500 scale-110' : 'border-white'}`}
          >
            <div className={`w-16 h-16 rounded-full transition-colors ${isStable ? 'bg-green-500' : 'bg-white'}`} />
          </button>

          <div className="w-12 h-12 flex items-center justify-center text-white/50">
            <CameraIcon size={22} />
          </div>
        </div>
      </div>
    </div>
  )
}

