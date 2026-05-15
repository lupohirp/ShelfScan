import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../store/auth'
import { Eye, EyeOff, ScanLine } from 'lucide-react'

export default function Login() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [showPw, setShowPw] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const login = useAuth((s) => s.login)
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!email || !password) return setError('Compila tutti i campi')
    setLoading(true)
    setError('')
    try {
      await login(email, password)
      navigate('/home', { replace: true })
    } catch {
      setError('Credenziali non valide')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="relative min-h-svh flex flex-col items-center justify-center px-6 overflow-hidden bg-[#E8E4DF]">
      {/* Animated Background Elements - Softer/Warm */}
      <div className="absolute top-[-10%] left-[-10%] w-[60%] h-[60%] bg-[#B4894D]/10 rounded-full blur-[120px] animate-pulse" />
      <div className="absolute bottom-[-10%] right-[-10%] w-[60%] h-[60%] bg-[#B4894D]/5 rounded-full blur-[120px] animate-pulse delay-1000" />
      
      <div className="w-full max-w-sm z-10 animate-fade-in">
        {/* Logo */}
        <div className="flex flex-col items-center mb-10">
          <div className="w-20 h-20 gradient-accent rounded-[24px] flex items-center justify-center mb-6 shadow-xl shadow-[#B4894D]/20 animate-float">
            <ScanLine size={40} className="text-white" />
          </div>
          <h1 className="text-[36px] font-bold tracking-tight text-[#1D1D1F]">ShelfScan</h1>
          <p className="text-[#6E6E73] text-[15px] mt-1 font-medium tracking-wide uppercase">Boutique Inventory Check</p>
        </div>

        {/* Form Card */}
        <div className="glass-card rounded-[32px] p-8 border-white/60">
          <form onSubmit={handleSubmit} className="space-y-5">
            <div>
              <label className="block text-[12px] font-semibold text-[#6E6E73] uppercase tracking-wider mb-2 ml-1">
                Email
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="nome@azienda.it"
                className="w-full h-13 px-4 bg-white/50 rounded-2xl text-[16px] outline-none border border-transparent focus:border-[#B4894D] focus:ring-4 focus:ring-[#B4894D]/10 transition-all placeholder:text-gray-400"
                autoComplete="email"
              />
            </div>

            <div>
              <label className="block text-[12px] font-semibold text-[#6E6E73] uppercase tracking-wider mb-2 ml-1">
                Password
              </label>
              <div className="relative">
                <input
                  type={showPw ? 'text' : 'password'}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="La tua password"
                  className="w-full h-13 px-4 pr-12 bg-white/50 rounded-2xl text-[16px] outline-none border border-transparent focus:border-[#B4894D] focus:ring-4 focus:ring-[#B4894D]/10 transition-all placeholder:text-gray-400"
                  autoComplete="current-password"
                />
                <button
                  type="button"
                  onClick={() => setShowPw(!showPw)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 p-2 active:scale-90 transition-transform"
                >
                  {showPw ? <EyeOff size={20} /> : <Eye size={20} />}
                </button>
              </div>
            </div>

            {error && (
              <p className="text-danger text-sm text-center font-medium animate-shake">{error}</p>
            )}

            <button
              type="submit"
              disabled={loading}
              className="w-full h-14 bg-[#1D1D1F] text-white rounded-2xl font-bold text-[17px] shadow-lg shadow-black/10 transition-all active:scale-[0.97] hover:brightness-110 disabled:opacity-50 mt-2"
            >
              {loading ? (
                <span className="inline-flex items-center gap-2">
                  <span className="w-5 h-5 border-[3px] border-white/30 border-t-white rounded-full animate-spin" />
                  Accesso...
                </span>
              ) : (
                'Accedi'
              )}
            </button>

            <button type="button" className="w-full text-[#0071E3] text-sm font-semibold py-2 active:opacity-70 transition-opacity">
              Password dimenticata?
            </button>
          </form>
        </div>
        
        <p className="text-center text-[#86868B] text-sm mt-8">
          Non hai un account? <span className="text-[#B4894D] font-bold">Contatta admin</span>
        </p>
      </div>
    </div>
  )
}
