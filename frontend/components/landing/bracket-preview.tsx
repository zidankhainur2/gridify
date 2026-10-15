'use client';

import * as React from 'react';
import { Trophy, CheckCircle2 } from 'lucide-react';
import { GameMeta, SUPPORTED_GAMES_META } from '@/components/ui/game-logo';

interface BracketPreviewProps {
  selectedGame?: GameMeta;
}

// Sample match data based on genre / game
const GAME_PREVIEWS: Record<
  string,
  {
    title: string;
    semi1: { t1: string; s1: number; t2: string; s2: number; winner: string };
    semi2: { t1: string; s1: number; t2: string; s2: number; winner: string };
    final: { t1: string; seed1: string; t2: string; seed2: string };
  }
> = {
  valorant: {
    title: 'VCT Challengers Final',
    semi1: { t1: 'Fnatic', s1: 2, t2: 'Sentinels', s2: 0, winner: 'Fnatic' },
    semi2: { t1: 'Paper Rex', s1: 2, t2: 'DRX', s2: 1, winner: 'Paper Rex' },
    final: { t1: 'Fnatic', seed1: '#1', t2: 'Paper Rex', seed2: '#2' },
  },
  'mobile-legends': {
    title: 'MPL Invitational Playoffs',
    semi1: { t1: 'ONIC Esports', s1: 2, t2: 'RRQ Hoshi', s2: 1, winner: 'ONIC Esports' },
    semi2: { t1: 'Echo PH', s1: 2, t2: 'Blacklist', s2: 0, winner: 'Echo PH' },
    final: { t1: 'ONIC Esports', seed1: '#1', t2: 'Echo PH', seed2: '#2' },
  },
  'dota-2': {
    title: 'The International Qualifier',
    semi1: { t1: 'Team Spirit', s1: 2, t2: 'Gaimin Gladiators', s2: 0, winner: 'Team Spirit' },
    semi2: { t1: 'Team Liquid', s1: 2, t2: 'BetBoom Team', s2: 1, winner: 'Team Liquid' },
    final: { t1: 'Team Spirit', seed1: '#1', t2: 'Team Liquid', seed2: '#2' },
  },
  'counter-strike-2': {
    title: 'Major Championship Finals',
    semi1: { t1: 'FaZe Clan', s1: 2, t2: 'Vitality', s2: 1, winner: 'FaZe Clan' },
    semi2: { t1: 'Natus Vincere', s1: 2, t2: 'G2 Esports', s2: 0, winner: 'Natus Vincere' },
    final: { t1: 'FaZe Clan', seed1: '#1', t2: 'Natus Vincere', seed2: '#2' },
  },
  'ea-sports-fc': {
    title: 'eChampions League Grand Final',
    semi1: { t1: 'Manchester City', s1: 3, t2: 'Real Madrid', s2: 1, winner: 'Manchester City' },
    semi2: { t1: 'Bayern Munich', s1: 2, t2: 'PSG', s2: 0, winner: 'Bayern Munich' },
    final: { t1: 'Manchester City', seed1: '#1', t2: 'Bayern Munich', seed2: '#2' },
  },
  'pubg-mobile': {
    title: 'PMSL Super League Grand Finals',
    semi1: { t1: 'Bigetron RA', s1: 2, t2: 'Alter Ego', s2: 1, winner: 'Bigetron RA' },
    semi2: { t1: 'Vampire Esports', s1: 2, t2: 'BOOM Esports', s2: 0, winner: 'Vampire Esports' },
    final: { t1: 'Bigetron RA', seed1: '#1', t2: 'Vampire Esports', seed2: '#2' },
  },
};

