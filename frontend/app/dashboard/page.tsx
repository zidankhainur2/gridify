import Link from 'next/link';
import { redirect } from 'next/navigation';
import {
  Plus,
  Trophy,
  Calendar,
  ArrowRight,
  Gamepad2,
  Activity,
  Users,
} from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { DashboardSummary, User } from '@/types/api';
import { Badge } from '@/components/ui/badge';
import { EmptyState } from '@/components/ui/empty-state';
import { DeleteTournamentButton } from '@/features/tournaments/components/delete-tournament-button';

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  });
};

function getStatusBadgeVariant(status: string) {
  switch (status) {
    case 'ONGOING':
      return 'success';
    case 'COMPLETED':
      return 'default';
    case 'DRAFT':
      return 'warning';
    default:
      return 'default';
  }
}

export default async function DashboardPage() {
  let user: User;
  let summary: DashboardSummary;

  try {
    user = await apiClient.users.getMe();
    summary = await apiClient.dashboard.getSummary();
  } catch {
    redirect('/login');
  }

  const userDisplayName = user.first_name || user.username;

  return (
    <div className="space-y-8 animate-in fade-in duration-200">
      {/* Top Header Banner */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl md:text-3xl font-bold tracking-tight text-text-primary">
            Ringkasan Dashboard
          </h1>
          <p className="text-sm text-text-muted mt-1">
            Selamat datang kembali, <span className="text-accent font-medium">{userDisplayName}</span>. Berikut statistik kompetisi Anda.
          </p>
        </div>
        <Link
          href="/dashboard/create"
          className="inline-flex items-center justify-center gap-2 px-5 py-2.5 rounded-lg bg-accent text-white font-medium hover:bg-accent-hover transition-colors shadow-sm self-start md:self-auto text-sm"
        >
          <Plus size={16} />
          Buat Turnamen Baru
        </Link>
      </div>

      {/* Stats Counter Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="p-5 rounded-xl bg-surface-raised border border-border-default flex items-center gap-4">
          <div className="w-12 h-12 rounded-lg bg-surface-elevated border border-border-default flex items-center justify-center shrink-0 text-amber-400">
            <Trophy size={22} />
          </div>
          <div>
            <p className="text-xs font-mono uppercase text-text-muted tracking-wider">
              Total Turnamen
            </p>
            <p className="text-2xl font-bold font-mono text-text-primary mt-0.5">
              {summary.total_tournaments}
            </p>
          </div>
        </div>

        <div className="p-5 rounded-xl bg-surface-raised border border-border-default flex items-center gap-4">
          <div className="w-12 h-12 rounded-lg bg-surface-elevated border border-border-default flex items-center justify-center shrink-0 text-emerald-400">
            <Activity size={22} />
          </div>
          <div>
            <p className="text-xs font-mono uppercase text-text-muted tracking-wider">
              Sedang Berlangsung
            </p>
            <p className="text-2xl font-bold font-mono text-text-primary mt-0.5">
              {summary.active_tournaments ?? (summary as any).ongoing_tournaments ?? 0}
            </p>
          </div>
        </div>

        <div className="p-5 rounded-xl bg-surface-raised border border-border-default flex items-center gap-4">
          <div className="w-12 h-12 rounded-lg bg-surface-elevated border border-border-default flex items-center justify-center shrink-0 text-sky-400">
            <Calendar size={22} />
          </div>
          <div>
            <p className="text-xs font-mono uppercase text-text-muted tracking-wider">
              Pertandingan Mendatang
            </p>
            <p className="text-2xl font-bold font-mono text-text-primary mt-0.5">
              {summary.upcoming_matches_count ?? (summary as any).upcoming_tournaments ?? 0}
            </p>
          </div>
        </div>
      </div>

      {/* Recent Tournaments Section */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Gamepad2 size={18} className="text-accent" />
            <h2 className="text-base font-semibold text-text-primary">
              Turnamen Terbaru
            </h2>
          </div>

          {summary.recent_tournaments && summary.recent_tournaments.length > 0 && (
            <Link
              href="/dashboard/tournaments"
              className="text-xs font-medium text-accent hover:text-accent-hover flex items-center gap-1 transition-colors"
            >
              Lihat Semua Turnamen <ArrowRight size={14} />
            </Link>
          )}
        </div>

        {!summary.recent_tournaments || summary.recent_tournaments.length === 0 ? (
          <EmptyState
            icon={Trophy}
            title="Belum Ada Turnamen"
            description="Buat kompetisi pertama Anda untuk memulai manajemen bracket otomatis dan scoring realtime."
            action={{
              label: 'Buat Turnamen Sekarang',
              href: '/dashboard/create',
            }}
          />
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {summary.recent_tournaments.map((t) => (
              <div
                key={t.id}
                className="group flex flex-col justify-between p-5 rounded-xl bg-surface-raised border border-border-default hover:border-accent/40 transition-colors relative"
              >
                <div>
                  <div className="flex justify-between items-start gap-2 mb-3">
                    <Badge variant={getStatusBadgeVariant(t.status)}>
                      {t.status}
                    </Badge>
                    <DeleteTournamentButton
                      id={t.id}
                      tournamentTitle={t.title}
                      className="opacity-0 group-hover:opacity-100 transition-opacity"
                    />
                  </div>

                  <Link
                    href={`/dashboard/tournaments/${t.id}`}
                    className="block font-bold text-base text-text-primary hover:text-accent transition-colors line-clamp-2 mb-2"
                  >
                    {t.title}
                  </Link>

                  <div className="flex items-center gap-2 text-xs text-text-muted mb-4">
                    <Gamepad2 size={14} className="shrink-0" />
                    <span className="truncate">{t.games?.name || 'Esports'}</span>
                  </div>
                </div>

                <div className="pt-3 border-t border-border-subtle flex items-center justify-between text-xs">
                  <span className="font-mono text-text-muted">
                    {t.format_type.replace(/_/g, ' ')}
                  </span>
                  <Link
                    href={`/dashboard/tournaments/${t.id}`}
                    className="inline-flex items-center gap-1 font-medium text-accent hover:text-accent-hover transition-colors"
                  >
                    Kelola <ArrowRight size={13} />
                  </Link>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}