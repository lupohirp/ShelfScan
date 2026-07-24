import { useState, useRef, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useScan } from '../store/scan'
import { X, Zap, ImageIcon, ScanLine, Camera as CameraIcon, Send } from 'lucide-react'
import type { CheckSession } from '../types'
import { validateCapturedImage } from '../lib/imageValidation'

export default function Camera() {
  const navigate = useNavigate()
  const selectedStore = useScan((s) => s.selectedStore)
  const setSession = useScan((s) => s.setSession)
  
  const [analyzing, setAnalyzing] = useState(false)
  const [capturedImages, setCapturedImages] = useState<string[]>([])
  const [pitch, setPitch] = useState(0)
  const [roll, setRoll] = useState(0)
  const [isStable, setIsStable] = useState(false)
  const [stream, setStream] = useState<MediaStream | null>(null)
  
  // Validation and guidance states
  const [pendingImage, setPendingImage] = useState<string | null>(null)
  const [validationErrors, setValidationErrors] = useState<string[]>([])
  const [showWarning, setShowWarning] = useState(false)
  const [realtimeFeedback, setRealtimeFeedback] = useState<string>('Allinea lo scaffale nel riquadro')
  const [realtimeFeedbackType, setRealtimeFeedbackType] = useState<'info' | 'warning' | 'success'>('info')
  const [cameraError, setCameraError] = useState<string | null>(null)
  const [torchSupported, setTorchSupported] = useState(false)
  const [torchOn, setTorchOn] = useState(false)

  // Streaming progress states
  const [streamProgressMessage, setStreamProgressMessage] = useState<string>('Avvio analisi AI...')
  const [streamCurrentStep, setStreamCurrentStep] = useState<{ current: number; total: number }>({ current: 0, total: 1 })
  const [streamMatchedItems, setStreamMatchedItems] = useState<{ sku: string; name: string; imageUrl: string; cropUrl: string; score: number }[]>([])

  const videoRef = useRef<HTMLVideoElement>(null)
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const fileRef = useRef<HTMLInputElement>(null)
  const stabilityTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  // Request 4K Camera & Lifecycle Management
  const setupCamera = useCallback(async () => {
    try {
      if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
        throw new Error("navigator.mediaDevices is undefined. Secure context (HTTPS) or localhost is required.")
      }
      const currentStream = await navigator.mediaDevices.getUserMedia({
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
      const track = currentStream.getVideoTracks()[0]
      // getCapabilities().torch is Android-only (Chromium). Safari has no torch API.
      const caps = (track?.getCapabilities?.() ?? {}) as MediaTrackCapabilities & { torch?: boolean }
      setTorchSupported(Boolean(caps.torch))
    } catch (err) {
      console.error('Error accessing camera:', err)
      const msg = err instanceof Error ? err.message : String(err)
      if (msg.toLowerCase().includes("secure context") || msg.toLowerCase().includes("undefined")) {
        setCameraError("L'accesso alla fotocamera richiede una connessione sicura (HTTPS) o localhost. Apri l'app tramite localhost o una connessione HTTPS.")
      } else {
        setCameraError("Accesso alla fotocamera negato o non disponibile. Verifica i permessi del browser.")
      }
    }
  }, [])

  useEffect(() => {
    setupCamera()
    
    return () => {
      if (stream) {
        stream.getTracks().forEach(track => track.stop())
        console.log('Camera stopped')
      }
    }
  }, [setupCamera])

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

  // Real-time Preview Analytics and guidance feedback
  useEffect(() => {
    if (analyzing || showWarning) return

    const interval = setInterval(() => {
      if (!videoRef.current) return
      const video = videoRef.current
      if (video.videoWidth === 0) return

      // Downsample and process on small offscreen canvas
      const canvas = document.createElement('canvas')
      canvas.width = 120
      canvas.height = 120
      const ctx = canvas.getContext('2d')
      if (!ctx) return
      ctx.drawImage(video, 0, 0, 120, 120)
      
      try {
        const imgData = ctx.getImageData(0, 0, 120, 120)
        const data = imgData.data
        
        let brightnessSum = 0
        const gray = new Float32Array(120 * 120)
        for (let i = 0; i < data.length; i += 4) {
          const r = data[i]
          const g = data[i+1]
          const b = data[i+2]
          const grayVal = 0.299 * r + 0.587 * g + 0.114 * b
          gray[i/4] = grayVal
          brightnessSum += grayVal
        }
        const brightness = brightnessSum / gray.length
        
        let edgeCount = 0
        const edgeThreshold = 30
        const width = 120
        const height = 120
        for (let y = 1; y < height - 1; y++) {
          for (let x = 1; x < width - 1; x++) {
            const idx = y * width + x
            const gx = 
              -1 * gray[idx - width - 1] + 1 * gray[idx - width + 1] +
              -2 * gray[idx - 1]         + 2 * gray[idx + 1] +
              -1 * gray[idx + width - 1] + 1 * gray[idx + width + 1]
            const gy = 
              -1 * gray[idx - width - 1] - 2 * gray[idx - width] - 1 * gray[idx - width + 1] +
              1 * gray[idx + width - 1]  + 2 * gray[idx + width]  + 1 * gray[idx + width + 1]
            const mag = Math.sqrt(gx*gx + gy*gy)
            if (mag > edgeThreshold) {
              edgeCount++
            }
          }
        }
        const edgeDensity = edgeCount / ((width-2)*(height-2))
        
        const isLevel = Math.abs(pitch - 90) < 7 && Math.abs(roll) < 7
        
        if (brightness < 45) {
          setRealtimeFeedback(torchSupported ? "Troppo buio: attiva il flash o cerca più luce." : "Troppo buio: cerca una zona più illuminata.")
          setRealtimeFeedbackType("warning")
        } else if (brightness > 225) {
          setRealtimeFeedback("Troppo luminoso o riflesso: cambia angolazione.")
          setRealtimeFeedbackType("warning")
        } else if (!isLevel) {
          setRealtimeFeedback("Allinea la livella al centro.")
          setRealtimeFeedbackType("info")
        } else if (edgeDensity < 0.055) {
          setRealtimeFeedback("Avvicinati allo scaffale.")
          setRealtimeFeedbackType("warning")
        } else {
          setRealtimeFeedback("Ottimo! Tieni fermo...")
          setRealtimeFeedbackType("success")
        }
      } catch (e) {
        // Skip analytics errors during canvas read
      }
    }, 500)

    return () => clearInterval(interval)
  }, [pitch, roll, analyzing, showWarning, torchSupported])

  const captureImage = useCallback(async () => {
    console.log('Capture triggered')
    if (!videoRef.current || !canvasRef.current) {
      console.error('Video or Canvas ref missing')
      return
    }
    
    const video = videoRef.current
    const canvas = canvasRef.current
    
    const isPortraitScreen = window.innerHeight > window.innerWidth
    const isLandscapeVideo = video.videoWidth > video.videoHeight
    const needsRotation = isPortraitScreen && isLandscapeVideo

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    let srcW = video.videoWidth
    let srcH = video.videoHeight
    if (needsRotation) {
      srcW = video.videoHeight
      srcH = video.videoWidth
    }

    const maxW = 1920
    const maxH = 1080
    let targetW = srcW
    let targetH = srcH
    if (targetW > maxW || targetH > maxH) {
      const ratio = Math.min(maxW / targetW, maxH / targetH)
      targetW = Math.round(targetW * ratio)
      targetH = Math.round(targetH * ratio)
    }

    canvas.width = targetW
    canvas.height = targetH

    if (needsRotation) {
      ctx.save()
      ctx.translate(canvas.width / 2, canvas.height / 2)
      ctx.rotate(90 * Math.PI / 180)
      ctx.drawImage(video, -targetH / 2, -targetW / 2, targetH, targetW)
      ctx.restore()
    } else {
      ctx.drawImage(video, 0, 0, targetW, targetH)
    }
    
    const imageData = canvas.toDataURL('image/jpeg', 0.85)
    
    // Validate image quality
    validateCapturedImage(imageData, (result) => {
      if (result.isValid) {
        setCapturedImages(prev => [...prev, imageData])
        // Visual flash effect on success
        const flash = document.createElement('div')
        flash.className = 'fixed inset-0 bg-white z-[200] animate-out fade-out duration-300'
        document.body.appendChild(flash)
        setTimeout(() => flash.remove(), 300)
      } else {
        setPendingImage(imageData)
        setValidationErrors(result.errors)
        setShowWarning(true)
      }
    })
  }, [videoRef, canvasRef])

  // Stability Detection & Auto-capture
  useEffect(() => {
    const isLevel = Math.abs(pitch - 90) < 7 && Math.abs(roll) < 7
    setIsStable(isLevel)

    if (isLevel && !analyzing && capturedImages.length === 0 && !showWarning) { 
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
  }, [pitch, roll, captureImage, analyzing, capturedImages.length, showWarning])

  const handleAnalysis = async () => {
    if (analyzing || capturedImages.length === 0) return
    setAnalyzing(true)

    // Stop camera stream and turn off hardware sensor/torch during AI analysis
    if (stream) {
      stream.getTracks().forEach(track => track.stop())
      setStream(null)
      if (videoRef.current) {
        videoRef.current.srcObject = null
      }
      console.log('Camera tracks stopped for AI analysis phase')
    }

    setStreamMatchedItems([])
    setStreamCurrentStep({ current: 0, total: capturedImages.length })
    setStreamProgressMessage(`Invio di ${capturedImages.length} foto all'IA...`)

    try {
      const formData = new FormData()
      
      for (let i = 0; i < capturedImages.length; i++) {
        const res = await fetch(capturedImages[i])
        const blob = await res.blob()
        formData.append('images', blob, `photo_${i}.jpg`)
      }
      
      const getApiUrl = () => {
        const envUrl = import.meta.env.VITE_API_URL;
        if (envUrl) {
          return envUrl;
        }
        const hostname = window.location.hostname;
        if (hostname === 'localhost' || hostname === '127.0.0.1' || hostname.startsWith('192.168.') || hostname.startsWith('10.')) {
          return `http://${hostname}:8080`;
        }
        if (import.meta.env.PROD) {
          const baseHost = hostname.replace(/^admin-/, '');
          return `https://api-${baseHost}`;
        }
        return `http://${hostname}:8080`;
      };
      const apiBase = getApiUrl();
      const analysisResponse = await fetch(`${apiBase}/analyze?stream=true`, {
        method: 'POST',
        headers: {
          'Accept': 'application/x-ndjson'
        },
        body: formData
      })
      
      if (!analysisResponse.ok) {
        const errText = await analysisResponse.text()
        throw new Error(`Analyze API failed: ${errText}`)
      }

      if (!analysisResponse.body) {
        throw new Error('ReadableStream non supportato nella risposta del browser.')
      }

      const reader = analysisResponse.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''
      let analysisData: {
        found: {name: string, sku: string, imageUrl: string, cropUrl?: string, score: number, count?: number}[],
        missing: {name: string, sku: string, imageUrl: string}[],
        imageResults: { detections: { desc: string, box?: number[], box_2d?: number[], crop_url?: string, sku?: string }[] }[]
      } | null = null

      while (true) {
        const { value, done } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() || ''

        for (const line of lines) {
          const trimmed = line.trim()
          if (!trimmed) continue
          try {
            const event = JSON.parse(trimmed)
            if (event.type === 'status') {
              setStreamProgressMessage(event.message)
            } else if (event.type === 'image_start') {
              setStreamCurrentStep({ current: event.image, total: event.total })
              setStreamProgressMessage(`Elaborazione foto ${event.image} di ${event.total}...`)
            } else if (event.type === 'item_matched') {
              setStreamMatchedItems(prev => [
                ...prev,
                { sku: event.sku, name: event.name, imageUrl: event.imageUrl, cropUrl: event.cropUrl, score: event.score }
              ])
              setStreamProgressMessage(`Trovato articolo: ${event.name}`)
            } else if (event.type === 'complete') {
              analysisData = event.data
            }
          } catch (e) {
            console.warn('NDJSON line parse error:', line, e)
          }
        }
      }

      if (buffer.trim()) {
        try {
          const event = JSON.parse(buffer.trim())
          if (event.type === 'complete') analysisData = event.data
        } catch (e) {
          console.warn('NDJSON trailing line parse error:', e)
        }
      }

      if (!analysisData) {
        throw new Error('Nessun risultato ricevuto dallo stream dell\'IA.')
      }

      console.log('AI Analysis streaming complete:', analysisData)
      
      const foundProducts: any[] = (analysisData.found || []).map(r => ({
        id: 'found-' + r.name,
        sku: r.sku || '',
        confidence: Math.round(r.score * 100),
        name: r.name,
        category: 'identified',
        imageUrl: r.imageUrl,
        cropUrl: r.cropUrl || '',
        status: 'active',
        count: r.count || 1
      }))

      const missingProducts: any[] = (analysisData.missing || []).map(m => ({
        id: 'missing-' + m.name,
        sku: m.sku || '',
        name: m.name,
        category: 'missing',
        imageUrl: m.imageUrl,
        status: 'active'
      }))

      const analyzedImages = (analysisData.imageResults || []).map((ir, idx) => ({
        capturedImage: capturedImages[idx] || capturedImages[0],
        detections: (ir.detections || []).map(d => ({
          desc: d.desc,
          box: d.box || d.box_2d || [],
          crop_url: d.crop_url || '',
          sku: d.sku || ''
        }))
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
        analyzedImages: analyzedImages
      }
      
      setSession(session)
      navigate('/scan/results', { replace: true })
    } catch (err) {
      console.error('Shelf Analysis failed:', err)
      const errorMsg = err instanceof Error ? err.message : String(err)
      alert('Errore durante l\'analisi della vetrina: ' + errorMsg)
      setAnalyzing(false)
      setupCamera()
    }
  }

  const readFileAsDataUrl = (file: File): Promise<string> => {
    return new Promise((resolve, reject) => {
      const reader = new FileReader()
      reader.onload = (ev) => resolve(ev.target?.result as string)
      reader.onerror = reject
      reader.readAsDataURL(file)
    })
  }

  const handleFileCapture = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files
    if (files && files.length > 0) {
      for (const file of Array.from(files)) {
        try {
          let imageData: string
          if ('createImageBitmap' in window) {
            const bitmap = await createImageBitmap(file, { imageOrientation: 'from-image' })
            const canvas = document.createElement('canvas')
            
            let w = bitmap.width
            let h = bitmap.height
            const maxW = 1920
            const maxH = 1080
            if (w > maxW || h > maxH) {
              const ratio = Math.min(maxW / w, maxH / h)
              w = Math.round(w * ratio)
              h = Math.round(h * ratio)
            }

            canvas.width = w
            canvas.height = h
            const ctx = canvas.getContext('2d')
            if (ctx) {
              ctx.drawImage(bitmap, 0, 0, w, h)
              imageData = canvas.toDataURL('image/jpeg', 0.85)
            } else {
              imageData = await readFileAsDataUrl(file)
            }
          } else {
            imageData = await readFileAsDataUrl(file)
          }

          validateCapturedImage(imageData, (result) => {
            if (result.isValid) {
              setCapturedImages(prev => [...prev, imageData])
            } else {
              setPendingImage(imageData)
              setValidationErrors(result.errors)
              setShowWarning(true)
            }
          })
        } catch (err) {
          console.error('Error processing gallery image:', err)
          try {
            const fallbackData = await readFileAsDataUrl(file)
            setCapturedImages(prev => [...prev, fallbackData])
          } catch (e) {
            // ignore fallback error
          }
        }
      }
    }
  }



  return (
    <div className="fixed inset-0 bg-black overflow-hidden select-none">
      <video
        ref={videoRef}
        autoPlay
        playsInline
        muted
        className="absolute inset-0 w-full h-full object-cover z-0"
      />
      
      <canvas ref={canvasRef} className="hidden" />

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

        {torchSupported ? (
          <button
            onClick={async () => {
              const track = stream?.getVideoTracks?.()[0]
              if (!track) return
              try {
                const next = !torchOn
                await track.applyConstraints({ advanced: [{ torch: next } as MediaTrackConstraintSet] })
                setTorchOn(next)
              } catch (err) {
                console.warn('Torch toggle failed:', err)
              }
            }}
            className={`w-12 h-12 backdrop-blur-xl rounded-full flex items-center justify-center pointer-events-auto active:scale-95 transition-colors ${torchOn ? 'bg-amber-400 text-black' : 'bg-black/50 text-white'}`}
            aria-label={torchOn ? 'Disattiva flash' : 'Attiva flash'}
          >
            <Zap size={20} />
          </button>
        ) : (
          <div className="w-12 h-12" />
        )}
      </div>

      <div className="absolute inset-0 flex flex-col items-center justify-center z-10 pointer-events-none">
        {/* Dynamic Status Banner */}
        <div className="mb-6 text-center px-6 transition-all duration-300 max-w-[85%]">
          <div className={`inline-flex items-center gap-2 px-4 py-2 rounded-full backdrop-blur-md border text-xs font-semibold uppercase tracking-wider shadow-lg ${
            realtimeFeedbackType === 'success' ? 'bg-green-500/20 text-green-400 border-green-500/30' :
            realtimeFeedbackType === 'warning' ? 'bg-amber-500/20 text-amber-400 border-amber-500/30' :
            'bg-black/50 text-white/95 border-white/10'
          }`}>
            <span className={`w-2 h-2 rounded-full ${
              realtimeFeedbackType === 'success' ? 'bg-green-500 animate-pulse' :
              realtimeFeedbackType === 'warning' ? 'bg-amber-500 animate-bounce' :
              'bg-white/60'
            }`} />
            {realtimeFeedback}
          </div>
        </div>

        {/* Leveling Indicator Bubble */}
        <div className={`w-28 h-28 border-2 rounded-full flex items-center justify-center transition-all duration-500 mb-6 ${isStable ? 'border-green-500 bg-green-500/10 scale-95 opacity-55' : 'border-white/20'}`}>
          <div 
            className={`w-3.5 h-3.5 rounded-full transition-colors duration-300 ${isStable ? 'bg-green-500' : 'bg-red-500'}`}
            style={{ transform: `translate(${roll * 2}px, ${(pitch - 90) * 2}px)` }}
          />
        </div>

        {/* Framing Grid / Target Box */}
        <div className={`relative w-[90%] aspect-[4/5] border-2 rounded-3xl transition-all duration-500 overflow-hidden ${isStable ? 'border-green-500 scale-105 shadow-[0_0_40px_rgba(34,197,94,0.15)]' : 'border-white/20'}`}>
          <div className={`absolute top-0 left-0 w-16 h-16 border-t-4 border-l-4 rounded-tl-3xl transition-colors ${isStable ? 'border-green-500' : 'border-white/40'}`} />
          <div className={`absolute top-0 right-0 w-16 h-16 border-t-4 border-r-4 rounded-tr-3xl transition-colors ${isStable ? 'border-green-500' : 'border-white/40'}`} />
          <div className={`absolute bottom-0 left-0 w-16 h-16 border-b-4 border-l-4 rounded-bl-3xl transition-colors ${isStable ? 'border-green-500' : 'border-white/40'}`} />
          <div className={`absolute bottom-0 right-0 w-16 h-16 border-b-4 border-r-4 rounded-br-3xl transition-colors ${isStable ? 'border-green-500' : 'border-white/40'}`} />
          
          {/* Rule of Thirds Grid Lines */}
          <div className="absolute inset-0 grid grid-cols-3 grid-rows-3 pointer-events-none opacity-20">
            <div className="border-r border-b border-white border-dashed" />
            <div className="border-r border-b border-white border-dashed" />
            <div className="border-b border-white border-dashed" />
            <div className="border-r border-b border-white border-dashed" />
            <div className="border-r border-b border-white border-dashed" />
            <div className="border-b border-white border-dashed" />
            <div className="border-r border-white border-dashed" />
            <div className="border-r border-white border-dashed" />
            <div />
          </div>

          {/* Leveling bar inside the guide box */}
          <div 
            className={`absolute left-6 right-6 h-[1.5px] top-1/2 -translate-y-1/2 transition-colors duration-300 pointer-events-none ${Math.abs(roll) < 5 ? 'bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.6)]' : 'bg-white/30'}`}
            style={{ transform: `rotate(${roll}deg)` }}
          />

          {isStable && (
            <div className="absolute inset-0 bg-gradient-to-b from-transparent via-green-500/10 to-transparent animate-scan-line rounded-3xl" />
          )}
        </div>
      </div>

      {capturedImages.length > 0 && (
        <div className="absolute bottom-40 left-0 right-0 px-6 z-50 flex gap-3 overflow-x-auto pb-4 no-scrollbar">
          {capturedImages.map((img, idx) => (
            <div key={idx} className="relative w-20 h-24 shrink-0 rounded-lg border-2 border-white/20 overflow-hidden bg-black shadow-xl animate-in fade-in slide-in-from-bottom-2">
              <img src={img} className="w-full h-full object-cover" />
              <button 
                onClick={() => setCapturedImages(prev => prev.filter((_, i) => i !== idx))}
                className="absolute top-1 right-1 w-5 h-5 bg-black/60 rounded-full flex items-center justify-center pointer-events-auto active:scale-90"
              >
                <X size={12} className="text-white" />
              </button>
            </div>
          ))}
        </div>
      )}

      <div className="absolute bottom-0 left-0 right-0 p-8 pb-16 flex justify-around items-center z-50 pointer-events-none bg-gradient-to-t from-black/90 to-transparent">
        <button
          onClick={() => fileRef.current?.click()}
          className="w-14 h-14 bg-white/10 backdrop-blur-2xl rounded-2xl flex items-center justify-center text-white pointer-events-auto active:bg-white/30 transition-all"
        >
          <ImageIcon size={28} />
        </button>
        <input ref={fileRef} type="file" accept="image/*" multiple className="hidden" onChange={handleFileCapture} />

        <button
          onClick={captureImage}
          aria-disabled={realtimeFeedbackType === 'warning'}
          className={`w-24 h-24 rounded-full border-[6px] flex items-center justify-center transition-all pointer-events-auto active:scale-90 ${
            realtimeFeedbackType === 'warning'
              ? 'border-white/30 opacity-50'
              : isStable
                ? 'border-green-500 scale-110 shadow-[0_0_30px_rgba(34,197,94,0.4)]'
                : 'border-white shadow-lg'
          }`}
          style={{ backgroundColor: 'transparent' }}
        >
          <div className={`w-18 h-18 rounded-full transition-colors ${
            realtimeFeedbackType === 'warning' ? 'bg-white/40' : isStable ? 'bg-green-500' : 'bg-white'
          }`} />
        </button>

        {capturedImages.length > 0 ? (
          <button
            onClick={handleAnalysis}
            className="w-14 h-14 bg-white text-black rounded-2xl flex items-center justify-center pointer-events-auto active:scale-95 transition-all shadow-[0_0_20px_rgba(255,255,255,0.2)] animate-in zoom-in relative hover:bg-gray-100"
          >
            <Send size={24} />
            <span className="absolute -top-1.5 -right-1.5 bg-black text-white text-[9px] font-bold px-1.5 py-0.5 rounded-full border border-white">
              {capturedImages.length}
            </span>
          </button>
        ) : (
          <div className="w-14 h-14 flex items-center justify-center text-white/20">
            <CameraIcon size={28} />
          </div>
        )}
      </div>

      {/* Camera Access Error Overlay */}
      {cameraError && (
        <div className="absolute inset-0 bg-black/85 backdrop-blur-sm z-40 flex flex-col items-center justify-center p-6 text-center text-white">
          <div className="bg-red-950/40 text-red-200 border border-red-500/30 p-5 rounded-2xl max-w-xs shadow-2xl">
            <X size={28} className="mx-auto mb-3 text-red-500" />
            <h3 className="font-bold text-sm mb-2">Fotocamera non raggiungibile</h3>
            <p className="text-xs text-white/70 leading-relaxed">{cameraError}</p>
            <p className="text-xs text-white/50 mt-4 font-semibold">Puoi comunque caricare una foto dalla galleria usando l'icona in basso.</p>
          </div>
        </div>
      )}

      {/* Error Handling and Retake / Skip Overlay */}
      {showWarning && pendingImage && (
        <div className="fixed inset-0 bg-black/90 backdrop-blur-md z-[200] flex flex-col justify-between p-6 animate-in fade-in duration-300">
          <div className="text-center pt-8">
            <div className="inline-flex items-center gap-2 bg-red-500/20 text-red-400 border border-red-500/30 px-4 py-2 rounded-full text-xs font-bold tracking-wider uppercase">
              <span className="w-2 h-2 rounded-full bg-red-500 animate-pulse" />
              Qualità immagine
            </div>
            <h3 className="text-white text-lg font-bold mt-4">La foto non soddisfa gli standard di qualità</h3>
            <p className="text-white/60 text-xs mt-2 px-6">Abbiamo rilevato problemi che possono influire sull'accuratezza del riconoscimento.</p>
          </div>

          <div className="max-w-xs w-full mx-auto my-auto flex flex-col gap-6">
            {/* Captured Image Preview with overlayed errors */}
            <div className="relative aspect-[3/4] rounded-2xl overflow-hidden border border-white/10 shadow-2xl">
              <img src={pendingImage} className="w-full h-full object-cover" />
              <div className="absolute inset-0 bg-gradient-to-t from-black/90 via-black/30 to-transparent flex flex-col justify-end p-5">
                <div className="space-y-3">
                  {validationErrors.map((err, idx) => (
                    <div key={idx} className="flex items-start gap-2 text-white text-xs">
                      <span className="mt-1 flex-shrink-0 w-1.5 h-1.5 rounded-full bg-red-500" />
                      <span className="font-medium text-left leading-relaxed">{err}</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>

          {/* Action buttons */}
          <div className="max-w-xs w-full mx-auto flex flex-col gap-3 pb-6">
            <button
              onClick={() => {
                setShowWarning(false)
                setPendingImage(null)
                setValidationErrors([])
              }}
              className="w-full py-3.5 bg-white text-black font-semibold rounded-xl flex items-center justify-center gap-2 active:scale-95 transition-all shadow-[0_4px_25px_rgba(255,255,255,0.15)]"
            >
              <CameraIcon size={16} />
              Rifai la foto
            </button>

            <button
              onClick={() => {
                if (pendingImage) {
                  setCapturedImages(prev => [...prev, pendingImage])
                }
                setShowWarning(false)
                setPendingImage(null)
                setValidationErrors([])
              }}
              className="w-full py-3 bg-white/10 text-white/70 hover:text-white font-medium rounded-xl active:scale-95 transition-all border border-white/10 text-xs"
            >
              Ignora avvisi e mantieni comunque
            </button>
          </div>
        </div>
      )}

      {/* Real-time Streaming Analysis Modal - Official Liu Jo Minimalist Palette */}
      {analyzing && (
        <div className="fixed inset-0 z-[99999] bg-black/90 backdrop-blur-2xl flex flex-col items-center justify-center p-6 text-white animate-in fade-in duration-300">
          <div className="w-full max-w-sm bg-black border border-white/20 rounded-2xl p-6 shadow-2xl flex flex-col items-center text-center space-y-6">
            
            {/* Minimalist Monochrome Spinner */}
            <div className="relative w-16 h-16 flex items-center justify-center">
              <div className="absolute inset-0 rounded-full border-2 border-white/10" />
              <div className="absolute inset-0 rounded-full border-2 border-t-white border-r-transparent border-b-transparent border-l-transparent animate-spin" />
              <ScanLine className="w-7 h-7 text-white animate-pulse" strokeWidth={1.5} />
            </div>

            <div className="space-y-2">
              <h3 className="text-sm font-bold tracking-[0.2em] uppercase text-white">
                ANALISI SCAFFALE AI
              </h3>
              <p className="text-xs text-white/60 font-normal leading-relaxed">
                {streamProgressMessage}
              </p>
            </div>

            {/* Progress Bar */}
            {streamCurrentStep.total > 0 && (
              <div className="w-full space-y-2">
                <div className="flex justify-between text-[11px] font-bold tracking-wider text-white/70 uppercase px-0.5">
                  <span>Foto {streamCurrentStep.current} / {streamCurrentStep.total}</span>
                  <span>{Math.round((streamCurrentStep.current / streamCurrentStep.total) * 100)}%</span>
                </div>
                <div className="w-full bg-white/10 h-1.5 rounded-full overflow-hidden border border-white/10">
                  <div 
                    className="bg-white h-full transition-all duration-300 rounded-full"
                    style={{ width: `${Math.max(5, Math.round((streamCurrentStep.current / streamCurrentStep.total) * 100))}%` }}
                  />
                </div>
              </div>
            )}

            {/* Live Streaming Item Matches */}
            {streamMatchedItems.length > 0 && (
              <div className="w-full text-left space-y-2.5 pt-4 border-t border-white/10">
                <div className="flex items-center justify-between">
                  <span className="text-[10px] font-bold uppercase tracking-[0.15em] text-white/60 flex items-center gap-2">
                    <span className="w-1.5 h-1.5 rounded-full bg-white animate-pulse" />
                    ARTICOLI RICONOSCIUTI ({streamMatchedItems.length})
                  </span>
                </div>

                <div className="max-h-44 overflow-y-auto space-y-2 pr-1 custom-scrollbar">
                  {streamMatchedItems.map((item, idx) => (
                    <div key={idx} className="flex items-center gap-3 bg-white/5 p-2.5 rounded-xl border border-white/10 text-xs animate-in slide-in-from-bottom-2">
                      {item.cropUrl || item.imageUrl ? (
                        <img src={item.cropUrl || item.imageUrl} alt={item.name} className="w-10 h-10 object-cover rounded-lg bg-black border border-white/10" />
                      ) : (
                        <div className="w-10 h-10 rounded-lg bg-white/10 flex items-center justify-center font-bold text-white text-xs">
                          LJ
                        </div>
                      )}
                      <div className="flex-1 min-w-0">
                        <p className="font-semibold text-white truncate text-xs">{item.name}</p>
                        <p className="text-[10px] text-white/50 font-mono mt-0.5">SKU: {item.sku}</p>
                      </div>
                      <span className="bg-white/10 text-white text-[10px] font-bold px-2 py-0.5 rounded border border-white/20">
                        {Math.round(item.score * 100)}%
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            )}

          </div>
        </div>
      )}
    </div>
  )
}

