import { Search, Plus, MoreHorizontal, Mail } from 'lucide-react'
import { useState } from 'react'

const mockUsers = [
  { id: '1', name: 'Marco Rossi', email: 'marco@liujo.it', role: 'rep', checks: 45, lastActive: '2026-03-15' },
  { id: '2', name: 'Anna Bianchi', email: 'anna@liujo.it', role: 'rep', checks: 32, lastActive: '2026-03-14' },
  { id: '3', name: 'Luigi Verdi', email: 'luigi@liujo.it', role: 'rep', checks: 28, lastActive: '2026-03-12' },
  { id: '4', name: 'Sara Neri', email: 'sara@liujo.it', role: 'admin', checks: 0, lastActive: '2026-03-16' },
  { id: '5', name: 'Paolo Conti', email: 'paolo@liujo.it', role: 'retailer', checks: 12, lastActive: '2026-03-10' },
]

const roleLabels: Record<string, { label: string; className: string }> = {
  admin: { label: 'Amministratore', className: 'bg-accent-light text-accent' },
  rep: { label: 'Rappresentante', className: 'bg-gray-100 text-gray-600' },
  retailer: { label: 'Rivenditore', className: 'bg-warning-light text-warning' },
}

export default function AdminUsers() {
  const [query, setQuery] = useState('')

  const filtered = mockUsers.filter(
    (u) =>
      u.name.toLowerCase().includes(query.toLowerCase()) ||
      u.email.toLowerCase().includes(query.toLowerCase())
  )

  return (
    <div className="max-w-6xl mx-auto animate-fade-in">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Utenti</h1>
        <button className="h-9 px-3 bg-black text-white rounded-xl text-[13px] font-medium flex items-center gap-1.5 hover:bg-gray-800 transition-colors">
          <Plus size={14} />
          Invita utente
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
              placeholder="Cerca utente..."
              className="w-full h-9 pl-9 pr-4 bg-gray-50 rounded-lg text-[13px] outline-none focus:ring-2 focus:ring-accent/30 placeholder:text-gray-400"
            />
          </div>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="text-xs text-gray-500 font-medium border-b border-gray-100">
                <th className="px-4 py-3">Utente</th>
                <th className="px-4 py-3 hidden sm:table-cell">Ruolo</th>
                <th className="px-4 py-3 hidden md:table-cell">Check</th>
                <th className="px-4 py-3 hidden md:table-cell">Ultimo accesso</th>
                <th className="px-4 py-3 w-10"></th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((user) => {
                const role = roleLabels[user.role]
                return (
                  <tr key={user.id} className="border-b border-gray-50 hover:bg-gray-50/50 transition-colors">
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2.5">
                        <div className="w-9 h-9 bg-gray-100 rounded-full flex items-center justify-center shrink-0 text-xs font-bold text-gray-500">
                          {user.name.split(' ').map(n => n[0]).join('')}
                        </div>
                        <div>
                          <p className="text-[13px] font-medium">{user.name}</p>
                          <p className="text-[11px] text-gray-400 flex items-center gap-1">
                            <Mail size={10} />
                            {user.email}
                          </p>
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-3 hidden sm:table-cell">
                      <span className={`text-[11px] font-medium px-2 py-0.5 rounded-full ${role.className}`}>
                        {role.label}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-[13px] text-gray-500 hidden md:table-cell">{user.checks}</td>
                    <td className="px-4 py-3 text-[13px] text-gray-500 hidden md:table-cell">
                      {new Date(user.lastActive).toLocaleDateString('it-IT', { day: 'numeric', month: 'short' })}
                    </td>
                    <td className="px-4 py-3">
                      <button className="text-gray-400 hover:text-gray-600 p-1">
                        <MoreHorizontal size={16} />
                      </button>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>

        <div className="px-4 py-3 border-t border-gray-100 text-xs text-gray-500">
          {filtered.length} utenti
        </div>
      </div>
    </div>
  )
}
