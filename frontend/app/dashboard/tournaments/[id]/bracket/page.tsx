import Link from 'next/link';
import { notFound } from 'next/navigation';
import { Swords, LayoutList, GitGraph, Target } from 'lucide-react';
import { apiClient } from '@/lib/api-client';
import { Tournament, Stage, Participant, Match, StageViewData } from '@/types/api';
import GenerateBracketButton from './generate-button';
import BracketVisualizer from '@/features/bracket/components/bracket-visualizer';
import StandingsTable from '@/features/bracket/components/standings-table';
import BattleRoyaleView from '@/features/bracket/components/br-view';
import RealtimeListener from '@/components/providers/tournament/realtime-listener';

export const dynamic = 'force-dynamic';

export default async function BracketPage({
  params,
  searchParams,
}: {
  params: Promise<{ id: string }>;
  searchParams: Promise<{ stage?: string }>;
}) {
  const { id } = await params;
  const { stage: stageIdParam } = await searchParams;

  let tournament: Tournament;
  let participants: Participant[] = [];

  try {
    tournament = await apiClient.tournaments.getById(id);
    participants = await apiClient.participants.list(id);
  } catch {
    notFound();
  }

  const stages = tournament.stages || [];
  const activeStage = stageIdParam
    ? stages.find((s) => s.id === stageIdParam) || stages[0]
    : stages[0];

  let stageViewData: StageViewData | null = null;
  if (activeStage) {
    try {
      stageViewData = await apiClient.bracket.getStageView(id, activeStage.id);
    } catch {
      stageViewData = null;
    }
  }

  const matches = stageViewData?.matches || [];
  const standings = stageViewData?.standings || [];
  const champions = stageViewData?.champions || [];
  const hasBracket = matches.length > 0;

  return (
    <div className="space-y-6 h-full flex flex-col animate-in fade-in duration-200">
      {/* Realtime SSE Listener */}
      <RealtimeListener tournamentId={id} />

      {/* Header Stage Navigation & Reset */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 pb-4 border-b border-border-default">
        <div>
          <h2 className="text-xl md:text-2xl font-bold tracking-tight text-text-primary flex items-center gap-2.5">
            {activeStage?.type === 'ROUND_ROBIN' ? (
              <LayoutList size={22} className="text-sky-400" />
            ) : activeStage?.type === 'LEADERBOARD' ? (
              <Target size={22} className="text-rose-400" />
            ) : (
              <GitGraph size={22} className="text-accent" />
            )}
            <span>
              {activeStage?.type === 'ROUND_ROBIN'
                ? 'Klasemen Babak Grup / Liga'
                : activeStage?.type === 'LEADERBOARD'
                ? 'Battle Royale Standings'
                : 'Bagan Bracket Pertandingan'}
            </span>
          </h2>
          <p className="text-xs text-text-muted mt-1 font-mono">
            Format Stage: <span className="text-text-primary uppercase">{activeStage?.type?.replace(/_/g, ' ') || 'Belum Digenerate'}</span>
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2.5">
          {/* Stage Switcher Pill (jika turnamen memiliki multi-stage seperti Hybrid UCL) */}
          {stages.length > 1 && (
            <div className="flex bg-surface-raised p-1 rounded-lg border border-border-default">
              {stages.map((s) => (
                <Link
                  key={s.id}
                  href={`/dashboard/tournaments/${id}/bracket?stage=${s.id}`}
                  className={`px-3 py-1.5 rounded-md text-xs font-medium transition-colors ${
                    activeStage?.id === s.id
                      ? 'bg-surface-elevated text-text-primary border border-border-subtle shadow-xs'
                      : 'text-text-muted hover:text-text-primary'
                  }`}
                >
                  {s.name}
                </Link>
              ))}
            </div>
          )}

          {/* Reset / Re-generate Button (Hanya jika DRAFT atau stage 1) */}
          {tournament.status === 'DRAFT' && (
            <GenerateBracketButton
              tournamentId={id}
              participantCount={participants.length}
              label={hasBracket ? 'Reset Bagan' : 'Generate Bagan'}
              variant={hasBracket ? 'secondary' : 'primary'}
            />
          )}
        </div>
      </div>

      {/* Main Bracket / Stage Area */}
      <div className="flex-1">
        {!hasBracket ? (
          <div className="flex flex-col items-center justify-center py-20 border border-dashed border-border-default rounded-xl bg-surface-raised/40 text-center">
            <div className="w-16 h-16 bg-surface-elevated rounded-xl flex items-center justify-center mb-4 text-text-muted">
              <Swords size={32} />
            </div>
            <h3 className="text-lg font-bold text-text-primary mb-1">
              Bagan Pertandingan Belum Tersusun
            </h3>
            <p className="text-xs text-text-muted max-w-md mb-6 leading-relaxed">
              Terdapat <span className="font-bold text-text-primary">{participants.length}</span> peserta terdaftar. Susun bagan pertandingan secara otomatis berdasarkan format turnamen.
            </p>
            <GenerateBracketButton
              tournamentId={id}
              participantCount={participants.length}
              label="Generate Bagan Pertandingan Sekarang"
              variant="primary"
            />
          </div>
        ) : (
          <>
            {activeStage?.type === 'ROUND_ROBIN' && (
              <StandingsTable
                matches={matches}
                participants={participants}
                tournamentId={id}
                standings={standings}
                isReadOnly={false}
              />
            )}

            {(activeStage?.type === 'SINGLE_ELIMINATION' ||
              activeStage?.type === 'DOUBLE_ELIMINATION') && (
              <BracketVisualizer
                matches={matches}
                allParticipants={participants}
                tournament={tournament}
                champions={champions}
                isReadOnly={false}
              />
            )}

            {activeStage?.type === 'LEADERBOARD' && (
              <BattleRoyaleView
                matches={matches}
                participants={participants}
                tournamentId={id}
                isReadOnly={false}
              />
            )}
          </>
        )}
      </div>
    </div>
  );
}