import { useNavigate } from 'react-router-dom'
import { useAuth } from '../store/auth'
import { mockHistory } from '../lib/mock-data'
import PageShell from '../components/PageShell'
import {
  ScanLine,
  TrendingUp,
  ChevronRight,
  MapPin,
  Clock,
  Grid3X3,
} from 'lucide-react'

export default function Home() {
  const user = useAuth((s) => s.user)
  const navigate = useNavigate()
  const recentChecks = mockHistory.slice(0, 3)

  const totalChecks = mockHistory.length
  const avgCoverage = Math.round(
    mockHistory.reduce((a, c) => a + c.coverage, 0) / mockHistory.length
  )

  return (
    <PageShell>
      {/* Header Section */}
      <div className="px-8 pt-16 pb-10 safe-top border-b border-gray-100 flex items-end justify-between bg-white">
        <div>
          <p className="text-gray-400 text-[10px] font-black uppercase tracking-[0.2em] mb-1">Store Associate</p>
          <h1 className="text-[32px] font-black tracking-tight text-black leading-none">
            {user?.firstName?.toUpperCase() || 'MARCO'}
          </h1>
        </div>
        <div className="w-10 h-10 border border-gray-100 flex items-center justify-center text-black">
          <Grid3X3 size={20} />
        </div>
      </div>

      {/* Main CTA */}
      <div className="p-8 border-b border-gray-100 bg-gray-50/50">
        <button
          onClick={() => navigate('/scan')}
          className="w-full bg-black text-white p-8 flex flex-col items-center justify-center gap-4 transition-all active:scale-[0.99]"
        >
          <ScanLine size={32} strokeWidth={1.5} />
          <div className="text-center">
            <span className="text-[16px] font-black uppercase tracking-[0.2em] block mb-1">Nuova Scansione</span>
            <span className="text-white/50 text-[11px] font-bold uppercase tracking-[0.1em]">
              Inventory Visual Check
            </span>
          </div>
        </button>
      </div>

      {/* Stats Bento */}
      <div className="px-8 py-10">
        <h2 className="text-[12px] font-black uppercase tracking-[0.25em] mb-6">Overview Mensile</h2>
        <div className="grid grid-cols-2 gap-4">
          <div className="lj-card p-6 flex flex-col justify-between aspect-square">
            <Clock size={18} className="text-gray-400" />
            <div>
              <p className="text-[32px] font-black tracking-tighter leading-none mb-1">{totalChecks}</p>
              <p className="text-[10px] text-gray-400 font-bold uppercase tracking-wider leading-tight">Check Completati</p>
            </div>
          </div>
          <div className="lj-card p-6 flex flex-col justify-between aspect-square">
            <TrendingUp size={18} className="text-gray-400" />
            <div>
              <p className="text-[32px] font-black tracking-tighter leading-none mb-1">{avgCoverage}%</p>
              <p className="text-[10px] text-gray-400 font-bold uppercase tracking-wider leading-tight">Average Coverage</p>
            </div>
          </div>
        </div>
      </div>

      {/* Recent Activity */}
      <div className="px-8 pb-32">
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-[12px] font-black uppercase tracking-[0.25em]">Attività Recente</h2>
          <button
            onClick={() => navigate('/history')}
            className="text-black text-[11px] font-black uppercase tracking-[0.15em] border-b-2 border-black pb-0.5"
          >
            Vedi tutto
          </button>
        </div>

        <div className="space-y-4">
          {recentChecks.map((check) => (
            <button
              key={check.id}
              onClick={() => navigate(`/history/${check.id}`)}
              className="w-full lj-card p-5 flex items-center gap-5 active:bg-gray-50 transition-all text-left"
            >
              {/* Minimal Coverage Indicator */}
              <div className="relative w-12 h-12 shrink-0 border border-gray-100 flex items-center justify-center font-black text-[13px]">
                {check.coverage}%
              </div>

              <div className="min-w-0 flex-1">
                <p className="text-[14px] font-black uppercase tracking-wider text-black truncate">
                  {check.store.name}
                </p>
                <div className="flex items-center gap-2 mt-1">
                  <MapPin size={11} className="text-gray-400 shrink-0" />
                  <span className="text-[11px] text-gray-400 font-bold uppercase tracking-wide truncate">
                    {check.store.city}
                  </span>
                </div>
              </div>

              <div className="flex flex-col items-end gap-1.5">
                <span className="text-[10px] text-gray-400 font-black uppercase">
                  {new Date(check.createdAt).toLocaleDateString('it-IT', {
                    day: 'numeric',
                    month: 'short',
                  })}
                </span>
                {check.status === 'draft' && (
                  <span className="text-[9px] font-black uppercase tracking-[0.15em] text-white bg-black px-2 py-1">
                    Draft
                  </span>
                )}
              </div>
            </button>
          ))}
        </div>
      </div>
    </PageShell>
  )
}
