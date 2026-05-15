import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useScan } from '../store/scan'
import { mockStores } from '../lib/mock-data'
import TopBar from '../components/TopBar'
import PageShell from '../components/PageShell'
import { Search, MapPin, Plus, ChevronRight } from 'lucide-react'
import type { Store } from '../types'

export default function StoreSelect() {
  const [query, setQuery] = useState('')
  const setStore = useScan((s) => s.setStore)
  const navigate = useNavigate()

  const filtered = mockStores.filter(
    (s) =>
      s.name.toLowerCase().includes(query.toLowerCase()) ||
      s.city.toLowerCase().includes(query.toLowerCase())
  )

  const handleSelect = (store: Store) => {
    setStore(store)
    navigate('/scan/camera')
  }

  return (
    <PageShell bottomNav={true}>
      <TopBar title="Seleziona Negozio" back />

      <div className="px-6 pt-6 pb-32">
        {/* Search */}
        <div className="relative mb-6">
          <Search size={20} className="absolute left-4 top-1/2 -translate-y-1/2 text-[#86868B]" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Cerca negozio o città..."
            className="w-full h-13 pl-12 pr-4 bg-white/60 border border-transparent rounded-[20px] text-[16px] outline-none shadow-sm shadow-black/5 focus:border-[#B4894D] focus:ring-4 focus:ring-[#B4894D]/5 transition-all placeholder:text-[#AEAEB2]"
          />
        </div>

        {/* Add new */}
        <button className="w-full flex items-center gap-4 p-4 mb-8 border-2 border-dashed border-[#B4894D]/20 rounded-[28px] text-[#B4894D] active:bg-[#B4894D]/5 transition-all group">
          <div className="w-12 h-12 bg-[#B4894D]/10 rounded-2xl flex items-center justify-center group-active:scale-95 transition-transform">
            <Plus size={24} />
          </div>
          <span className="text-[16px] font-bold">Aggiungi nuovo negozio</span>
        </button>

        <div className="flex items-center justify-between mb-4 px-1">
          <h2 className="text-[13px] font-bold text-[#86868B] uppercase tracking-wider">Negozi nelle vicinanze</h2>
          <span className="text-[13px] font-bold text-[#0071E3]">{filtered.length}</span>
        </div>

        {/* Store list */}
        <div className="space-y-4">
          {filtered.map((store) => (
            <button
              key={store.id}
              onClick={() => handleSelect(store)}
              className="w-full flex items-center gap-4 p-5 glass-card rounded-[28px] border-white/60 active:scale-[0.99] transition-all text-left"
            >
              <div className="w-12 h-12 bg-[#F8F2EA] rounded-2xl flex items-center justify-center border border-[#B4894D]/10">
                <MapPin size={20} className="text-[#B4894D]" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-[16px] font-bold text-[#1D1D1F] truncate">{store.name}</p>
                <p className="text-[13px] text-[#86868B] font-medium truncate mt-0.5">{store.address}, {store.city}</p>
              </div>
              <ChevronRight size={18} className="text-[#D2D2D7] shrink-0" />
            </button>
          ))}
        </div>

        {filtered.length === 0 && (
          <div className="text-center py-16">
            <div className="w-16 h-16 bg-black/5 rounded-full flex items-center justify-center mx-auto mb-4">
              <Search size={24} className="text-[#AEAEB2]" />
            </div>
            <p className="text-[#86868B] font-bold">Nessun negozio trovato</p>
          </div>
        )}
      </div>
    </PageShell>
  )
}
