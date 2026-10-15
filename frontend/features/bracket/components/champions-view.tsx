'use client';

import { Trophy, Medal, Award } from 'lucide-react';
import { Tournament, Match, Participant, ChampionResult, Champions } from '@/types/api';

interface ChampionsViewProps {
  tournament: Tournament;
  matches: Match[];
  participants: Participant[];
  champions?: Champions | ChampionResult[];
  standings?: any[];
}

export default function ChampionsView({
  tournament,
  matches,
  participants,
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
  if (!first && (tournament.format_type === 'SINGLE_ELIMINATION' || tournament.format_type === 'DOUBLE_ELIMINATION')) {
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

  // Jika belum ada pemenang, jangan render podium
  if (!first) return null;

  return (
    <div className="w-full flex flex-col items-center justify-center p-6 md:p-8 rounded-xl bg-surface-raised border border-border-default shadow-sm mb-6 animate-in fade-in duration-300">
      <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-surface-elevated border border-border-default text-xs font-mono text-amber-400 mb-8">
        <Trophy size={14} />
        <span>PODIUM JUARA TURNAMEN</span>
      </div>

      <div className="flex items-end justify-center gap-3 sm:gap-6 md:gap-12 w-full max-w-2xl px-2">
        {/* Juara 2 (Perak) */}
        {second && (
          <div className="flex flex-col items-center flex-1 order-1">
            <div className="w-12 h-12 md:w-14 md:h-14 rounded-full bg-zinc-800 border-2 border-zinc-500 flex items-center justify-center text-zinc-300 shadow-xs mb-3">
              <Medal size={24} />
            </div>
            <div className="text-center w-full">
              <p className="font-bold text-sm md:text-base text-text-primary truncate">
                {second.name}
              </p>
              <span className="text-[11px] font-mono font-medium text-text-muted uppercase">
                Juara 2 • Runner-Up
              </span>
            </div>
            <div className="w-full h-16 md:h-20 bg-surface-elevated rounded-t-lg border-t-2 border-zinc-500/60 mt-3 flex items-center justify-center font-mono font-black text-xl text-zinc-400">
              2
            </div>
          </div>
        )}

        {/* Juara 1 (Emas) - Tengah & Tertinggi */}
        {first && (
          <div className="flex flex-col items-center flex-1 order-2 z-10">
            <div className="w-16 h-16 md:w-20 md:h-20 rounded-full bg-amber-500/10 border-2 border-amber-400 flex items-center justify-center text-amber-400 shadow-md mb-3">
              <Trophy size={32} />
            </div>
            <div className="text-center w-full">
              <p className="font-black text-base md:text-xl text-text-primary truncate">
                {first.name}
              </p>
              <span className="text-xs font-mono font-bold text-amber-400 uppercase tracking-wide">
                Juara 1 • Champion
              </span>
            </div>
            <div className="w-full h-24 md:h-28 bg-surface-elevated rounded-t-lg border-t-2 border-amber-400 mt-3 flex items-center justify-center font-mono font-black text-3xl text-amber-400">
              1
            </div>
          </div>
        )}

        {/* Juara 3 (Perunggu) */}
        {third && (
          <div className="flex flex-col items-center flex-1 order-3">
            <div className="w-12 h-12 md:w-14 md:h-14 rounded-full bg-amber-900/30 border-2 border-amber-700 flex items-center justify-center text-amber-600 shadow-xs mb-3">
              <Award size={24} />
            </div>
            <div className="text-center w-full">
              <p className="font-bold text-sm md:text-base text-text-primary truncate">
                {third.name}
              </p>
              <span className="text-[11px] font-mono font-medium text-text-muted uppercase">
                Juara 3 • Bronze
              </span>
            </div>
            <div className="w-full h-12 md:h-16 bg-surface-elevated rounded-t-lg border-t-2 border-amber-700/60 mt-3 flex items-center justify-center font-mono font-black text-lg text-amber-700">
              3
            </div>
          </div>
        )}
      </div>
    </div>
  );
}