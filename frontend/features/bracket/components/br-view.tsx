'use client';

import { useMemo } from 'react';
import { Match, Participant, BRLeaderboardRow, Tournament, Champions, ChampionResult } from '@/types/api';
import BRGroupList from './br-group-list';
import BRLeaderboard from './br-leaderboard';
import BRMatchInput from './br-match-input';
import ChampionsView from './champions-view';
import {
  transformParticipantsToGroups,
  isMultiGroupSystem,
} from '../utils/br-helpers';
import { Target } from 'lucide-react';

interface Props {
  matches: Match[];
  participants: Participant[];
  tournamentId: string;
  brLeaderboard?: BRLeaderboardRow[];
  champions?: Champions | ChampionResult[];
  tournament?: Tournament;
  isReadOnly?: boolean;
}

export default function BattleRoyaleView({
  matches,
  participants,
  tournamentId,
  brLeaderboard,
  champions,
  tournament,
  isReadOnly = false,
}: Props) {
  const groups = useMemo(
    () => transformParticipantsToGroups(participants),
    [participants]
  );
  const showGrouping = isMultiGroupSystem(groups);

  const totalMatches = matches.length;
  const completedMatches = matches.filter((m) => m.status === 'COMPLETED').length;
  const isFinished = totalMatches > 0 && totalMatches === completedMatches;

  return (
    <div className="space-y-8 animate-in fade-in duration-300">
      {/* 0. Champions Banner (Jika Seluruh Round Selesai) */}
      {champions && isFinished && (
        <ChampionsView
          tournament={tournament}
          matches={matches}
          participants={participants}
          champions={champions}
        />
      )}

      {/* 1. Group Distribution (Jika Multi-Group) */}
      {showGrouping && (
        <section>
          <BRGroupList groups={groups} />
        </section>
      )}

      {/* 2. Global Leaderboard */}
      <section>
        <BRLeaderboard
          matches={matches}
          participants={participants}
          tournamentId={tournamentId}
          brLeaderboard={brLeaderboard}
          isReadOnly={isReadOnly}
        />
      </section>

      {/* 3. Match Rounds List */}
      <section className="space-y-4">
        <div className="flex items-center gap-2">
          <Target size={18} className="text-accent" />
          <h3 className="font-semibold text-base text-text-primary">
            Jadwal Putaran & Input Hasil
          </h3>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {matches.map((match) => (
            <BRMatchInput
              key={match.id}
              match={match}
              participants={participants}
              tournamentId={tournamentId}
              isReadOnly={isReadOnly}
            />
          ))}
        </div>

        {matches.length === 0 && (
          <div className="text-center py-12 border border-dashed border-border-default rounded-xl bg-surface-raised text-text-muted text-xs">
            Belum ada jadwal pertandingan Battle Royale.
          </div>
        )}
      </section>
    </div>
  );
}
