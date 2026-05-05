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
    <PageShell bottomNav={false}>
      <TopBar title="Seleziona Negozio" back />

      <div className="px-5 pt-4">
        {/* Search */}
        <div className="relative mb-4">
          <Search size={18} className="absolute left-3.5 top-1/2 -translate-y-1/2 text-gray-400" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Cerca negozio o città..."
            className="w-full h-11 pl-10 pr-4 bg-gray-100 rounded-xl text-[15px] outline-none focus:ring-2 focus:ring-accent/30 transition-shadow placeholder:text-gray-400"
          />
        </div>

        {/* Add new */}
        <button className="w-full flex items-center gap-3 p-3.5 mb-3 border border-dashed border-gray-300 rounded-2xl text-accent active:bg-accent-light transition-colors">
          <div className="w-10 h-10 bg-accent-light rounded-xl flex items-center justify-center">
            <Plus size={20} />
          </div>
          <span className="text-[15px] font-medium">Aggiungi nuovo negozio</span>
        </button>

        {/* Store list */}
        <div className="space-y-1.5">
          {filtered.map((store) => (
            <button
              key={store.id}
              onClick={() => handleSelect(store)}
              className="w-full flex items-center gap-3 p-3.5 bg-gray-50 rounded-2xl active:bg-gray-100 transition-colors text-left"
            >
              <div className="w-10 h-10 bg-white rounded-xl flex items-center justify-center border border-gray-200">
                <MapPin size={18} className="text-gray-600" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-[15px] font-semibold truncate">{store.name}</p>
                <p className="text-xs text-gray-500 truncate">{store.address}, {store.city}</p>
              </div>
              <ChevronRight size={16} className="text-gray-300 shrink-0" />
            </button>
          ))}
        </div>

        {filtered.length === 0 && (
          <div className="text-center py-12">
            <p className="text-gray-400 text-sm">Nessun negozio trovato</p>
          </div>
        )}
      </div>
    </PageShell>
  )
}
