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
    <PageShell>
      <TopBar title="SELEZIONA STORE" back />

      <div className="px-8 pt-10 pb-40">
        {/* Search */}
        <div className="relative mb-12">
          <Search size={20} className="absolute left-0 top-1/2 -translate-y-1/2 text-black" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="CERCA NEGOZIO O CITTÀ..."
            className="w-full pl-8 py-4 border-b border-gray-100 text-[12px] font-black uppercase tracking-[0.2em] outline-none focus:border-black transition-colors bg-transparent placeholder:text-gray-300"
          />
        </div>

        {/* Add new */}
        <button className="w-full flex items-center justify-center gap-4 py-6 mb-12 border border-gray-100 bg-gray-50/30 text-black active:bg-gray-100 transition-all group">
          <Plus size={20} strokeWidth={1} />
          <span className="text-[12px] font-black uppercase tracking-[0.2em]">Nuovo Store</span>
        </button>

        <div className="mb-8">
          <h2 className="text-[10px] font-black text-gray-400 uppercase tracking-[0.3em]">Negozi suggeriti</h2>
        </div>

        {/* Store list */}
        <div className="space-y-0">
          {filtered.map((store) => (
            <button
              key={store.id}
              onClick={() => handleSelect(store)}
              className="w-full py-8 border-b border-gray-50 flex items-center justify-between active:opacity-50 transition-opacity text-left group"
            >
              <div className="flex-1 min-w-0 pr-4">
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-[14px] font-black uppercase tracking-[0.1em] text-black">{store.name}</span>
                  <div className="w-1.5 h-1.5 bg-black rounded-full opacity-0 group-hover:opacity-100 transition-opacity" />
                </div>
                <div className="flex items-center gap-2 text-gray-400 font-bold uppercase text-[10px] tracking-widest">
                  <MapPin size={10} />
                  <span>{store.city} — {store.address}</span>
                </div>
              </div>
              <ChevronRight size={18} className="text-black opacity-30 group-hover:opacity-100 transition-opacity" strokeWidth={1.5} />
            </button>
          ))}
        </div>

        {filtered.length === 0 && (
          <div className="text-center py-20 grayscale opacity-40">
            <Search size={32} className="mx-auto mb-4" strokeWidth={1} />
            <p className="text-[12px] font-black uppercase tracking-[0.2em]">Nessun risultato</p>
          </div>
        )}
      </div>
    </PageShell>
  )
}
