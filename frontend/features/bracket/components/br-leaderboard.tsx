'use client';

import { useMemo } from 'react';
import { Trophy, Medal, Flame } from 'lucide-react';
import { Match, Participant, StandingRow, BRLeaderboardRow } from '@/types/api';

interface BRLeaderboardProps {
  matches: Match[];
  participants: Participant[];
  tournamentId?: string;
  brLeaderboard?: BRLeaderboardRow[];
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
  brLeaderboard,
  standings,
}: BRLeaderboardProps) {
  const leaderboard = useMemo(() => {
    // 1. Prioritas Utama: Gunakan brLeaderboard yang dihitung oleh backend Go
    if (brLeaderboard && brLeaderboard.length > 0) {
      return brLeaderboard.map((r, idx) => {
        const teamId = r.participant?.id || r.participant_id || `team-${idx}`;
        const teamName = r.participant?.name || r.participant_name || 'Tim';
        return {
          teamId,
          teamName,
          matchesPlayed: r.matches_played,
          wwcd: r.wins,
          placementPoints: r.placement_points,
          killPoints: r.kill_points,
          totalKills: r.total_kills,
          totalPoints: r.total_points,
        };
      });
    }

    // 2. Standings fallback jika tersedia
    if (standings && standings.length > 0) {
      return standings.map((s, idx) => ({
        teamId: s.participant?.id || s.participant_id || `standing-${idx}`,
        teamName: s.participant?.name || s.participant_name || 'Tim',
        matchesPlayed: s.played,
        wwcd: s.won,
        placementPoints: s.placement_points || 0,
        killPoints: s.kill_points || 0,
        totalKills: s.total_kills || s.goals_for,
        totalPoints: s.total_points || s.points,
      }));
    }

    // 3. Fallback kalkulasi lokal dari match scores jika backend belum mengirim leaderboard
    const stats: Record<
      string,
      {
        teamId: string;
        teamName: string;
        matchesPlayed: number;
        wwcd: number;
        placementPoints: number;
        killPoints: number;
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
        placementPoints: 0,
        killPoints: 0,
        totalKills: 0,
        totalPoints: 0,
      };
    });

    matches.forEach((m) => {
      if (m.status !== 'COMPLETED') return;
      const resList = (m.br_results || m.scores?.results || []) as any[];

      resList.forEach((r) => {
        const tId = r.participant_id || r.team_id || r.teamId;
        const team = stats[tId];
        if (!team) return;

        const rank = typeof r.placement_rank === 'number' ? r.placement_rank : typeof r.rank === 'number' ? r.rank : 99;
        const kills = typeof r.kills === 'number' ? r.kills : 0;
        const pPts = typeof r.placement_points === 'number' ? r.placement_points : (PLACEMENT_POINTS[rank] || 0);

        team.matchesPlayed += 1;
        team.totalKills += kills;
        team.placementPoints += pPts;
        team.killPoints += kills;
        if (rank === 1) team.wwcd += 1;

        team.totalPoints += pPts + kills;
      });
    });

    return Object.values(stats).sort(
      (a, b) =>
        b.totalPoints - a.totalPoints ||
        b.wwcd - a.wwcd ||
        b.placementPoints - a.placementPoints ||
        b.totalKills - a.totalKills
    );
  }, [matches, participants, brLeaderboard, standings]);

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
          Skor: Penempatan + Jumlah Eliminasi (Kill)
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
              <th className="px-3 py-3 text-center w-20 text-text-secondary font-medium">
                Placement
              </th>
              <th className="px-3 py-3 text-center w-16 text-rose-400 font-bold">
                Kills
              </th>
              <th className="px-4 py-3 text-center w-24 font-bold text-text-primary">
                TOTAL PTS
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border-subtle">
            {leaderboard.map((row, idx) => (
              <tr
                key={row.teamId}
                className={`hover:bg-surface-elevated/40 transition-colors ${
                  idx === 0
                    ? 'bg-amber-500/10'
                    : idx < 3
                    ? 'bg-surface-elevated/30'
                    : ''
                }`}
              >
                <td className="px-4 py-3 text-center font-mono font-medium text-text-muted">
                  {idx === 0 ? (
                    <Medal size={15} className="text-amber-400 mx-auto" />
                  ) : idx === 1 ? (
                    <Medal size={15} className="text-slate-400 mx-auto" />
                  ) : idx === 2 ? (
                    <Medal size={15} className="text-amber-600 mx-auto" />
                  ) : (
                    idx + 1
                  )}
                </td>
                <td className="px-4 py-3 font-semibold text-text-primary">
                  <div className="flex items-center gap-2">
                    <span>{row.teamName}</span>
                    {idx === 0 && row.totalPoints > 0 && (
                      <span className="text-[9px] font-mono bg-amber-500/15 text-amber-400 px-1.5 py-0.5 rounded border border-amber-500/30 font-bold">
                        LEADER
                      </span>
                    )}
                  </div>
                </td>
                <td className="px-3 py-3 text-center font-mono text-text-secondary">
                  {row.matchesPlayed}
                </td>
                <td className="px-3 py-3 text-center font-mono text-amber-400 font-bold">
                  {row.wwcd}
                </td>
                <td className="px-3 py-3 text-center font-mono text-text-secondary">
                  {row.placementPoints}
                </td>
                <td className="px-3 py-3 text-center font-mono text-rose-400 font-bold">
                  {row.totalKills}
                </td>
                <td className="px-4 py-3 text-center font-mono font-bold text-sm text-accent">
                  {row.totalPoints}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
