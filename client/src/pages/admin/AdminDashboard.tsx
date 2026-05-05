import {
  ClipboardList,
  Users,
  TrendingUp,
  Package,
  ArrowUpRight,
  ArrowDownRight,
} from 'lucide-react'

const kpis = [
  { label: 'Check totali', value: '127', change: '+12%', up: true, icon: ClipboardList },
  { label: 'Utenti attivi', value: '18', change: '+3', up: true, icon: Users },
  { label: 'Coverage medio', value: '64%', change: '+5%', up: true, icon: TrendingUp },
  { label: 'Prodotti catalogo', value: '156', change: '—', up: true, icon: Package },
]

const topExposed = [
  { name: 'Anello Precious Heart', sku: 'LJ-RING-001', exposure: 89 },
  { name: 'Collana Diamond Chain', sku: 'LJ-NECK-001', exposure: 82 },
  { name: 'Orecchini Stud Diamond', sku: 'LJ-EARR-001', exposure: 78 },
  { name: 'Bracciale Tennis Classic', sku: 'LJ-BRAC-001', exposure: 71 },
  { name: 'Anello Infinity Gold', sku: 'LJ-RING-002', exposure: 65 },
]

const leastExposed = [
  { name: 'Bracciale Charm Collection', sku: 'LJ-BRAC-003', exposure: 12 },
  { name: 'Orecchini Drop Pearl', sku: 'LJ-EARR-003', exposure: 18 },
  { name: 'Anello Solitaire', sku: 'LJ-RING-003', exposure: 22 },
  { name: 'Collana Minimal Bar', sku: 'LJ-NECK-003', exposure: 28 },
  { name: 'Bracciale Chain Link', sku: 'LJ-BRAC-002', exposure: 33 },
]

export default function AdminDashboard() {
  return (
    <div className="max-w-6xl mx-auto animate-fade-in">
      <h1 className="text-2xl font-bold mb-6">Dashboard</h1>

      {/* KPIs */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 mb-8">
        {kpis.map(({ label, value, change, up, icon: Icon }) => (
          <div key={label} className="bg-white rounded-2xl p-4 border border-gray-100">
            <div className="flex items-center justify-between mb-3">
              <div className="w-9 h-9 bg-gray-50 rounded-xl flex items-center justify-center">
                <Icon size={18} className="text-gray-600" />
              </div>
              {change !== '—' && (
                <span className={`text-xs font-medium flex items-center gap-0.5 ${up ? 'text-success' : 'text-danger'}`}>
                  {up ? <ArrowUpRight size={12} /> : <ArrowDownRight size={12} />}
                  {change}
                </span>
              )}
            </div>
            <p className="text-2xl font-bold tracking-tight">{value}</p>
            <p className="text-xs text-gray-500 mt-0.5">{label}</p>
          </div>
        ))}
      </div>

      {/* Charts placeholder + tables */}
      <div className="grid lg:grid-cols-2 gap-6">
        {/* Top exposed */}
        <div className="bg-white rounded-2xl border border-gray-100 p-5">
          <h2 className="text-[15px] font-semibold mb-4">Prodotti più esposti</h2>
          <div className="space-y-3">
            {topExposed.map((p, i) => (
              <div key={p.sku} className="flex items-center gap-3">
                <span className="w-6 text-xs text-gray-400 font-medium text-right">{i + 1}</span>
                <div className="flex-1 min-w-0">
                  <p className="text-[13px] font-medium truncate">{p.name}</p>
                  <p className="text-[11px] text-gray-400 font-mono">{p.sku}</p>
                </div>
                <div className="w-20 flex items-center gap-2">
                  <div className="flex-1 h-1.5 bg-gray-100 rounded-full overflow-hidden">
                    <div className="h-full bg-success rounded-full" style={{ width: `${p.exposure}%` }} />
                  </div>
                  <span className="text-xs font-medium text-gray-600 w-8 text-right">{p.exposure}%</span>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Least exposed */}
        <div className="bg-white rounded-2xl border border-gray-100 p-5">
          <h2 className="text-[15px] font-semibold mb-4">Prodotti meno esposti</h2>
          <div className="space-y-3">
            {leastExposed.map((p, i) => (
              <div key={p.sku} className="flex items-center gap-3">
                <span className="w-6 text-xs text-gray-400 font-medium text-right">{i + 1}</span>
                <div className="flex-1 min-w-0">
                  <p className="text-[13px] font-medium truncate">{p.name}</p>
                  <p className="text-[11px] text-gray-400 font-mono">{p.sku}</p>
                </div>
                <div className="w-20 flex items-center gap-2">
                  <div className="flex-1 h-1.5 bg-gray-100 rounded-full overflow-hidden">
                    <div className="h-full bg-danger rounded-full" style={{ width: `${p.exposure}%` }} />
                  </div>
                  <span className="text-xs font-medium text-gray-600 w-8 text-right">{p.exposure}%</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Trend placeholder */}
      <div className="mt-6 bg-white rounded-2xl border border-gray-100 p-5">
        <h2 className="text-[15px] font-semibold mb-4">Trend check mensili</h2>
        <div className="h-48 flex items-end gap-2 px-2">
          {[35, 42, 28, 65, 52, 78, 61, 90, 72, 85, 95, 110].map((v, i) => (
            <div key={i} className="flex-1 flex flex-col items-center gap-1">
              <div
                className="w-full bg-black/80 rounded-t-md transition-all hover:bg-accent"
                style={{ height: `${(v / 120) * 100}%` }}
              />
              <span className="text-[9px] text-gray-400">
                {['G', 'F', 'M', 'A', 'M', 'G', 'L', 'A', 'S', 'O', 'N', 'D'][i]}
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
