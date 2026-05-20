import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../store/auth'
import { Eye, EyeOff } from 'lucide-react'

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
    <div className="min-h-svh flex flex-col items-center justify-center px-8 bg-white animate-fade-in">
      <div className="w-full max-w-[360px]">
        {/* Logo Section */}
        <div className="flex flex-col items-center mb-16">
          <h1 className="text-[42px] font-black tracking-[-0.05em] leading-none mb-1">LIU·JO</h1>
          <p className="text-[11px] font-bold uppercase tracking-[0.3em] text-gray-400">Inventory Management</p>
        </div>

        {/* Welcome Text */}
        <div className="mb-10 text-center">
          <h2 className="text-[20px] font-bold uppercase tracking-[0.1em] mb-2">Benvenuto</h2>
          <p className="text-[14px] text-gray-500 font-medium">Accedi per gestire il tuo store</p>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="space-y-8">
          <div>
            <label className="block text-[10px] font-black uppercase tracking-[0.2em] text-black mb-1">
              Email
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="Indirizzo email"
              className="lj-input"
              autoComplete="email"
            />
          </div>

          <div>
            <label className="block text-[10px] font-black uppercase tracking-[0.2em] text-black mb-1">
              Password
            </label>
            <div className="relative">
              <input
                type={showPw ? 'text' : 'password'}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Inserisci la password"
                className="lj-input"
                autoComplete="current-password"
              />
              <button
                type="button"
                onClick={() => setShowPw(!showPw)}
                className="absolute right-0 top-1/2 -translate-y-1/2 text-gray-400 p-2"
              >
                {showPw ? <EyeOff size={18} /> : <Eye size={18} />}
              </button>
            </div>
          </div>

          {error && (
            <p className="text-danger text-[13px] text-center font-bold">{error}</p>
          )}

          <div className="pt-4">
            <button
              type="submit"
              disabled={loading}
              className="w-full lj-button flex items-center justify-center gap-3"
            >
              {loading ? (
                <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
              ) : null}
              Accedi
            </button>
          </div>

          <div className="flex flex-col gap-4 text-center">
            <button type="button" className="text-[12px] font-bold uppercase tracking-[0.1em] text-gray-400 hover:text-black transition-colors">
              Password dimenticata?
            </button>
          </div>
        </form>
      </div>
      
      <div className="fixed bottom-10 text-center">
        <p className="text-[10px] font-bold uppercase tracking-[0.15em] text-gray-300">
          Official Partner Portal
        </p>
      </div>
    </div>
  )
}
