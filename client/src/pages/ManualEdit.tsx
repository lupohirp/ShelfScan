import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useScan } from '../store/scan'
import TopBar from '../components/TopBar'
import PageShell from '../components/PageShell'
import { Search, Check, Plus, Minus, Package } from 'lucide-react'

export default function ManualEdit() {
  const [query, setQuery] = useState('')
  const navigate = useNavigate()
  const session = useScan((s) => s.currentSession)
  const toggleProduct = useScan((s) => s.toggleProduct)

  if (!session) {
    navigate('/home', { replace: true })
    return null
  }

  const allProducts = [...session.foundProducts, ...session.missingProducts]
  const filtered = query
    ? allProducts.filter(
        (p) =>
          p.name.toLowerCase().includes(query.toLowerCase()) ||
          p.sku.toLowerCase().includes(query.toLowerCase())
      )
    : allProducts

  const handleConfirm = () => {
    navigate('/scan/results', { replace: true })
  }

  return (
    <PageShell bottomNav={false}>
      <TopBar title="Modifica Manuale" back />

      <div className="px-5 pt-4">
        {/* Search */}
        <div className="relative mb-4">
          <Search size={18} className="absolute left-3.5 top-1/2 -translate-y-1/2 text-gray-400" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Cerca per nome o SKU..."
            className="w-full h-11 pl-10 pr-4 bg-gray-100 rounded-xl text-[15px] outline-none focus:ring-2 focus:ring-accent/30 transition-shadow placeholder:text-gray-400"
          />
        </div>

        <p className="text-xs text-gray-500 mb-3 ml-1">
          Tocca un prodotto per cambiarne lo stato
        </p>

        {/* Product list */}
        <div className="space-y-2 pb-28">
          {filtered.map((product) => {
            const isFound = session.foundProducts.some((p) => p.id === product.id)
            return (
              <button
                key={product.id}
                onClick={() => toggleProduct(product.id)}
                className={`w-full flex items-center gap-3 p-3.5 rounded-2xl transition-all text-left ${
                  isFound
                    ? 'bg-success-light/50 border border-success/20'
                    : 'bg-gray-50 border border-transparent'
                }`}
              >
                <div
                  className={`w-10 h-10 rounded-xl flex items-center justify-center shrink-0 ${
                    isFound ? 'bg-success text-white' : 'bg-gray-200 text-gray-400'
                  }`}
                >
                  {isFound ? <Check size={18} /> : <Package size={18} />}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-[14px] font-semibold truncate">{product.name}</p>
                  <p className="text-xs text-gray-500">{product.sku} · {product.category}</p>
                </div>
                <div
                  className={`w-8 h-8 rounded-full flex items-center justify-center shrink-0 ${
                    isFound
                      ? 'bg-danger-light text-danger'
                      : 'bg-success-light text-success'
                  }`}
                >
                  {isFound ? <Minus size={16} /> : <Plus size={16} />}
                </div>
              </button>
            )
          })}
        </div>
      </div>

      {/* Confirm button */}
      <div className="fixed bottom-0 left-0 right-0 bg-white/80 backdrop-blur-xl border-t border-gray-200 p-4 safe-bottom">
        <div className="max-w-lg mx-auto">
          <button
            onClick={handleConfirm}
            className="w-full h-12 bg-black text-white rounded-xl font-semibold text-[15px] active:scale-[0.98] transition-transform"
          >
            Conferma modifiche
          </button>
        </div>
      </div>
    </PageShell>
  )
}
