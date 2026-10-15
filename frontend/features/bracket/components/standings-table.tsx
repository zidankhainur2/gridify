'use client';

import { Swords, Trophy, Medal } from 'lucide-react';
import LeagueMatchCard from './league-match-card';
import AdvanceButton from './advance-button';
import { Match, Participant, StandingRow, GroupStanding } from '@/types/api';

interface StandingsTableProps {
  matches: Match[];
  participants: Participant[];
  tournamentId: string;
  standings?: GroupStanding[] | StandingRow[];
  isReadOnly?: boolean;
}

export default function StandingsTable({
  matches,
  participants,
  tournamentId,
  standings,
  isReadOnly = false,
}: StandingsTableProps) {
  // 1. Grouping standings by group_name jika ada (misal UCL Group A, B, dll)
  const participantMap = new Map<string, Participant>();
  participants.forEach((p) => participantMap.set(p.id, p));

  // Hitung klasemen fallback jika standings array tidak disediakan backend
  const calculatedRows = (() => {
    if (standings && standings.length > 0) {
      if ('rows' in (standings[0] as any)) {
        const flat: StandingRow[] = [];
        (standings as GroupStanding[]).forEach((g) => {
          g.rows.forEach((r) => {
            flat.push({
              ...r,
              group_name: r.group_name || g.group_label,
            });
          });
        });
        return flat;
      }
      return standings as StandingRow[];
    }

    const rowMap = new Map<string, StandingRow>();
    participants.forEach((p) => {
      rowMap.set(p.id, {
        participant: p,
        participant_id: p.id,
        participant_name: p.name,
        group_name: p.group_name || 'Liga Utama',
        played: 0,
        won: 0,
        drawn: 0,
        lost: 0,
        goals_for: 0,
        goals_against: 0,
        goal_difference: 0,
        points: 0,
        rank: 0,
      });
    });

    matches.forEach((m) => {
      if (m.status === 'COMPLETED') {
        const sA = m.participant_a_score ?? m.scores?.a ?? 0;
        const sB = m.participant_b_score ?? m.scores?.b ?? 0;
        const rA = m.participant_a_id ? rowMap.get(m.participant_a_id) : null;
        const rB = m.participant_b_id ? rowMap.get(m.participant_b_id) : null;

        if (rA && rB) {
          rA.played++;
          rA.goals_for += sA;
          rA.goals_against += sB;
          rA.goal_difference = rA.goals_for - rA.goals_against;

          rB.played++;
          rB.goals_for += sB;
          rB.goals_against += sA;
          rB.goal_difference = rB.goals_for - rB.goals_against;

          if (sA > sB) {
            rA.won++;
            rA.points += 3;
            rB.lost++;
          } else if (sB > sA) {
            rB.won++;
            rB.points += 3;
            rA.lost++;
          } else {
            rA.drawn++;
            rA.points += 1;
            rB.drawn++;
            rB.points += 1;
          }
        }
      }
    });

    return Array.from(rowMap.values()).sort((a, b) => {
      if (b.points !== a.points) return b.points - a.points;
      if (b.goal_difference !== a.goal_difference)
        return b.goal_difference - a.goal_difference;
      return b.goals_for - a.goals_for;
    });
  })();

  // Kelompokkan per group_name
  const groupedStandings = new Map<string, StandingRow[]>();
  calculatedRows.forEach((r) => {
    const g = r.group_name || 'Liga Utama';
    if (!groupedStandings.has(g)) groupedStandings.set(g, []);
    groupedStandings.get(g)!.push(r);
  });

  const totalMatches = matches.length;
  const completedMatches = matches.filter((m) => m.status === 'COMPLETED').length;
  const isFinished = totalMatches > 0 && totalMatches === completedMatches;

  return (
    <div className="space-y-8">
      {/* Tables Per Group */}
      {Array.from(groupedStandings.entries()).map(([grpName, rows]) => (
        <div
          key={grpName}
          className="rounded-xl bg-surface-raised border border-border-default overflow-hidden shadow-xs"
        >
          <div className="px-6 py-3.5 border-b border-border-subtle bg-surface-elevated flex items-center justify-between">
            <h3 className="font-bold text-sm text-text-primary flex items-center gap-2">
              <Trophy size={16} className="text-accent" />
              <span>{grpName}</span>
            </h3>
            <span className="text-[11px] font-mono text-text-muted">
              {rows.length} Tim Terdaftar
            </span>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-xs text-left">
              <thead className="bg-surface-base text-text-muted font-mono uppercase text-[10px] tracking-wider border-b border-border-subtle">
                <tr>
                  <th className="px-4 py-3 text-center w-12">#</th>
                  <th className="px-4 py-3">Tim / Pemain</th>
                  <th className="px-3 py-3 text-center w-14">MP</th>
                  <th className="px-3 py-3 text-center w-14 text-emerald-400">W</th>
                  <th className="px-3 py-3 text-center w-14 text-text-muted">D</th>
                  <th className="px-3 py-3 text-center w-14 text-rose-400">L</th>
                  <th className="px-3 py-3 text-center w-16">GD</th>
                  <th className="px-4 py-3 text-center w-18 font-bold text-text-primary">
                    PTS
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border-subtle">
                {rows.map((row, idx) => (
                  <tr
                    key={row.participant_id}
                    className={`hover:bg-surface-elevated/40 transition-colors ${
                      idx < 2 && grpName.includes('Group')
                        ? 'bg-accent/5'
                        : ''
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
                      <div className="flex items-center gap-2">
                        <span>{row.participant_name}</span>
                        {idx < 2 && grpName.includes('Group') && (
                          <span className="text-[9px] font-mono bg-emerald-500/10 text-emerald-400 px-1.5 py-0.5 rounded border border-emerald-500/20 font-bold">
                            Q
                          </span>
                        )}
                      </div>
                    </td>
                    <td className="px-3 py-3 text-center font-mono text-text-secondary">
                      {row.played}
                    </td>
                    <td className="px-3 py-3 text-center font-mono text-emerald-400 font-medium">
                      {row.won}
                    </td>
                    <td className="px-3 py-3 text-center font-mono text-text-muted">
                      {row.drawn}
                    </td>
                    <td className="px-3 py-3 text-center font-mono text-rose-400">
                      {row.lost}
                    </td>
                    <td className="px-3 py-3 text-center font-mono text-text-muted">
                      {row.goal_difference > 0
                        ? `+${row.goal_difference}`
                        : row.goal_difference}
                    </td>
                    <td className="px-4 py-3 text-center font-mono font-bold text-sm text-accent">
                      {row.points}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      ))}

      {/* Match Schedule List */}
      <div className="pt-4 space-y-4">
        <div className="flex items-center gap-2">
          <Swords size={18} className="text-accent" />
          <h3 className="font-semibold text-base text-text-primary">
            Jadwal & Hasil Pertandingan
          </h3>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {matches.map((m) => (
            <LeagueMatchCard
              key={m.id}
              match={m}
              allParticipants={participants}
              isReadOnly={isReadOnly}
            />
          ))}
        </div>

        {matches.length === 0 && (
          <div className="text-center py-8 rounded-xl bg-surface-raised border border-border-dashed border-border-default text-text-muted text-xs">
            Belum ada jadwal pertandingan.
          </div>
        )}
      </div>

      {/* Advance to Knockout Action */}
      {isFinished && !isReadOnly && (
        <AdvanceButton tournamentId={tournamentId} />
      )}
    </div>
  );
}