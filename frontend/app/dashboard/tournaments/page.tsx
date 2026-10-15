import Link from 'next/link';
import { redirect } from 'next/navigation';
import {
  Search,
  Plus,
  Gamepad2,
  Calendar,
  ArrowRight,
  Trophy,
} from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { Tournament } from '@/types/api';
import { Badge, TournamentStatusBadge } from '@/components/ui/badge';
import { EmptyState } from '@/components/ui/empty-state';
import { GameLogoBadge } from '@/components/ui/game-logo';
import { DeleteTournamentButton } from '@/features/tournaments/components/delete-tournament-button';

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  });
};

type SearchParams = Promise<{ q?: string; status?: string }>;

export default async function TournamentsPage(props: {
  searchParams: SearchParams;
}) {
  const searchParams = await props.searchParams;
  const query = searchParams.q || '';
  const statusFilter = searchParams.status || 'ALL';

  let tournaments: Tournament[] = [];

  try {
    const res = await apiClient.tournaments.list({
      q: query || undefined,
      status: statusFilter === 'ALL' ? undefined : statusFilter,
    });
    tournaments = Array.isArray(res) ? res : [];
  } catch {
    redirect('/login');
  }

  return (
    <div className="space-y-6 animate-in fade-in duration-200">
      {/* Header & New Action */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl md:text-3xl font-bold tracking-tight text-text-primary">
            Turnamen Saya
          </h1>
          <p className="text-sm text-text-muted mt-1">
            Kelola turnamen yang Anda selenggarakan, atur jadwal, peserta, dan skor bracket.
          </p>
        </div>
        <Link
          href="/dashboard/create"
          className="inline-flex items-center justify-center gap-2 px-5 py-2.5 rounded-lg bg-accent text-white font-medium hover:bg-accent-hover transition-colors shadow-sm self-start sm:self-auto text-sm"
        >
          <Plus size={16} />
          Buat Turnamen
        </Link>
      </div>

      {/* Filter and Search Bar */}
      <div className="flex flex-col sm:flex-row gap-3 bg-surface-raised p-2 rounded-xl border border-border-default">
        {/* Search Input Form */}
        <form className="relative flex-1 flex items-center gap-2">
          <div className="relative flex-1">
            <Search
              className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted"
              size={18}
            />
            <input
              type="text"
              name="q"
              defaultValue={query}
              placeholder="Cari nama turnamen..."
              className="w-full pl-9 pr-4 py-2 bg-surface-base border border-border-subtle rounded-lg text-sm text-text-primary placeholder:text-text-muted focus:outline-none focus:border-accent transition-colors"
            />
            {statusFilter !== 'ALL' && (
              <input type="hidden" name="status" value={statusFilter} />
            )}
          </div>
          <button
            type="submit"
            className="px-3.5 py-2 bg-surface-elevated hover:bg-surface-overlay text-text-primary border border-border-default rounded-lg text-xs font-semibold shrink-0 cursor-pointer"
          >
            Cari
          </button>
        </form>

        {/* Status Filters */}
        <div className="flex bg-surface-base p-1 rounded-lg border border-border-subtle overflow-x-auto shrink-0">
          {[
            { id: 'ALL', label: 'Semua' },
            { id: 'ONGOING', label: 'Ongoing' },
            { id: 'DRAFT', label: 'Draft' },
            { id: 'COMPLETED', label: 'Selesai' },
          ].map((tab) => {
            const isActive = statusFilter === tab.id;
            return (
              <Link
                key={tab.id}
                href={`/dashboard/tournaments?status=${tab.id}${query ? `&q=${encodeURIComponent(query)}` : ''}`}
                className={`px-3 py-1.5 rounded-md text-xs font-medium transition-colors whitespace-nowrap ${
                  isActive
                    ? 'bg-surface-elevated text-text-primary border border-border-default shadow-xs'
                    : 'text-text-muted hover:text-text-primary'
                }`}
              >
                {tab.label}
              </Link>
            );
          })}
        </div>
      </div>

      {/* Tournament Cards Grid */}
      {(!tournaments || tournaments.length === 0) ? (
        <EmptyState
          icon={Trophy}
          title={query ? 'Turnamen Tidak Ditemukan' : 'Belum Ada Turnamen'}
          description={
            query
              ? `Tidak ditemukan turnamen dengan kata kunci "${query}". Coba ubah kata kunci atau reset filter.`
              : 'Anda belum memiliki turnamen di kategori ini. Mulai buat kompetisi sekarang.'
          }
          action={
            query
              ? {
                  label: 'Reset Pencarian',
                  href: '/dashboard/tournaments',
                }
              : {
                  label: 'Buat Turnamen Baru',
                  href: '/dashboard/create',
                }
          }
        />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {(tournaments || []).map((t) => {
            const game = t.game || t.games;
            return (
              <div
                key={t.id}
                className="group flex flex-col justify-between p-5 rounded-xl bg-surface-raised border border-border-default hover:border-accent/40 transition-colors"
              >
                <div>
                  <div className="flex justify-between items-start gap-2 mb-3">
                    <TournamentStatusBadge status={t.status} />
                    <DeleteTournamentButton
                      id={t.id}
                      tournamentTitle={t.title}
                      className="opacity-100 sm:opacity-0 sm:group-hover:opacity-100 transition-opacity"
                    />
                  </div>

                  <Link
                    href={`/dashboard/tournaments/${t.id}`}
                    className="block font-bold text-base text-text-primary hover:text-accent transition-colors line-clamp-2 mb-2"
                  >
                    {t.title}
                  </Link>

                  <div className="space-y-1.5 text-xs text-text-muted mb-4">
                    <div className="flex items-center gap-2">
                      <GameLogoBadge slug={game?.slug} name={game?.name} genre={game?.genre} size="xs" />
                      <span className="truncate font-medium text-text-secondary">{game?.name || 'Esports'}</span>
                    </div>
                    <div className="flex items-center gap-2 pl-0.5">
                      <Calendar size={14} className="shrink-0" />
                      <span className="font-mono">{formatDate(t.created_at)}</span>
                    </div>
                  </div>
                </div>

                <div className="pt-3 border-t border-border-subtle flex items-center justify-between text-xs">
                  <span className="font-mono text-text-muted uppercase text-[11px]">
                    {t.format_type.replace(/_/g, ' ')}
                  </span>
                  <Link
                    href={`/dashboard/tournaments/${t.id}`}
                    className="inline-flex items-center gap-1 font-semibold text-accent hover:text-accent-hover transition-colors"
                  >
                    Kelola <ArrowRight size={13} />
                  </Link>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}