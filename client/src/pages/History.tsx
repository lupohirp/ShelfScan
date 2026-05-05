import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { mockHistory } from '../lib/mock-data'
import PageShell from '../components/PageShell'
import BottomNav from '../components/BottomNav'
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
      <div className="px-5 pt-14 pb-4 safe-top">
        <h1 className="text-[28px] font-bold tracking-tight">Storico</h1>
      </div>

      <div className="px-5">
        {/* Search */}
        <div className="flex gap-2 mb-4">
          <div className="relative flex-1">
            <Search size={18} className="absolute left-3.5 top-1/2 -translate-y-1/2 text-gray-400" />
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Cerca per negozio..."
              className="w-full h-11 pl-10 pr-4 bg-gray-100 rounded-xl text-[15px] outline-none focus:ring-2 focus:ring-accent/30 transition-shadow placeholder:text-gray-400"
            />
          </div>
          <button className="w-11 h-11 bg-gray-100 rounded-xl flex items-center justify-center text-gray-600">
            <Filter size={18} />
          </button>
        </div>

        {/* List */}
        <div className="space-y-2">
          {filtered.map((check) => (
            <button
              key={check.id}
              onClick={() => navigate(`/history/${check.id}`)}
              className="w-full bg-gray-50 rounded-2xl p-4 flex items-center gap-3 active:bg-gray-100 transition-colors text-left"
            >
              {/* Coverage circle */}
              <div className="relative w-12 h-12 shrink-0">
                <svg viewBox="0 0 36 36" className="w-full h-full -rotate-90">
                  <path
                    d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                    fill="none"
                    stroke="#E8E8ED"
                    strokeWidth="3"
                  />
                  <path
                    d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                    fill="none"
                    stroke={check.coverage >= 70 ? '#34C759' : check.coverage >= 40 ? '#FF9500' : '#FF3B30'}
                    strokeWidth="3"
                    strokeDasharray={`${check.coverage}, 100`}
                    strokeLinecap="round"
                  />
                </svg>
                <span className="absolute inset-0 flex items-center justify-center text-[11px] font-bold">
                  {check.coverage}%
                </span>
              </div>

              <div className="flex-1 min-w-0">
                <p className="text-[15px] font-semibold truncate">{check.store.name}</p>
                <div className="flex items-center gap-1 mt-0.5">
                  <MapPin size={12} className="text-gray-400 shrink-0" />
                  <span className="text-xs text-gray-500 truncate">{check.store.city}</span>
                </div>
              </div>

              <div className="flex flex-col items-end gap-1">
                <span className="text-xs text-gray-500">
                  {new Date(check.createdAt).toLocaleDateString('it-IT', {
                    day: 'numeric',
                    month: 'short',
                  })}
                </span>
                <span className="text-xs text-gray-400">
                  {check.foundProducts.length}/{check.foundProducts.length + check.missingProducts.length}
                </span>
                {check.status === 'draft' && (
                  <span className="text-[10px] font-medium text-warning bg-warning-light px-2 py-0.5 rounded-full">
                    Bozza
                  </span>
                )}
              </div>
              <ChevronRight size={16} className="text-gray-300 shrink-0" />
            </button>
          ))}
        </div>

        {filtered.length === 0 && (
          <div className="text-center py-16">
            <p className="text-gray-400 text-sm">Nessun check trovato</p>
          </div>
        )}
      </div>

      <BottomNav />
    </PageShell>
  )
}
