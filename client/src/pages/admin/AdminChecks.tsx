import { useState } from 'react'
import { mockHistory } from '../../lib/mock-data'
import { Search, Filter, Download, MapPin, User } from 'lucide-react'

export default function AdminChecks() {
  const [query, setQuery] = useState('')

  const filtered = mockHistory.filter(
    (c) =>
      c.store.name.toLowerCase().includes(query.toLowerCase()) ||
      c.store.city.toLowerCase().includes(query.toLowerCase())
  )

  return (
    <div className="max-w-6xl mx-auto animate-fade-in">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Check History</h1>
        <button className="h-9 px-3 bg-white border border-gray-200 rounded-xl text-[13px] font-medium flex items-center gap-1.5 hover:bg-gray-50 transition-colors">
          <Download size={14} />
          Esporta CSV
        </button>
      </div>

      <div className="bg-white rounded-2xl border border-gray-100">
        <div className="p-3 border-b border-gray-100 flex gap-2">
          <div className="relative flex-1">
            <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Cerca per negozio..."
              className="w-full h-9 pl-9 pr-4 bg-gray-50 rounded-lg text-[13px] outline-none focus:ring-2 focus:ring-accent/30 placeholder:text-gray-400"
            />
          </div>
          <button className="h-9 px-3 bg-gray-50 rounded-lg text-[13px] text-gray-600 flex items-center gap-1.5">
            <Filter size={14} />
            Filtri
          </button>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="text-xs text-gray-500 font-medium border-b border-gray-100">
                <th className="px-4 py-3">Negozio</th>
                <th className="px-4 py-3 hidden sm:table-cell">Utente</th>
                <th className="px-4 py-3">Coverage</th>
                <th className="px-4 py-3 hidden md:table-cell">Prodotti</th>
                <th className="px-4 py-3">Data</th>
                <th className="px-4 py-3">Status</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((check) => (
                <tr key={check.id} className="border-b border-gray-50 hover:bg-gray-50/50 transition-colors cursor-pointer">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <MapPin size={14} className="text-gray-400 shrink-0" />
                      <div>
                        <p className="text-[13px] font-medium">{check.store.name}</p>
                        <p className="text-[11px] text-gray-400">{check.store.city}</p>
                      </div>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-[13px] text-gray-500 hidden sm:table-cell">
                    <span className="flex items-center gap-1">
                      <User size={12} />
                      Marco R.
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <div className="w-12 h-1.5 bg-gray-100 rounded-full overflow-hidden">
                        <div
                          className="h-full rounded-full"
                          style={{
                            width: `${check.coverage}%`,
                            backgroundColor: check.coverage >= 70 ? '#34C759' : check.coverage >= 40 ? '#FF9500' : '#FF3B30',
                          }}
                        />
                      </div>
                      <span className="text-[13px] font-medium">{check.coverage}%</span>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-[13px] text-gray-500 hidden md:table-cell">
                    {check.foundProducts.length}/{check.foundProducts.length + check.missingProducts.length}
                  </td>
                  <td className="px-4 py-3 text-[13px] text-gray-500">
                    {new Date(check.createdAt).toLocaleDateString('it-IT', { day: 'numeric', month: 'short', year: 'numeric' })}
                  </td>
                  <td className="px-4 py-3">
                    <span className={`text-[11px] font-medium px-2 py-0.5 rounded-full ${
                      check.status === 'finalized'
                        ? 'bg-success-light text-success'
                        : 'bg-warning-light text-warning'
                    }`}>
                      {check.status === 'finalized' ? 'Completato' : 'Bozza'}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <div className="px-4 py-3 border-t border-gray-100 text-xs text-gray-500">
          {filtered.length} check
        </div>
      </div>
    </div>
  )
}
