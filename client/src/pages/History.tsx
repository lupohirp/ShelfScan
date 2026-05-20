import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { mockHistory } from '../lib/mock-data'
import PageShell from '../components/PageShell'
import { Search, MapPin, Filter } from 'lucide-react'

export default function History() {
  const [query, setQuery] = useState('')
  const navigate = useNavigate()

  const filtered = mockHistory.filter(
    (c) =>
      c.store.name.toLowerCase().includes(query.toLowerCase()) ||
      c.store.city.toLowerCase().includes(query.toLowerCase())
  )

  return (
    <PageShell>
      {/* Header */}
      <div className="px-8 pt-16 pb-10 safe-top border-b border-gray-100 bg-white">
        <h1 className="text-[32px] font-black tracking-tight text-black leading-none">STORICO</h1>
        <p className="text-gray-400 text-[10px] font-black uppercase tracking-[0.2em] mt-2">Check Archive</p>
      </div>

      <div className="px-8 pt-8">
        {/* Search */}
        <div className="flex gap-4 mb-10">
          <div className="relative flex-1">
            <Search size={18} className="absolute left-0 top-1/2 -translate-y-1/2 text-black" />
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="CERCA PER NEGOZIO..."
              className="w-full pl-8 py-3 border-b border-gray-100 text-[12px] font-black uppercase tracking-[0.15em] outline-none focus:border-black transition-colors bg-transparent placeholder:text-gray-300"
            />
          </div>
          <button className="w-12 h-12 border border-gray-100 flex items-center justify-center text-black active:bg-gray-50 transition-colors">
            <Filter size={18} />
          </button>
        </div>

        {/* List */}
        <div className="space-y-0 pb-32">
          {filtered.map((check) => (
            <button
              key={check.id}
              onClick={() => navigate(`/history/${check.id}`)}
              className="w-full py-8 border-b border-gray-50 flex items-center gap-6 active:opacity-50 transition-opacity text-left group"
            >
              {/* Minimal Coverage Indicator */}
              <div className="relative w-14 h-14 shrink-0 border-2 border-black flex items-center justify-center font-black text-[15px] group-hover:bg-black group-hover:text-white transition-colors">
                {check.coverage}%
              </div>

              <div className="flex-1 min-w-0">
                <p className="text-[14px] font-black uppercase tracking-wider text-black truncate">{check.store.name}</p>
                <div className="flex items-center gap-2 mt-1.5">
                  <MapPin size={11} className="text-gray-400 shrink-0" />
                  <span className="text-[11px] text-gray-400 font-bold uppercase tracking-widest truncate">{check.store.city}</span>
                </div>
              </div>

              <div className="flex flex-col items-end gap-2">
                <span className="text-[10px] text-black font-black uppercase">
                  {new Date(check.createdAt).toLocaleDateString('it-IT', {
                    day: 'numeric',
                    month: 'short',
                  })}
                </span>
                <span className="text-[10px] text-gray-400 font-bold uppercase tracking-tighter">
                  {check.foundProducts.length}/{check.foundProducts.length + check.missingProducts.length} ID
                </span>
              </div>
            </button>
          ))}
        </div>

        {filtered.length === 0 && (
          <div className="text-center py-20 grayscale opacity-40">
            <Search size={32} className="mx-auto mb-4" strokeWidth={1} />
            <p className="text-[12px] font-black uppercase tracking-[0.2em]">Nessun archivio trovato</p>
          </div>
        )}
      </div>
    </PageShell>
  )
}
