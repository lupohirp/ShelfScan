import { useState } from 'react'
import { mockProducts } from '../../lib/mock-data'
import {
  Search,
  Plus,
  Upload,
  MoreHorizontal,
  Image,
  X,
} from 'lucide-react'

export default function AdminProducts() {
  const [query, setQuery] = useState('')
  const [showAdd, setShowAdd] = useState(false)

  const filtered = mockProducts.filter(
    (p) =>
      p.name.toLowerCase().includes(query.toLowerCase()) ||
      p.sku.toLowerCase().includes(query.toLowerCase()) ||
      p.category.toLowerCase().includes(query.toLowerCase())
  )

  return (
    <div className="max-w-6xl mx-auto animate-fade-in">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Prodotti</h1>
        <div className="flex gap-2">
          <button className="h-9 px-3 bg-white border border-gray-200 rounded-xl text-[13px] font-medium flex items-center gap-1.5 hover:bg-gray-50 transition-colors">
            <Upload size={14} />
            Importa CSV
          </button>
          <button
            onClick={() => setShowAdd(true)}
            className="h-9 px-3 bg-black text-white rounded-xl text-[13px] font-medium flex items-center gap-1.5 hover:bg-gray-800 transition-colors"
          >
            <Plus size={14} />
            Aggiungi
          </button>
        </div>
      </div>

      {/* Search & filters */}
      <div className="bg-white rounded-2xl border border-gray-100 mb-4">
        <div className="p-3 border-b border-gray-100">
          <div className="relative">
            <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Cerca per nome, SKU o categoria..."
              className="w-full h-9 pl-9 pr-4 bg-gray-50 rounded-lg text-[13px] outline-none focus:ring-2 focus:ring-accent/30 placeholder:text-gray-400"
            />
          </div>
        </div>

        {/* Table */}
        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="text-xs text-gray-500 font-medium border-b border-gray-100">
                <th className="px-4 py-3">Prodotto</th>
                <th className="px-4 py-3">SKU</th>
                <th className="px-4 py-3 hidden sm:table-cell">Categoria</th>
                <th className="px-4 py-3 hidden md:table-cell">Immagini</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3 w-10"></th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((product) => (
                <tr key={product.id} className="border-b border-gray-50 hover:bg-gray-50/50 transition-colors">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2.5">
                      <div className="w-9 h-9 bg-gray-100 rounded-lg flex items-center justify-center shrink-0">
                        <Image size={16} className="text-gray-400" />
                      </div>
                      <span className="text-[13px] font-medium truncate max-w-48">{product.name}</span>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-[13px] text-gray-500 font-mono">{product.sku}</td>
                  <td className="px-4 py-3 text-[13px] text-gray-500 hidden sm:table-cell">{product.category}</td>
                  <td className="px-4 py-3 hidden md:table-cell">
                    <span className="text-[12px] text-gray-400">3 foto</span>
                  </td>
                  <td className="px-4 py-3">
                    <span className={`text-[11px] font-medium px-2 py-0.5 rounded-full ${
                      product.status === 'active'
                        ? 'bg-success-light text-success'
                        : 'bg-gray-100 text-gray-500'
                    }`}>
                      {product.status === 'active' ? 'Attivo' : 'Disattivato'}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <button className="text-gray-400 hover:text-gray-600 p-1">
                      <MoreHorizontal size={16} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <div className="px-4 py-3 border-t border-gray-100 text-xs text-gray-500">
          {filtered.length} prodotti
        </div>
      </div>

      {/* Add Product Modal */}
      {showAdd && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-black/30" onClick={() => setShowAdd(false)} />
          <div className="relative bg-white rounded-2xl shadow-xl w-full max-w-md p-6 animate-fade-in">
            <div className="flex items-center justify-between mb-5">
              <h2 className="text-lg font-bold">Nuovo Prodotto</h2>
              <button onClick={() => setShowAdd(false)} className="text-gray-400 hover:text-gray-600 p-1">
                <X size={20} />
              </button>
            </div>

            <div className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1">SKU</label>
                <input className="w-full h-10 px-3 bg-gray-50 rounded-xl text-[14px] outline-none focus:ring-2 focus:ring-accent/30" placeholder="LJ-RING-004" />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1">Nome prodotto</label>
                <input className="w-full h-10 px-3 bg-gray-50 rounded-xl text-[14px] outline-none focus:ring-2 focus:ring-accent/30" placeholder="Nome del prodotto" />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1">Categoria</label>
                <select className="w-full h-10 px-3 bg-gray-50 rounded-xl text-[14px] outline-none focus:ring-2 focus:ring-accent/30">
                  <option>Anelli</option>
                  <option>Collane</option>
                  <option>Bracciali</option>
                  <option>Orecchini</option>
                </select>
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-500 mb-1">Immagini</label>
                <div className="border-2 border-dashed border-gray-200 rounded-xl p-6 text-center">
                  <Upload size={24} className="mx-auto text-gray-300 mb-2" />
                  <p className="text-xs text-gray-400">Trascina o clicca per caricare</p>
                  <p className="text-[10px] text-gray-300 mt-1">JPG, PNG · Max 5 immagini</p>
                </div>
              </div>
            </div>

            <div className="flex gap-3 mt-6">
              <button onClick={() => setShowAdd(false)} className="flex-1 h-10 bg-gray-100 rounded-xl text-[14px] font-medium hover:bg-gray-200 transition-colors">
                Annulla
              </button>
              <button className="flex-1 h-10 bg-black text-white rounded-xl text-[14px] font-medium hover:bg-gray-800 transition-colors">
                Salva
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
