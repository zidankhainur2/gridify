import Link from 'next/link';
import { notFound } from 'next/navigation';
import {
  Calendar,
  Trophy,
  Users,
  Gamepad2,
  LayoutList,
  GitGraph,
  Target,
  Share2,
} from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { Tournament, Stage, Participant, Match, StageViewData } from '@/types/api';
import BracketVisualizer from '@/features/bracket/components/bracket-visualizer';
import StandingsTable from '@/features/bracket/components/standings-table';
import BattleRoyaleView from '@/features/bracket/components/br-view';
import RealtimeListener from '@/components/providers/tournament/realtime-listener';
import { Badge } from '@/components/ui/badge';
import ShareButton from '@/features/tournaments/components/share-button';

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
        if (r.participant) participantMap.set(r.participant.id, r.participant);
      });
    });
  }
  const participants = Array.from(participantMap.values());
  const hasBracket = matches.length > 0;

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
        <div className="max-w-7xl mx-auto px-4 md:px-8 py-3.5 flex items-center justify-between">
          <Link href="/" className="flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-lg bg-surface-elevated border border-border-default flex items-center justify-center text-accent">
              <Gamepad2 size={18} />
            </div>
            <span className="font-bold tracking-tight text-text-primary">Gridify</span>
          </Link>

          <div className="flex items-center gap-3">
            <ShareButton tournamentId={id} />
            <Link
              href="/login"
              className="text-xs font-medium text-text-muted hover:text-text-primary px-3 py-1.5 rounded-lg border border-border-subtle hover:border-border-default transition-colors"
            >
              Masuk Organizer
            </Link>
          </div>
        </div>
      </header>

      {/* Hero Tournament Spectator Banner */}
      <div className="border-b border-border-default bg-surface-raised/50 py-12 md:py-16 px-4 md:px-8">
        <div className="max-w-4xl mx-auto text-center space-y-4">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-surface-elevated border border-border-default text-xs font-mono text-accent">
            <Trophy size={14} />
            <span>SPECTATOR LIVE VIEW</span>
          </div>

          <h1 className="text-3xl md:text-5xl font-black tracking-tight text-text-primary">
            {tournament.title}
          </h1>

          {tournament.description && (
            <p className="text-sm md:text-base text-text-muted max-w-2xl mx-auto leading-relaxed">
              {tournament.description}
            </p>
          )}

          {/* Quick Info Capsule */}
          <div className="flex flex-wrap justify-center items-center gap-3 pt-2 text-xs text-text-muted">
            <div className="flex items-center gap-1.5 bg-surface-elevated px-3 py-1.5 rounded-lg border border-border-subtle">
              <Gamepad2 size={14} className="text-accent" />
              <span className="font-medium text-text-primary">
                {tournament.games?.name || 'Esports'}
              </span>
            </div>
            <div className="flex items-center gap-1.5 bg-surface-elevated px-3 py-1.5 rounded-lg border border-border-subtle">
              <Users size={14} className="text-sky-400" />
              <span className="font-mono">{participants.length} Peserta</span>
            </div>
            <div className="flex items-center gap-1.5 bg-surface-elevated px-3 py-1.5 rounded-lg border border-border-subtle">
              <Calendar size={14} className="text-amber-400" />
              <span className="font-mono">{formatDate(tournament.created_at)}</span>
            </div>
            <Badge
              variant={
                tournament.status === 'ONGOING'
                  ? 'success'
                  : tournament.status === 'COMPLETED'
                  ? 'default'
                  : 'warning'
              }
            >
              {tournament.status}
            </Badge>
          </div>
        </div>
      </div>

      {/* Main Bracket Content */}
      <main className="max-w-7xl mx-auto px-4 md:px-8 py-8 md:py-12 space-y-8">
        {/* Stage Selection Navigation */}
        {stages.length > 1 && (
          <div className="flex justify-center">
            <div className="flex bg-surface-raised p-1 rounded-lg border border-border-default overflow-x-auto">
              {stages.map((s) => (
                <Link
                  key={s.id}
                  href={`/t/${id}?stage=${s.id}`}
                  className={`px-4 py-2 rounded-md text-xs font-medium transition-colors whitespace-nowrap ${
                    activeStage?.id === s.id
                      ? 'bg-surface-elevated text-text-primary border border-border-subtle shadow-xs font-bold'
                      : 'text-text-muted hover:text-text-primary'
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
          <div className="text-center py-20 text-text-muted text-sm">
            Belum ada jadwal pertandingan yang dirilis.
          </div>
        ) : !hasBracket ? (
          <div className="text-center py-20 rounded-xl bg-surface-raised border border-border-dashed border-border-default space-y-2">
            <Trophy size={28} className="mx-auto text-text-muted opacity-40 mb-2" />
            <h3 className="font-bold text-base text-text-primary">
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
                <LayoutList size={20} className="text-sky-400" />
              ) : activeStage.type === 'LEADERBOARD' ? (
                <Target size={20} className="text-rose-400" />
              ) : (
                <GitGraph size={20} className="text-accent" />
              )}
              <h2 className="text-lg font-bold text-text-primary">
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
              />
            )}
          </div>
        )}
      </main>

      {/* Footer */}
      <footer className="border-t border-border-subtle py-8 text-center text-xs font-mono text-text-muted">
        © {new Date().getFullYear()} Gridify. Esports Tournament System.
      </footer>
    </div>
  );
}