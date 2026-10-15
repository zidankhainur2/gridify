'use client';

import { Trophy, Medal, Award } from 'lucide-react';
import { Tournament, Match, Participant, ChampionResult, Champions } from '@/types/api';

interface ChampionsViewProps {
  tournament?: Tournament;
  matches?: Match[];
  participants?: Participant[];
  champions?: Champions | ChampionResult[];
  standings?: any[];
}

export default function ChampionsView({
  tournament,
  matches = [],
  participants = [],
  champions,
  standings,
}: ChampionsViewProps) {
  let first: Participant | null = null;
  let second: Participant | null = null;
  let third: Participant | null = null;

  // 1. Prioritas utama: Gunakan kalkulasi champions dari backend Go
  if (champions) {
    if (Array.isArray(champions)) {
      champions.forEach((c) => {
        if (c.rank === 1 && c.participant) first = c.participant;
        if (c.rank === 2 && c.participant) second = c.participant;
        if (c.rank === 3 && c.participant) third = c.participant;
      });
    } else {
      if (champions.first) first = champions.first;
      if (champions.second) second = champions.second;
      if (champions.third) third = champions.third;
    }
  }

  // 2. Fallback untuk Elimination: Jika belum ada champions array dari backend
  if (!first && tournament && (tournament.format_type === 'SINGLE_ELIMINATION' || tournament.format_type === 'DOUBLE_ELIMINATION')) {
    const completedMatches = matches.filter((m) => m.status === 'COMPLETED');
    const endMatches = completedMatches
      .filter((m) => !m.next_match_id)
      .sort((a, b) => b.match_number - a.match_number);

    const grandFinal = endMatches[0];
    if (grandFinal && grandFinal.winner_id) {
      first = participants.find((p) => p.id === grandFinal.winner_id) || null;
      const loserId =
        grandFinal.winner_id === grandFinal.participant_a_id
          ? grandFinal.participant_b_id
          : grandFinal.participant_a_id;
      second = participants.find((p) => p.id === loserId) || null;

      // Bronze Match atau Lower Bracket Final
      if (endMatches.length > 1 && endMatches[1].winner_id) {
        third = participants.find((p) => p.id === endMatches[1].winner_id) || null;
      }
    }
  }

  // 3. Fallback untuk Standings / Liga / Battle Royale
  if (!first && standings && standings.length > 0) {
    first = standings[0]?.participant || participants.find((p) => p.id === standings[0]?.participant_id) || null;
    second = standings[1]?.participant || participants.find((p) => p.id === standings[1]?.participant_id) || null;
    third = standings[2]?.participant || participants.find((p) => p.id === standings[2]?.participant_id) || null;
  }

  // Jika belum ada pemenang, jangan render banner podium
  if (!first) return null;

  return (
    <div className="w-full rounded-xl bg-surface-raised border border-border-default overflow-hidden shadow-xs mb-6">
      <div className="px-5 py-3 border-b border-border-subtle bg-surface-elevated/70 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Trophy className="h-4 w-4 text-amber-400" />
          <span className="font-display font-semibold text-xs tracking-wider uppercase text-text-primary">
            Podium Kejuaraan
          </span>
        </div>
        <span className="text-[11px] font-mono font-medium text-emerald-400 bg-emerald-500/10 px-2 py-0.5 rounded border border-emerald-500/20">
          TURNAMEN SELESAI
        </span>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-3 p-4">
        {/* Juara 2 (Runner-Up) */}
        {second ? (
          <div className="flex items-center gap-3.5 p-3.5 rounded-lg bg-surface-base border border-border-default order-2 md:order-1">
            <div className="h-10 w-10 shrink-0 rounded-full bg-slate-800/80 border border-slate-600 flex items-center justify-center text-slate-300 font-bold text-sm">
              <Medal className="h-5 w-5 text-slate-300" />
            </div>
            <div className="min-w-0 flex-1">
              <span className="text-[10px] font-mono font-bold uppercase tracking-wider text-slate-400">
                Juara 2 • Runner-Up
              </span>
              <p className="font-semibold text-sm text-text-primary truncate">
                {second.name}
              </p>
            </div>
          </div>
        ) : (
          <div className="hidden md:block order-2 md:order-1" />
        )}

        {/* Juara 1 (Champion) - High-contrast Gold Highlight */}
        <div className="flex items-center gap-3.5 p-3.5 rounded-lg bg-amber-500/10 border-2 border-amber-400/80 shadow-xs order-1 md:order-2">
          <div className="h-11 w-11 shrink-0 rounded-full bg-amber-400 text-slate-950 flex items-center justify-center font-bold text-base shadow-xs">
            <Trophy className="h-6 w-6" />
          </div>
          <div className="min-w-0 flex-1">
            <span className="text-[10px] font-mono font-bold uppercase tracking-wider text-amber-400">
              Juara 1 • Champion
            </span>
            <p className="font-display font-bold text-base text-text-primary truncate">
              {first.name}
            </p>
          </div>
        </div>

        {/* Juara 3 (Bronze) */}
        {third ? (
          <div className="flex items-center gap-3.5 p-3.5 rounded-lg bg-surface-base border border-border-default order-3">
            <div className="h-10 w-10 shrink-0 rounded-full bg-amber-950/40 border border-amber-800/60 flex items-center justify-center text-amber-500 font-bold text-sm">
              <Award className="h-5 w-5 text-amber-500" />
            </div>
            <div className="min-w-0 flex-1">
              <span className="text-[10px] font-mono font-bold uppercase tracking-wider text-amber-600">
                Juara 3 • Bronze
              </span>
              <p className="font-semibold text-sm text-text-primary truncate">
                {third.name}
              </p>
            </div>
          </div>
        ) : (
          <div className="hidden md:block order-3" />
        )}
      </div>
    </div>
  );
}