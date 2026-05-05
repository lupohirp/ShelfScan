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
    <div className="min-h-svh flex flex-col items-center justify-center px-6 bg-white">
      <div className="w-full max-w-sm animate-fade-in">
        {/* Logo */}
        <div className="flex flex-col items-center mb-12">
          <div className="w-16 h-16 bg-black rounded-2xl flex items-center justify-center mb-4">
            <ScanLine size={32} className="text-white" />
          </div>
          <h1 className="text-[28px] font-bold tracking-tight">ShelfScan</h1>
          <p className="text-gray-500 text-sm mt-1">Visual Inventory Check</p>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-gray-500 uppercase tracking-wider mb-1.5 ml-1">
              Email
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="nome@azienda.it"
              className="w-full h-12 px-4 bg-gray-100 rounded-xl text-[16px] outline-none focus:ring-2 focus:ring-accent/30 transition-shadow placeholder:text-gray-400"
              autoComplete="email"
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-gray-500 uppercase tracking-wider mb-1.5 ml-1">
              Password
            </label>
            <div className="relative">
              <input
                type={showPw ? 'text' : 'password'}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Inserisci password"
                className="w-full h-12 px-4 pr-12 bg-gray-100 rounded-xl text-[16px] outline-none focus:ring-2 focus:ring-accent/30 transition-shadow placeholder:text-gray-400"
                autoComplete="current-password"
              />
              <button
                type="button"
                onClick={() => setShowPw(!showPw)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 p-1"
              >
                {showPw ? <EyeOff size={20} /> : <Eye size={20} />}
              </button>
            </div>
          </div>

          {error && (
            <p className="text-danger text-sm text-center">{error}</p>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full h-12 bg-black text-white rounded-xl font-semibold text-[16px] transition-all active:scale-[0.98] disabled:opacity-50"
          >
            {loading ? (
              <span className="inline-flex items-center gap-2">
                <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                Accesso...
              </span>
            ) : (
              'Accedi'
            )}
          </button>

          <button type="button" className="w-full text-accent text-sm font-medium py-2">
            Password dimenticata?
          </button>
        </form>
      </div>
    </div>
  )
}
