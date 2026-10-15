import Link from 'next/link';
import { notFound } from 'next/navigation';
import {
  Trophy,
  Users,
  Gamepad2,
  Calendar,
  Clock,
  ArrowRight,
  AlertCircle,
  PlayCircle,
} from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { Tournament, Participant, Match } from '@/types/api';
import { Badge, TournamentStatusBadge } from '@/components/ui/badge';
import { GameLogoBadge } from '@/components/ui/game-logo';
import ShareButton from '@/features/tournaments/components/share-button';
import { DeleteTournamentButton } from '@/features/tournaments/components/delete-tournament-button';

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  });
};

export default async function TournamentOverviewPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  let tournament: Tournament;
  let participants: Participant[] = [];
  let matches: Match[] = [];

  try {
    tournament = await apiClient.tournaments.getById(id);
    participants = await apiClient.participants.list(id);

    // Ambil stage pertama jika ada untuk preview match
    if (tournament.stages && tournament.stages.length > 0) {
      const stageData = await apiClient.bracket.getStageView(
        id,
        tournament.stages[0].id
      );
      matches = stageData.matches || [];
    }
  } catch {
    notFound();
  }

  const upcomingMatches = matches
    .filter((m) => m.status !== 'COMPLETED')
    .slice(0, 5);

  const participantMap = new Map<string, string>();
  participants.forEach((p) => participantMap.set(p.id, p.name));

  const game = tournament.game || tournament.games;

  return (
    <div className="space-y-6 animate-in fade-in duration-200 pb-10">
      {/* Overview Hero Banner */}
      <div className="rounded-xl p-6 md:p-8 border border-border-default bg-surface-raised">
        <div className="flex flex-col md:flex-row md:items-start justify-between gap-6">
          <div className="space-y-3">
            <div className="flex flex-wrap items-center gap-2">
              <span className="inline-flex items-center gap-2 px-2.5 py-1 rounded-md bg-surface-elevated text-text-secondary text-xs font-medium border border-border-default">
                <GameLogoBadge slug={game?.slug} name={game?.name} genre={game?.genre} size="xs" />
                <span className="font-semibold text-text-primary">{game?.name || 'Esports Game'}</span>
              </span>
              <TournamentStatusBadge status={tournament.status} />
            </div>

            <h1 className="text-2xl md:text-3xl font-bold tracking-tight text-text-primary">
              {tournament.title}
            </h1>

            <p className="text-text-muted max-w-3xl text-sm leading-relaxed">
              {tournament.description || 'Tidak ada deskripsi tambahan untuk turnamen ini.'}
            </p>
          </div>

          <div className="flex items-center gap-2 self-start md:self-auto shrink-0">
            <ShareButton tournamentId={id} />
            <DeleteTournamentButton
              id={id}
              tournamentTitle={tournament.title}
              redirectTo="/dashboard/tournaments"
              className="bg-surface-elevated border border-border-default"
            />
          </div>
        </div>
      </div>

      {/* Stats Counter Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="p-5 rounded-xl bg-surface-raised border border-border-default flex items-center gap-4">
          <div className="w-12 h-12 rounded-lg bg-surface-elevated border border-border-default flex items-center justify-center shrink-0 text-sky-400">
            <Users size={22} />
          </div>
          <div>
            <p className="text-xs font-mono uppercase text-text-muted tracking-wider">
              Total Peserta
            </p>
            <p className="text-2xl font-bold font-mono text-text-primary mt-0.5">
              {participants.length}
            </p>
          </div>
        </div>

        <div className="p-5 rounded-xl bg-surface-raised border border-border-default flex items-center gap-4">
          <div className="w-12 h-12 rounded-lg bg-surface-elevated border border-border-default flex items-center justify-center shrink-0 text-purple-400">
            <Trophy size={22} />
          </div>
          <div>
            <p className="text-xs font-mono uppercase text-text-muted tracking-wider">
              Format Bagan
            </p>
            <p className="text-sm font-semibold text-text-primary mt-1 font-mono uppercase truncate max-w-48">
              {tournament.format_type.replace(/_/g, ' ')}
            </p>
          </div>
        </div>

        <div className="p-5 rounded-xl bg-surface-raised border border-border-default flex items-center gap-4">
          <div className="w-12 h-12 rounded-lg bg-surface-elevated border border-border-default flex items-center justify-center shrink-0 text-amber-400">
            <Calendar size={22} />
          </div>
          <div>
            <p className="text-xs font-mono uppercase text-text-muted tracking-wider">
              Dibuat Tanggal
            </p>
            <p className="text-sm font-semibold font-mono text-text-primary mt-1">
              {formatDate(tournament.created_at)}
            </p>
          </div>
        </div>
      </div>

      {/* Match Schedule & Quick Actions */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Clock size={16} className="text-accent" />
              <h2 className="text-base font-semibold text-text-primary">
                Jadwal Pertandingan Aktif
              </h2>
            </div>
            <Link
              href={`/dashboard/tournaments/${id}/bracket`}
              className="text-xs font-medium text-accent hover:text-accent-hover flex items-center gap-1 transition-colors"
            >
              Lihat Bagan Lengkap <ArrowRight size={13} />
            </Link>
          </div>

          {upcomingMatches.length === 0 ? (
            <div className="p-8 rounded-xl bg-surface-raised border border-border-dashed border-border-default text-center">
              <AlertCircle size={24} className="mx-auto text-text-muted mb-2 opacity-50" />
              <p className="text-sm font-medium text-text-secondary">
                Belum ada jadwal pertandingan aktif
              </p>
              <p className="text-xs text-text-muted mt-1">
                {tournament.status === 'DRAFT'
                  ? 'Bagan belum digenerate. Daftarkan minimal 2 peserta di tab Peserta lalu mulai turnamen.'
                  : 'Seluruh match telah selesai dipertandingkan.'}
              </p>
            </div>
          ) : (
            <div className="space-y-2.5">
              {upcomingMatches.map((m) => {
                const p1Name = m.participant_a_id
                  ? participantMap.get(m.participant_a_id) || 'Peserta A'
                  : 'TBD';
                const p2Name = m.participant_b_id
                  ? participantMap.get(m.participant_b_id) || 'Peserta B'
                  : 'TBD';

                return (
                  <div
                    key={m.id}
                    className="p-4 rounded-xl bg-surface-raised border border-border-default flex flex-col sm:flex-row items-center justify-between gap-4"
                  >
                    <div className="flex items-center gap-3 w-full sm:w-auto">
                      <span className="text-xs font-mono text-text-muted bg-surface-base px-2 py-1 rounded border border-border-subtle">
                        M#{m.match_number} • R{m.round_number}
                      </span>
                    </div>

                    <div className="flex-1 flex items-center justify-center gap-4 text-sm w-full">
                      <span
                        className={`flex-1 text-right truncate ${
                          m.participant_a_id ? 'font-medium text-text-primary' : 'text-text-muted italic'
                        }`}
                      >
                        {p1Name}
                      </span>
                      <span className="text-xs font-bold text-text-muted px-2 py-0.5 rounded bg-surface-base border border-border-subtle shrink-0">
                        VS
                      </span>
                      <span
                        className={`flex-1 text-left truncate ${
                          m.participant_b_id ? 'font-medium text-text-primary' : 'text-text-muted italic'
                        }`}
                      >
                        {p2Name}
                      </span>
                    </div>

                    <Link
                      href={`/dashboard/tournaments/${id}/bracket`}
                      className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-surface-elevated hover:bg-surface-base border border-border-default text-xs font-medium text-accent transition-colors shrink-0"
                    >
                      <PlayCircle size={14} />
                      <span>Input Skor</span>
                    </Link>
                  </div>
                );
              })}
            </div>
          )}
        </div>

        {/* Action Capsule */}
        <div className="space-y-4">
          <div className="p-5 rounded-xl bg-surface-raised border border-border-default space-y-3">
            <h3 className="font-semibold text-sm text-text-primary">
              Alur Manajemen Turnamen
            </h3>
            <p className="text-xs text-text-muted leading-relaxed">
              Tambahkan peserta kompetisi di tab Peserta. Setelah kuota terpenuhi, bagan bracket akan terbuat secara otomatis saat turnamen dimulai.
            </p>
            <div className="space-y-2 pt-1">
              <Link
                href={`/dashboard/tournaments/${id}/participants`}
                className="block w-full py-2.5 text-center rounded-lg bg-surface-base hover:bg-surface-elevated border border-border-default text-text-primary text-xs font-medium transition-colors"
              >
                Kelola Peserta ({participants.length})
              </Link>
              <Link
                href={`/dashboard/tournaments/${id}/bracket`}
                className="block w-full py-2.5 text-center rounded-lg bg-accent text-white text-xs font-medium hover:bg-accent-hover transition-colors shadow-xs"
              >
                Buka Bagan Turnamen
              </Link>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}