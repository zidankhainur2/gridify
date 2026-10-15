'use client';

import { useMemo } from 'react';
import { Trophy, Medal, Skull } from 'lucide-react';
import { Match, Participant, StandingRow } from '@/types/api';

interface BRLeaderboardProps {
  matches: Match[];
  participants: Participant[];
  tournamentId: string;
  standings?: StandingRow[];
  isReadOnly?: boolean;
}

const PLACEMENT_POINTS: Record<number, number> = {
  1: 10,
  2: 6,
  3: 5,
  4: 4,
  5: 3,
  6: 2,
  7: 1,
  8: 1,
};

export default function BRLeaderboard({
  matches,
  participants,
  standings,
}: BRLeaderboardProps) {
  const leaderboard = useMemo(() => {
    // Jika backend Go sudah menyediakan pre-calculated standings
    if (standings && standings.length > 0) {
      return standings.map((s) => ({
        teamId: s.participant_id,
        teamName: s.participant_name,
        matchesPlayed: s.played,
        wwcd: s.won,
        totalKills: s.goals_for,
        totalPoints: s.points,
      }));
    }

    // Fallback: hitung dari match scores
    const stats: Record<
      string,
      {
        teamId: string;
        teamName: string;
        matchesPlayed: number;
        wwcd: number;
        totalKills: number;
        totalPoints: number;
      }
    > = {};

    participants.forEach((p) => {
      stats[p.id] = {
        teamId: p.id,
        teamName: p.name,
        matchesPlayed: 0,
        wwcd: 0,
        totalKills: 0,
        totalPoints: 0,
      };
    });

    matches.forEach((m) => {
      if (m.status !== 'COMPLETED' || !m.scores?.results) return;
      const resList = m.scores.results as any[];

      resList.forEach((r) => {
        const tId = r.team_id || r.teamId;
        const team = stats[tId];
        if (!team) return;

        const rank = typeof r.rank === 'number' ? r.rank : 99;
        const kills = typeof r.kills === 'number' ? r.kills : 0;

        team.matchesPlayed += 1;
        team.totalKills += kills;
        if (rank === 1) team.wwcd += 1;

        const pPts = PLACEMENT_POINTS[rank] || 0;
        team.totalPoints += pPts + kills;
      });
    });

    return Object.values(stats).sort(
      (a, b) =>
        b.totalPoints - a.totalPoints ||
        b.wwcd - a.wwcd ||
        b.totalKills - a.totalKills
    );
  }, [matches, participants, standings]);

  return (
    <div className="rounded-xl bg-surface-raised border border-border-default overflow-hidden shadow-xs">
      <div className="px-6 py-4 border-b border-border-subtle bg-surface-elevated flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Trophy size={18} className="text-amber-400" />
          <h3 className="font-bold text-sm text-text-primary">
            Klasemen Global Battle Royale
          </h3>
        </div>
        <div className="text-[11px] font-mono text-text-muted">
          Skor: Penempatan + Jumlah Kill
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-xs text-left">
          <thead className="bg-surface-base text-text-muted font-mono uppercase text-[10px] tracking-wider border-b border-border-subtle">
            <tr>
              <th className="px-4 py-3 text-center w-12">#</th>
              <th className="px-4 py-3">Nama Tim</th>
              <th className="px-3 py-3 text-center w-16">Match</th>
              <th className="px-3 py-3 text-center w-16 text-amber-400 font-bold">
                WWCD
              </th>
              <th className="px-3 py-3 text-center w-16 text-rose-400 font-bold">
                Kills
              </th>
              <th className="px-4 py-3 text-center w-20 font-bold text-text-primary">
                TOTAL PTS
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border-subtle">
            {leaderboard.map((row, idx) => (
              <tr
                key={row.teamId}
                className={`hover:bg-surface-elevated/40 transition-colors ${
                  idx < 3 ? 'bg-amber-500/5' : ''
                }`}
              >
                <td className="px-4 py-3 text-center font-mono font-medium text-text-muted">
                  {idx === 0 ? (
                    <Medal size={15} className="text-amber-400 mx-auto" />
                  ) : idx === 1 ? (
                    <Medal size={15} className="text-zinc-400 mx-auto" />
                  ) : idx === 2 ? (
                    <Medal size={15} className="text-amber-700 mx-auto" />
                  ) : (
                    idx + 1
                  )}
                </td>
                <td className="px-4 py-3 font-semibold text-text-primary">
                  {row.teamName}
                </td>
                <td className="px-3 py-3 text-center font-mono text-text-secondary">
                  {row.matchesPlayed}
                </td>
                <td className="px-3 py-3 text-center font-mono font-bold text-amber-400">
                  {row.wwcd}
                </td>
                <td className="px-3 py-3 text-center font-mono font-bold text-rose-400">
                  {row.totalKills}
                </td>
                <td className="px-4 py-3 text-center font-mono font-black text-sm text-accent">
                  {row.totalPoints}
                </td>
              </tr>
            ))}

            {leaderboard.length === 0 && (
              <tr>
                <td colSpan={6} className="px-6 py-8 text-center text-text-muted">
                  Belum ada data pertandingan Battle Royale.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
