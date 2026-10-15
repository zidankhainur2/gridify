import Link from 'next/link';
import { notFound } from 'next/navigation';
import {
  Calendar,
  Users,
  Gamepad2,
  LayoutList,
  GitGraph,
  Target,
  Trophy,
} from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { Participant, StageViewData } from '@/types/api';
import BracketVisualizer from '@/features/bracket/components/bracket-visualizer';
import StandingsTable from '@/features/bracket/components/standings-table';
import BattleRoyaleView from '@/features/bracket/components/br-view';
import RealtimeListener from '@/components/providers/tournament/realtime-listener';
import { Badge, TournamentStatusBadge } from '@/components/ui/badge';
import ShareButton from '@/features/tournaments/components/share-button';
import { GameLogoBadge } from '@/components/ui/game-logo';

export const dynamic = 'force-dynamic';

export default async function PublicTournamentPage({
  params,
  searchParams,
}: {
  params: Promise<{ id: string }>;
  searchParams: Promise<{ stage?: string }>;
}) {
  const { id } = await params;
  const { stage: stageIdParam } = await searchParams;

  let stageViewData: StageViewData;

  try {
    stageViewData = await apiClient.public.getTournament(id, stageIdParam);
  } catch {
    notFound();
  }

  const tournament = stageViewData.tournament;
  if (!tournament) {
    notFound();
  }

  const stages = stageViewData.stages || (stageViewData.stage ? [stageViewData.stage] : []);
  const activeStage =
    stageViewData.stage ||
    (stageIdParam ? stages.find((s) => s.id === stageIdParam) || stages[0] : stages[0]);

  const matches = stageViewData.matches || [];
  const standings = stageViewData.standings || [];
  const champions = stageViewData.champions;

  // Extract participants from matches & standings
  const participantMap = new Map<string, Participant>();
  matches.forEach((m) => {
    if (m.participant_a) participantMap.set(m.participant_a.id, m.participant_a);
    if (m.participant_b) participantMap.set(m.participant_b.id, m.participant_b);
  });
  if (standings) {
    standings.forEach((g) => {
      g.rows?.forEach((r) => {
        if (r.participant) {
          participantMap.set(r.participant.id, r.participant);
        } else if (r.participant_id) {
          participantMap.set(r.participant_id, {
            id: r.participant_id,
            name: r.participant_name || 'Peserta',
            tournament_id: id,
            seed_number: 0,
            is_verified: true,
            created_at: '',
          });
        }
      });
    });
  }
  const participants = Array.from(participantMap.values());
  const hasBracket = matches.length > 0;
  const game = tournament.game || tournament.games;
  const gameName = game?.name || 'Esports';
  const hasKnockoutMatches = stages.some(
    (s) => s.type === 'SINGLE_ELIMINATION' || s.type === 'DOUBLE_ELIMINATION'
  );

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('id-ID', {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
    });
  };

  return (
    <div className="min-h-screen bg-surface-base text-text-primary font-sans selection:bg-accent/20">
      {/* Realtime SSE Listener */}
      <RealtimeListener tournamentId={id} />

      {/* Top Navbar */}
      <header className="border-b border-border-default bg-surface-raised/90 backdrop-blur-md sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 md:px-8 py-3 flex items-center justify-between">
          <Link href="/" className="flex items-center gap-2.5 group">
            <div className="w-8 h-8 rounded-lg bg-surface-elevated border border-border-default flex items-center justify-center text-accent group-hover:border-accent/40 transition-colors">
              <Gamepad2 size={18} />
            </div>
            <span className="font-display font-bold tracking-tight text-text-primary">Gridify</span>
          </Link>

          <div className="flex items-center gap-3">
            <ShareButton tournamentId={id} />
            <Link
              href="/login"
              className="text-xs font-semibold text-text-muted hover:text-text-primary px-3 py-1.5 rounded-lg border border-border-subtle hover:border-border-default bg-surface-elevated/50 transition-colors"
            >
              Masuk Organizer
            </Link>
          </div>
        </div>
      </header>

      {/* Compact Broadcast Header (paruh atas layar terlihat langsung) */}
      <div className="border-b border-border-default bg-surface-raised/60 py-6 md:py-8 px-4 md:px-8">
        <div className="max-w-7xl mx-auto flex flex-col md:flex-row md:items-center justify-between gap-6">
          <div className="space-y-2">
            <div className="flex flex-wrap items-center gap-2.5">
              <GameLogoBadge
                slug={game?.slug}
                gameName={gameName}
                size="sm"
                className="shadow-xs"
              />
              <span className="text-xs font-mono font-medium text-text-muted uppercase tracking-wider">
                {gameName}
              </span>
              <span className="text-border-strong text-xs">•</span>
              <TournamentStatusBadge status={tournament.status} />
            </div>

            <h1 className="text-2xl md:text-3xl lg:text-4xl font-display font-bold tracking-tight text-text-primary">
              {tournament.title}
            </h1>

            {tournament.description && (
              <p className="text-xs md:text-sm text-text-secondary max-w-2xl leading-relaxed">
                {tournament.description}
              </p>
            )}
          </div>

          {/* Quick Stats Capsule */}
          <div className="flex items-center gap-2 self-start md:self-auto flex-wrap text-xs text-text-muted">
            <div className="flex items-center gap-1.5 bg-surface-elevated px-3 py-1.5 rounded-lg border border-border-subtle">
              <Users size={14} className="text-secondary" />
              <span className="font-mono font-medium text-text-primary">{participants.length}</span>
              <span>Peserta</span>
            </div>
            <div className="flex items-center gap-1.5 bg-surface-elevated px-3 py-1.5 rounded-lg border border-border-subtle">
              <Calendar size={14} className="text-amber-400" />
              <span className="font-mono">{formatDate(tournament.created_at)}</span>
            </div>
          </div>
        </div>
      </div>

      {/* Main Bracket Content */}
      <main className="max-w-7xl mx-auto px-4 md:px-8 py-6 md:py-8 space-y-6">
        {/* Stage Selection Navigation */}
        {stages.length > 1 && (
          <div className="flex justify-start md:justify-center overflow-x-auto pb-1">
            <div className="flex bg-surface-raised p-1 rounded-lg border border-border-default gap-1">
              {stages.map((s) => (
                <Link
                  key={s.id}
                  href={`/t/${id}?stage=${s.id}`}
                  className={`px-4 py-2 rounded-md text-xs font-medium transition-colors whitespace-nowrap ${
                    activeStage?.id === s.id
                      ? 'bg-surface-elevated text-text-primary border border-border-subtle shadow-xs font-bold'
                      : 'text-text-muted hover:text-text-primary hover:bg-surface-elevated/40'
                  }`}
                >
                  {s.name}
                </Link>
              ))}
            </div>
          </div>
        )}

        {/* Stage View Canvas */}
        {!activeStage ? (
          <div className="text-center py-16 text-text-muted text-sm bg-surface-raised/40 rounded-xl border border-border-subtle">
            Belum ada jadwal pertandingan yang dirilis.
          </div>
        ) : !hasBracket ? (
          <div className="text-center py-16 rounded-xl bg-surface-raised border border-border-dashed border-border-default space-y-2">
            <Trophy size={28} className="mx-auto text-text-muted opacity-40 mb-2" />
            <h3 className="font-display font-bold text-base text-text-primary">
              Bagan Belum Dirilis
            </h3>
            <p className="text-xs text-text-muted max-w-sm mx-auto">
              Penyelenggara turnamen belum mempublikasikan bagan atau jadwal pertandingan untuk babak ini.
            </p>
          </div>
        ) : (
          <div className="space-y-6">
            <div className="flex items-center gap-2.5 pb-2 border-b border-border-subtle">
              {activeStage.type === 'ROUND_ROBIN' ? (
                <LayoutList size={20} className="text-secondary" />
              ) : activeStage.type === 'LEADERBOARD' ? (
                <Target size={20} className="text-accent" />
              ) : (
                <GitGraph size={20} className="text-accent" />
              )}
              <h2 className="text-base md:text-lg font-display font-bold text-text-primary">
                {activeStage.type === 'ROUND_ROBIN'
                  ? 'Klasemen Babak Grup / Liga'
                  : activeStage.type === 'LEADERBOARD'
                  ? 'Battle Royale Leaderboard'
                  : 'Bagan Pertandingan'}
              </h2>
            </div>

            {/* Visualizer Per Tipe Format */}
            {activeStage.type === 'ROUND_ROBIN' && (
              <StandingsTable
                matches={matches}
                participants={participants}
                tournamentId={id}
                standings={standings}
                isReadOnly={true}
                champions={champions}
                tournament={tournament}
                hasKnockoutMatches={hasKnockoutMatches}
              />
            )}

            {(activeStage.type === 'SINGLE_ELIMINATION' ||
              activeStage.type === 'DOUBLE_ELIMINATION') && (
              <BracketVisualizer
                matches={matches}
                allParticipants={participants}
                tournament={tournament}
                champions={champions}
                isReadOnly={true}
              />
            )}

            {activeStage.type === 'LEADERBOARD' && (
              <BattleRoyaleView
                matches={matches}
                participants={participants}
                tournamentId={id}
                isReadOnly={true}
                brLeaderboard={stageViewData.br_leaderboard}
                champions={champions}
                tournament={tournament}
              />
            )}
          </div>
        )}
      </main>

      {/* Footer */}
      <footer className="border-t border-border-subtle py-8 text-center text-xs font-mono text-text-muted">
        © {new Date().getFullYear()} Gridify. Realtime Esports Tournament Platform.
      </footer>
    </div>
  );
}