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
          <p className="text-[#86868B] text-[13px] font-bold uppercase tracking-[0.1em]">Bentornato,</p>
          <h1 className="text-[36px] font-bold tracking-tight text-[#1D1D1F] mt-1">
            {user?.firstName || 'Marco'}
          </h1>
        </div>
        <div className="w-12 h-12 glass-card rounded-2xl flex items-center justify-center text-[#B4894D] border-white/60">
          <LayoutGrid size={24} />
        </div>
      </div>

      {/* Primary CTA */}
      <div className="px-6 mb-8">
        <button
          onClick={() => navigate('/scan')}
          className="w-full gradient-accent text-white rounded-[32px] p-6 flex items-center gap-5 shadow-xl shadow-[#B4894D]/25 active:scale-[0.98] transition-all hover:brightness-110"
        >
          <div className="w-14 h-14 bg-white/20 backdrop-blur-md rounded-[20px] flex items-center justify-center shrink-0 border border-white/30">
            <ScanLine size={28} />
          </div>
          <div className="text-left">
            <span className="text-[20px] font-bold block leading-tight">Nuovo Check</span>
            <span className="text-white/80 text-[14px] font-medium tracking-wide uppercase">
              Analisi Boutique
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
          <div className="glass-card rounded-[28px] p-5 border-white/60">
            <div className="w-10 h-10 bg-[#F8F2EA] rounded-xl flex items-center justify-center mb-4 border border-[#B4894D]/10">
              <Clock size={20} className="text-[#B4894D]" />
            </div>
            <p className="text-[28px] font-bold tracking-tight text-[#1D1D1F]">{totalChecks}</p>
            <p className="text-[12px] text-[#86868B] font-bold uppercase tracking-wider">Check mensili</p>
          </div>
          <div className="glass-card rounded-[28px] p-5 border-white/60">
            <div className="w-10 h-10 bg-[#E8FAF0] rounded-xl flex items-center justify-center mb-4 border border-[#30D158]/10">
              <TrendingUp size={20} className="text-[#30D158]" />
            </div>
            <p className="text-[28px] font-bold tracking-tight text-[#1D1D1F]">{avgCoverage}%</p>
            <p className="text-[12px] text-[#86868B] font-bold uppercase tracking-wider">Coverage medio</p>
          </div>
        </div>
      </div>

      {/* Recent Activity */}
      <div className="px-6 pb-32">
        <div className="flex items-center justify-between mb-5 px-1">
          <h2 className="text-[18px] font-bold text-[#1D1D1F]">Attività Recente</h2>
          <button
            onClick={() => navigate('/history')}
            className="text-[#0071E3] text-[14px] font-bold"
          >
            Vedi tutto
          </button>
        </div>

        <div className="space-y-4">
          {recentChecks.map((check) => (
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

              <div className="min-w-0 flex-1">
                <p className="text-[16px] font-bold text-[#1D1D1F] truncate">
                  {check.store.name}
                </p>
                <div className="flex items-center gap-1.5 mt-1">
                  <MapPin size={13} className="text-[#86868B] shrink-0" />
                  <span className="text-[13px] text-[#86868B] font-medium truncate">
                    {check.store.city}
                  </span>
                  <span className="text-[#D2D2D7] mx-0.5">·</span>
                  <span className="text-[13px] text-[#86868B] font-medium">
                    {new Date(check.createdAt).toLocaleDateString('it-IT', {
                      day: 'numeric',
                      month: 'short',
                    })}
                  </span>
                </div>
              </div>

              <div className="flex flex-col items-end gap-1.5">
                <div className="w-7 h-7 rounded-full bg-[#B4894D]/10 border border-[#B4894D]/20 flex items-center justify-center text-[11px] font-bold text-[#B4894D]">
                  {check.foundProducts.length}
                </div>
                {check.status === 'draft' && (
                  <span className="text-[10px] font-bold uppercase tracking-wider text-[#FF9500] bg-[#FFF4E6] px-2.5 py-1 rounded-full border border-[#FF9500]/10">
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
