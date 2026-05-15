import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { mockHistory } from '../lib/mock-data'
import PageShell from '../components/PageShell'
import { Search, MapPin, ChevronRight, Filter } from 'lucide-react'

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
      <div className="px-6 pt-12 pb-6 safe-top">
        <h1 className="text-[36px] font-bold tracking-tight text-[#1D1D1F]">Storico</h1>
      </div>

      <div className="px-6">
        {/* Search */}
        <div className="flex gap-3 mb-6">
          <div className="relative flex-1">
            <Search size={20} className="absolute left-4 top-1/2 -translate-y-1/2 text-[#86868B]" />
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Cerca per negozio..."
              className="w-full h-13 pl-12 pr-4 bg-white/60 border border-transparent rounded-[20px] text-[16px] outline-none shadow-sm shadow-black/5 focus:border-[#B4894D] focus:ring-4 focus:ring-[#B4894D]/5 transition-all placeholder:text-[#AEAEB2]"
            />
          </div>
          <button className="w-13 h-13 glass-card rounded-[20px] flex items-center justify-center text-[#1D1D1F] border-white/60 active:scale-95 transition-transform">
            <Filter size={20} />
          </button>
        </div>

        {/* List */}
        <div className="space-y-4 pb-32">
          {filtered.map((check) => (
            <button
              key={check.id}
              onClick={() => navigate(`/history/${check.id}`)}
              className="w-full glass-card rounded-[28px] p-5 flex items-center gap-4 border-white/60 active:scale-[0.99] transition-all text-left"
            >
              {/* Luxury Coverage Circle */}
              <div className="relative w-14 h-14 shrink-0">
                <svg viewBox="0 0 36 36" className="w-full h-full -rotate-90">
                  <path
                    d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                    fill="none"
                    stroke="#F1F5F9"
                    strokeWidth="3"
                  />
                  <path
                    d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                    fill="none"
                    stroke={check.coverage >= 70 ? '#30D158' : check.coverage >= 40 ? '#FF9500' : '#FF3B30'}
                    strokeWidth="3"
                    strokeDasharray={`${check.coverage}, 100`}
                    strokeLinecap="round"
                  />
                </svg>
                <span className="absolute inset-0 flex items-center justify-center text-[12px] font-bold text-[#1D1D1F]">
                  {check.coverage}%
                </span>
              </div>

              <div className="flex-1 min-w-0">
                <p className="text-[16px] font-bold text-[#1D1D1F] truncate">{check.store.name}</p>
                <div className="flex items-center gap-1.5 mt-1">
                  <MapPin size={13} className="text-[#86868B] shrink-0" />
                  <span className="text-[13px] text-[#86868B] font-medium truncate">{check.store.city}</span>
                </div>
              </div>

              <div className="flex flex-col items-end gap-1.5">
                <span className="text-[12px] text-[#1D1D1F] font-bold">
                  {new Date(check.createdAt).toLocaleDateString('it-IT', {
                    day: 'numeric',
                    month: 'short',
                  })}
                </span>
                <span className="text-[11px] text-[#86868B] font-bold">
                  {check.foundProducts.length}/{check.foundProducts.length + check.missingProducts.length}
                </span>
                {check.status === 'draft' && (
                  <span className="text-[10px] font-bold uppercase tracking-wider text-[#FF9500] bg-[#FFF4E6] px-2.5 py-1 rounded-full border border-[#FF9500]/10">
                    Bozza
                  </span>
                )}
              </div>
              <ChevronRight size={18} className="text-[#D2D2D7] shrink-0" />
            </button>
          ))}
        </div>

        {filtered.length === 0 && (
          <div className="text-center py-20">
            <div className="w-16 h-16 bg-black/5 rounded-full flex items-center justify-center mx-auto mb-4">
              <Search size={24} className="text-[#AEAEB2]" />
            </div>
            <p className="text-[#86868B] font-bold">Nessun check trovato</p>
          </div>
        )}
      </div>
    </PageShell>
  )
}
