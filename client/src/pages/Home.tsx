import { useNavigate } from 'react-router-dom'
import { useAuth } from '../store/auth'
import { mockHistory } from '../lib/mock-data'
import PageShell from '../components/PageShell'
import BottomNav from '../components/BottomNav'
import {
  ScanLine,
  TrendingUp,
  ChevronRight,
  MapPin,
  Clock,
  LayoutGrid,
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
      {/* Header */}
      <div className="px-6 pt-12 pb-6 safe-top flex items-end justify-between">
        <div>
          <p className="text-gray-500 text-[13px] font-bold uppercase tracking-wider">Bentornato,</p>
          <h1 className="text-[32px] font-bold tracking-tight text-gray-900 mt-1">
            {user?.firstName || 'Marco'}
          </h1>
        </div>
        <div className="w-12 h-12 bg-white rounded-2xl shadow-sm border border-gray-100 flex items-center justify-center text-accent">
          <LayoutGrid size={24} />
        </div>
      </div>

      {/* Primary CTA */}
      <div className="px-6 mb-8">
        <button
          onClick={() => navigate('/scan')}
          className="w-full gradient-accent text-white rounded-[32px] p-6 flex items-center gap-5 shadow-xl shadow-indigo-500/25 active:scale-[0.98] transition-all hover:brightness-110"
        >
          <div className="w-14 h-14 bg-white/20 backdrop-blur-md rounded-[20px] flex items-center justify-center shrink-0 border border-white/20">
            <ScanLine size={28} />
          </div>
          <div className="text-left">
            <span className="text-[19px] font-bold block">Nuovo Check</span>
            <span className="text-white/70 text-[14px] font-medium">
              Analizza scaffali e vetrine
            </span>
          </div>
          <div className="ml-auto w-10 h-10 rounded-full bg-white/10 flex items-center justify-center">
            <ChevronRight size={22} className="text-white/80" />
          </div>
        </button>
      </div>

      {/* Stats Bento Grid */}
      <div className="px-6 mb-10">
        <div className="grid grid-cols-2 gap-4">
          <div className="bg-white rounded-[28px] p-5 border border-gray-100 shadow-sm shadow-gray-200/50">
            <div className="w-10 h-10 bg-indigo-50 rounded-xl flex items-center justify-center mb-4">
              <Clock size={20} className="text-accent" />
            </div>
            <p className="text-[28px] font-bold tracking-tight text-gray-900">{totalChecks}</p>
            <p className="text-[13px] text-gray-500 font-bold uppercase tracking-tight">Check mensili</p>
          </div>
          <div className="bg-white rounded-[28px] p-5 border border-gray-100 shadow-sm shadow-gray-200/50">
            <div className="w-10 h-10 bg-emerald-50 rounded-xl flex items-center justify-center mb-4">
              <TrendingUp size={20} className="text-emerald-500" />
            </div>
            <p className="text-[28px] font-bold tracking-tight text-gray-900">{avgCoverage}%</p>
            <p className="text-[13px] text-gray-500 font-bold uppercase tracking-tight">Coverage medio</p>
          </div>
        </div>
      </div>

      {/* Recent Activity */}
      <div className="px-6 pb-32">
        <div className="flex items-center justify-between mb-5">
          <h2 className="text-[17px] font-bold text-gray-900">Attività Recente</h2>
          <button
            onClick={() => navigate('/history')}
            className="text-accent text-[14px] font-bold"
          >
            Vedi tutto
          </button>
        </div>

        <div className="space-y-4">
          {recentChecks.map((check) => (
            <button
              key={check.id}
              onClick={() => navigate(`/history/${check.id}`)}
              className="w-full bg-white rounded-[28px] p-5 flex items-center gap-4 border border-gray-100 shadow-sm shadow-gray-200/50 active:bg-gray-50 active:scale-[0.99] transition-all text-left"
            >
              {/* Modern Coverage Circle */}
              <div className="relative w-14 h-14 shrink-0">
                <svg viewBox="0 0 36 36" className="w-full h-full -rotate-90">
                  <path
                    d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                    fill="none"
                    stroke="#F1F5F9"
                    strokeWidth="3.5"
                  />
                  <path
                    d="M18 2.0845 a 15.9155 15.9155 0 0 1 0 31.831 a 15.9155 15.9155 0 0 1 0 -31.831"
                    fill="none"
                    stroke={check.coverage >= 70 ? '#10B981' : check.coverage >= 40 ? '#F59E0B' : '#EF4444'}
                    strokeWidth="3.5"
                    strokeDasharray={`${check.coverage}, 100`}
                    strokeLinecap="round"
                    className="transition-all duration-1000"
                  />
                </svg>
                <span className="absolute inset-0 flex items-center justify-center text-[12px] font-extrabold text-gray-900">
                  {check.coverage}%
                </span>
              </div>

              <div className="min-w-0 flex-1">
                <p className="text-[16px] font-bold text-gray-900 truncate">
                  {check.store.name}
                </p>
                <div className="flex items-center gap-1.5 mt-1">
                  <MapPin size={13} className="text-gray-400 shrink-0" />
                  <span className="text-[13px] text-gray-500 font-medium truncate">
                    {check.store.city}
                  </span>
                  <span className="text-gray-300 mx-0.5">·</span>
                  <span className="text-[13px] text-gray-500 font-medium">
                    {new Date(check.createdAt).toLocaleDateString('it-IT', {
                      day: 'numeric',
                      month: 'short',
                    })}
                  </span>
                </div>
              </div>

              <div className="flex flex-col items-end gap-1.5">
                <div className="flex -space-x-1.5">
                  <div className="w-6 h-6 rounded-full bg-indigo-50 border-2 border-white flex items-center justify-center text-[10px] font-bold text-accent">
                    {check.foundProducts.length}
                  </div>
                </div>
                {check.status === 'draft' && (
                  <span className="text-[10px] font-bold uppercase tracking-wider text-warning bg-warning-light px-2.5 py-1 rounded-full border border-warning/10">
                    Bozza
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
