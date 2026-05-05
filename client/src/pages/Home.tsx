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
      <div className="px-5 pt-14 pb-6 safe-top">
        <p className="text-gray-500 text-sm">Bentornato,</p>
        <h1 className="text-[28px] font-bold tracking-tight mt-0.5">
          {user?.firstName || 'Marco'}
        </h1>
      </div>

      {/* CTA */}
      <div className="px-5 mb-6">
        <button
          onClick={() => navigate('/scan')}
          className="w-full bg-black text-white rounded-2xl p-5 flex items-center gap-4 active:scale-[0.98] transition-transform"
        >
          <div className="w-12 h-12 bg-white/15 rounded-xl flex items-center justify-center shrink-0">
            <ScanLine size={24} />
          </div>
          <div className="text-left">
            <span className="text-[17px] font-semibold block">Nuovo Check</span>
            <span className="text-white/60 text-sm">
              Scansiona una vetrina
            </span>
          </div>
          <ChevronRight size={20} className="ml-auto text-white/40" />
        </button>
      </div>

      {/* Stats */}
      <div className="px-5 mb-8">
        <div className="grid grid-cols-2 gap-3">
          <div className="bg-gray-50 rounded-2xl p-4">
            <div className="flex items-center gap-2 mb-2">
              <Clock size={16} className="text-gray-400" />
              <span className="text-xs text-gray-500 font-medium">Questo mese</span>
            </div>
            <p className="text-[28px] font-bold tracking-tight">{totalChecks}</p>
            <p className="text-xs text-gray-500">check completati</p>
          </div>
          <div className="bg-gray-50 rounded-2xl p-4">
            <div className="flex items-center gap-2 mb-2">
              <TrendingUp size={16} className="text-success" />
              <span className="text-xs text-gray-500 font-medium">Coverage medio</span>
            </div>
            <p className="text-[28px] font-bold tracking-tight">{avgCoverage}%</p>
            <p className="text-xs text-gray-500">prodotti esposti</p>
          </div>
        </div>
      </div>

      {/* Recent checks */}
      <div className="px-5">
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-[15px] font-semibold text-gray-900">Check recenti</h2>
          <button
            onClick={() => navigate('/history')}
            className="text-accent text-sm font-medium"
          >
            Vedi tutti
          </button>
        </div>

        <div className="space-y-2">
          {recentChecks.map((check) => (
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

              <div className="min-w-0 flex-1">
                <p className="text-[15px] font-semibold truncate">
                  {check.store.name}
                </p>
                <div className="flex items-center gap-1 mt-0.5">
                  <MapPin size={12} className="text-gray-400 shrink-0" />
                  <span className="text-xs text-gray-500 truncate">
                    {check.store.city}
                  </span>
                  <span className="text-xs text-gray-300 mx-1">·</span>
                  <span className="text-xs text-gray-500">
                    {new Date(check.createdAt).toLocaleDateString('it-IT', {
                      day: 'numeric',
                      month: 'short',
                    })}
                  </span>
                </div>
              </div>

              <div className="flex flex-col items-end gap-0.5">
                <span className="text-xs text-gray-500">
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
      </div>

      <BottomNav />
    </PageShell>
  )
}
