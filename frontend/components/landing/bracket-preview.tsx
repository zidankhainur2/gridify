'use client';

import * as React from 'react';
import { Trophy, CheckCircle2 } from 'lucide-react';

export function BracketPreview() {
  const [selectedWinner, setSelectedWinner] = React.useState<'teamA' | 'teamB' | null>('teamA');

  return (
    <div className="w-full max-w-4xl mx-auto mt-12 p-6 rounded-2xl bg-[#121215] border border-[#27272A] shadow-2xl relative overflow-hidden">
      {/* Top Header bar */}
      <div className="flex items-center justify-between pb-4 mb-6 border-b border-[#27272A]/70 text-xs text-zinc-400">
        <div className="flex items-center gap-2">
          <span className="flex h-2 w-2 rounded-full bg-emerald-500 animate-pulse" />
          <span className="font-semibold text-zinc-200">Pratinjau Interaktif Bagan Turnamen</span>
          <span className="px-2 py-0.5 rounded bg-zinc-800 text-[11px] text-zinc-300 font-mono">LIVE SSE</span>
        </div>
        <div className="hidden sm:flex items-center gap-3 font-mono">
          <span>8 Tim</span>
          <span>•</span>
          <span>Best of 3</span>
          <span>•</span>
          <span className="text-emerald-400">Grand Final</span>
        </div>
      </div>

      {/* Bracket Tree Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 items-center">
        {/* Semifinals Column */}
        <div className="space-y-4">
          <div className="text-[11px] uppercase tracking-wider font-semibold text-zinc-500 px-1">
            Semifinals (Selesai)
          </div>

          {/* Match 1 */}
          <div className="p-3 rounded-lg bg-[#18181B] border border-[#27272A] text-xs space-y-1.5 shadow-sm">
            <div className="flex justify-between items-center text-zinc-100 font-medium">
              <span className="flex items-center gap-1.5">
                <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400 shrink-0" />
                Fnatic
              </span>
              <span className="font-mono font-bold text-emerald-400">2</span>
            </div>
            <div className="flex justify-between items-center text-zinc-500">
              <span>Sentinels</span>
              <span className="font-mono">0</span>
            </div>
          </div>

          {/* Match 2 */}
          <div className="p-3 rounded-lg bg-[#18181B] border border-[#27272A] text-xs space-y-1.5 shadow-sm">
            <div className="flex justify-between items-center text-zinc-100 font-medium">
              <span className="flex items-center gap-1.5">
                <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400 shrink-0" />
                Paper Rex
              </span>
              <span className="font-mono font-bold text-emerald-400">2</span>
            </div>
            <div className="flex justify-between items-center text-zinc-500">
              <span>DRX</span>
              <span className="font-mono">1</span>
            </div>
          </div>
        </div>

        {/* Grand Final Column */}
        <div className="md:col-span-2 space-y-4">
          <div className="text-[11px] uppercase tracking-wider font-semibold text-zinc-500 px-1 flex items-center justify-between">
            <span>Grand Final (Klik Pemenang)</span>
            <span className="text-zinc-400 text-[10px]">Coba Interaksi ↓</span>
          </div>

          <div className="p-4 rounded-xl bg-[#18181B] border-2 border-[#3F3F46] text-xs space-y-3 shadow-md">
            <div
              onClick={() => setSelectedWinner('teamA')}
              className={`flex justify-between items-center p-2.5 rounded-lg cursor-pointer transition-all ${
                selectedWinner === 'teamA'
                  ? 'bg-emerald-950/40 border border-emerald-600/50 text-white font-semibold'
                  : 'bg-[#121215] text-zinc-300 hover:bg-zinc-800/60'
              }`}
            >
              <div className="flex items-center gap-2">
                <span className="h-5 w-5 rounded bg-zinc-800 text-zinc-200 flex items-center justify-center font-mono font-bold text-[10px]">
                  #1
                </span>
                <span>Fnatic</span>
                {selectedWinner === 'teamA' && (
                  <span className="text-[10px] px-1.5 py-0.5 rounded bg-emerald-500/20 text-emerald-400">Juara 1</span>
                )}
              </div>
              <span className="font-mono text-sm font-bold text-emerald-400">
                {selectedWinner === 'teamA' ? '3' : '1'}
              </span>
            </div>

            <div
              onClick={() => setSelectedWinner('teamB')}
              className={`flex justify-between items-center p-2.5 rounded-lg cursor-pointer transition-all ${
                selectedWinner === 'teamB'
                  ? 'bg-emerald-950/40 border border-emerald-600/50 text-white font-semibold'
                  : 'bg-[#121215] text-zinc-300 hover:bg-zinc-800/60'
              }`}
            >
              <div className="flex items-center gap-2">
                <span className="h-5 w-5 rounded bg-zinc-800 text-zinc-200 flex items-center justify-center font-mono font-bold text-[10px]">
                  #2
                </span>
                <span>Paper Rex</span>
                {selectedWinner === 'teamB' && (
                  <span className="text-[10px] px-1.5 py-0.5 rounded bg-emerald-500/20 text-emerald-400">Juara 1</span>
                )}
              </div>
              <span className="font-mono text-sm font-bold text-emerald-400">
                {selectedWinner === 'teamB' ? '3' : '2'}
              </span>
            </div>

            {/* Winner Trophy banner */}
            <div className="pt-2 flex items-center justify-center gap-2 text-xs font-medium text-amber-400 border-t border-[#27272A]">
              <Trophy className="h-4 w-4" />
              <span>
                Pemenang:{' '}
                <strong className="text-zinc-100">
                  {selectedWinner === 'teamA' ? 'Fnatic' : 'Paper Rex'}
                </strong>
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
