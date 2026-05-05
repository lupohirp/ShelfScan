import { useState } from 'react'
import { mockStores } from '../../lib/mock-data'
import { Search, Plus, MapPin, MoreHorizontal } from 'lucide-react'

export default function AdminStores() {
  const [query, setQuery] = useState('')

  const filtered = mockStores.filter(
    (s) =>
      s.name.toLowerCase().includes(query.toLowerCase()) ||
      s.city.toLowerCase().includes(query.toLowerCase())
  )

  return (
    <div className="max-w-6xl mx-auto animate-fade-in">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Negozi</h1>
        <button className="h-9 px-3 bg-black text-white rounded-xl text-[13px] font-medium flex items-center gap-1.5 hover:bg-gray-800 transition-colors">
          <Plus size={14} />
          Aggiungi
        </button>
      </div>

      <div className="bg-white rounded-2xl border border-gray-100">
        <div className="p-3 border-b border-gray-100">
          <div className="relative">
            <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Cerca negozio..."
              className="w-full h-9 pl-9 pr-4 bg-gray-50 rounded-lg text-[13px] outline-none focus:ring-2 focus:ring-accent/30 placeholder:text-gray-400"
            />
          </div>
        </div>

        <div className="divide-y divide-gray-50">
          {filtered.map((store) => (
            <div key={store.id} className="flex items-center gap-3 px-4 py-3.5 hover:bg-gray-50/50 transition-colors">
              <div className="w-10 h-10 bg-gray-100 rounded-xl flex items-center justify-center shrink-0">
                <MapPin size={18} className="text-gray-500" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-[14px] font-medium truncate">{store.name}</p>
                <p className="text-xs text-gray-500">{store.address}, {store.city}</p>
              </div>
              <span className="text-xs text-gray-400 hidden sm:block">12 check</span>
              <button className="text-gray-400 hover:text-gray-600 p-1">
                <MoreHorizontal size={16} />
              </button>
            </div>
          ))}
        </div>

        <div className="px-4 py-3 border-t border-gray-100 text-xs text-gray-500">
          {filtered.length} negozi
        </div>
      </div>
    </div>
  )
}