export function BracketPreview({ selectedGame }: BracketPreviewProps) {
  const [selectedWinner, setSelectedWinner] = React.useState<'teamA' | 'teamB'>('teamA');

  const slug = selectedGame?.slug || 'valorant';
  const previewData = GAME_PREVIEWS[slug] || GAME_PREVIEWS['valorant'];

  // Reset or adjust winner when game changes
  React.useEffect(() => {
    setSelectedWinner('teamA');
  }, [slug]);

  return (
    <div className="w-full max-w-4xl mx-auto mt-10 p-5 md:p-7 rounded-2xl bg-surface-raised border border-border-default shadow-2xl relative overflow-hidden text-left">
      {/* Top Header bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between pb-4 mb-6 border-b border-border-subtle gap-2 text-xs text-text-muted">
        <div className="flex items-center gap-2.5">
          <span className="flex h-2.5 w-2.5 rounded-full bg-emerald-500 animate-pulse" />
          <span className="font-semibold text-text-primary">
            {previewData.title}
          </span>
          <span className="px-2 py-0.5 rounded bg-surface-elevated text-[10px] text-text-secondary font-mono border border-border-subtle">
            LIVE PREVIEW
          </span>
        </div>
        <div className="flex items-center gap-3 font-mono text-[11px] text-text-muted">
          <span>8 Tim</span>
          <span>•</span>
          <span>Single Elimination</span>
          <span>•</span>
          <span className="text-emerald-400 font-semibold">Grand Final</span>
        </div>
      </div>

      {/* Bracket Tree Grid with SVG Connectors */}
      <div className="relative">
        {/* SVG Connector Overlay for Desktop */}
        <svg
          className="hidden md:block absolute inset-0 pointer-events-none z-10 w-full h-full"
          preserveAspectRatio="none"
          viewBox="0 0 800 240"
        >
          {/* Connector Semifinal 1 -> Grand Final Slot A */}
          <path
            d="M 260 55 H 330 V 85 H 375"
            fill="none"
            stroke="var(--color-success)"
            strokeWidth="2"
            strokeLinecap="round"
          />
          <circle cx="375" cy="85" r="3" fill="var(--color-success)" />

          {/* Connector Semifinal 2 -> Grand Final Slot B */}
          <path
            d="M 260 175 H 330 V 145 H 375"
            fill="none"
            stroke="var(--color-success)"
            strokeWidth="2"
            strokeLinecap="round"
          />
          <circle cx="375" cy="145" r="3" fill="var(--color-success)" />
        </svg>

        <div className="grid grid-cols-1 md:grid-cols-12 gap-6 items-center relative z-20">
          {/* Semifinals Column (4 cols) */}
          <div className="md:col-span-5 space-y-4">
            <div className="text-[11px] uppercase tracking-wider font-mono font-semibold text-text-muted px-1">
              Semifinals (Selesai)
            </div>

            {/* Match 1 */}
            <div className="p-3.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs space-y-2 shadow-xs">
              <div className="flex justify-between items-center text-text-primary font-medium">
                <span className="flex items-center gap-1.5 truncate">
                  <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400 shrink-0" />
                  <span className="truncate">{previewData.semi1.t1}</span>
                </span>
                <span className="font-mono font-bold text-emerald-400 ml-2">
                  {previewData.semi1.s1}
                </span>
              </div>
              <div className="flex justify-between items-center text-text-muted">
                <span className="truncate pl-5">{previewData.semi1.t2}</span>
                <span className="font-mono ml-2">{previewData.semi1.s2}</span>
              </div>
            </div>

            {/* Match 2 */}
            <div className="p-3.5 rounded-xl bg-surface-elevated border border-border-subtle text-xs space-y-2 shadow-xs">
              <div className="flex justify-between items-center text-text-primary font-medium">
                <span className="flex items-center gap-1.5 truncate">
                  <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400 shrink-0" />
                  <span className="truncate">{previewData.semi2.t1}</span>
                </span>
                <span className="font-mono font-bold text-emerald-400 ml-2">
                  {previewData.semi2.s1}
                </span>
              </div>
              <div className="flex justify-between items-center text-text-muted">
                <span className="truncate pl-5">{previewData.semi2.t2}</span>
                <span className="font-mono ml-2">{previewData.semi2.s2}</span>
              </div>
            </div>
          </div>

          {/* Spacer on Desktop (1 col) */}
          <div className="hidden md:block md:col-span-1" />

          {/* Grand Final Column (6 cols) */}
          <div className="md:col-span-6 space-y-4">
            <div className="text-[11px] uppercase tracking-wider font-mono font-semibold text-text-muted px-1 flex items-center justify-between">
              <span className="text-amber-400 font-bold">🏆 Grand Final</span>
              <span className="text-text-muted text-[10px] font-sans">
                Klik tim untuk simulasi juara ↓
              </span>
            </div>

            <div className="p-4 rounded-xl bg-surface-elevated border-2 border-border-strong text-xs space-y-3 shadow-md">
              {/* Participant A */}
              <div
                onClick={() => setSelectedWinner('teamA')}
                className={`flex justify-between items-center p-3 rounded-lg cursor-pointer transition-all ${
                  selectedWinner === 'teamA'
                    ? 'bg-emerald-500/12 border border-emerald-500/40 text-text-primary font-bold'
                    : 'bg-surface-base text-text-secondary hover:bg-surface-raised border border-border-subtle'
                }`}
              >
                <div className="flex items-center gap-2.5 min-w-0">
                  <span className="h-5 w-5 rounded bg-surface-raised border border-border-subtle text-text-muted flex items-center justify-center font-mono font-bold text-[10px] shrink-0">
                    {previewData.final.seed1}
                  </span>
                  <span className="truncate">{previewData.final.t1}</span>
                  {selectedWinner === 'teamA' && (
                    <span className="text-[10px] px-2 py-0.5 rounded-full bg-emerald-500/20 text-emerald-400 font-mono font-bold shrink-0">
                      CHAMPION
                    </span>
                  )}
                </div>
                <span className="font-mono text-sm font-bold text-emerald-400 ml-2">
                  {selectedWinner === 'teamA' ? '3' : '1'}
                </span>
              </div>

              {/* Participant B */}
              <div
                onClick={() => setSelectedWinner('teamB')}
                className={`flex justify-between items-center p-3 rounded-lg cursor-pointer transition-all ${
                  selectedWinner === 'teamB'
                    ? 'bg-emerald-500/12 border border-emerald-500/40 text-text-primary font-bold'
                    : 'bg-surface-base text-text-secondary hover:bg-surface-raised border border-border-subtle'
                }`}
              >
                <div className="flex items-center gap-2.5 min-w-0">
                  <span className="h-5 w-5 rounded bg-surface-raised border border-border-subtle text-text-muted flex items-center justify-center font-mono font-bold text-[10px] shrink-0">
                    {previewData.final.seed2}
                  </span>
                  <span className="truncate">{previewData.final.t2}</span>
                  {selectedWinner === 'teamB' && (
                    <span className="text-[10px] px-2 py-0.5 rounded-full bg-emerald-500/20 text-emerald-400 font-mono font-bold shrink-0">
                      CHAMPION
                    </span>
                  )}
                </div>
                <span className="font-mono text-sm font-bold text-emerald-400 ml-2">
                  {selectedWinner === 'teamB' ? '3' : '2'}
                </span>
              </div>

              {/* Winner Trophy banner */}
              <div className="pt-2.5 flex items-center justify-center gap-2 text-xs font-medium text-amber-400 border-t border-border-subtle">
                <Trophy className="h-4 w-4 shrink-0" />
                <span>
                  Juara Turnamen:{' '}
                  <strong className="text-text-primary font-bold">
                    {selectedWinner === 'teamA'
                      ? previewData.final.t1
                      : previewData.final.t2}
                  </strong>
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
