'use client';

import { useState, useEffect, useRef, useCallback } from 'react';
import MatchCard from './match-card';
import ChampionsView from './champions-view';
import { Match, Participant, Tournament, ChampionResult, Champions } from '@/types/api';

interface BracketVisualizerProps {
  matches: Match[];
  allParticipants: Participant[];
  tournament: Tournament;
  champions?: Champions | ChampionResult[];
  isReadOnly?: boolean;
}

interface ConnectorPath {
  id: string;
  d: string;
}

export default function BracketVisualizer({
  matches,
  allParticipants,
  tournament,
  champions,
  isReadOnly = false,
}: BracketVisualizerProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [paths, setPaths] = useState<ConnectorPath[]>([]);

  // Pisahkan Bracket
  const upperMatches = matches.filter(
    (m) => m.round_number > 0 && m.round_number < 999
  );
  const lowerMatches = matches.filter((m) => m.round_number < 0);
  const grandFinalMatches = matches.filter((m) => m.round_number === 999);

  const hasLower = lowerMatches.length > 0;

  // Grouping per round
  const groupRounds = (list: Match[]) => {
    const map = new Map<number, Match[]>();
    list.forEach((m) => {
      const r = Math.abs(m.round_number);
      if (!map.has(r)) map.set(r, []);
      map.get(r)!.push(m);
    });

    return Array.from(map.entries())
      .sort(([a], [b]) => a - b)
      .map(([, mList]) => mList.sort((a, b) => a.match_number - b.match_number));
  };

  const ubRounds = groupRounds(upperMatches);
  const lbRounds = groupRounds(lowerMatches);

  // Kalkulasi Garis Konektor SVG Dinamis
  const calculatePaths = useCallback(() => {
    if (!containerRef.current) return;
    const container = containerRef.current;
    const cRect = container.getBoundingClientRect();

    const newPaths: ConnectorPath[] = [];

    matches.forEach((m) => {
      if (!m.next_match_id) return;

      const sourceEl = document.getElementById(`match-card-${m.id}`);
      const targetEl = document.getElementById(`match-card-${m.next_match_id}`);

      if (sourceEl && targetEl) {
        const sRect = sourceEl.getBoundingClientRect();
        const tRect = targetEl.getBoundingClientRect();

        // Titik keluar sisi kanan kartu sumber
        const x1 = sRect.right - cRect.left + container.scrollLeft;
        const y1 = sRect.top + sRect.height / 2 - cRect.top + container.scrollTop;

        // Titik masuk sisi kiri kartu target
        const x2 = tRect.left - cRect.left + container.scrollLeft;
        const y2 = tRect.top + tRect.height / 2 - cRect.top + container.scrollTop;

        // Garis orthogonal presisi: Horizontal -> Vertikal -> Horizontal
        const midX = (x1 + x2) / 2;
        const d = `M ${x1} ${y1} H ${midX} V ${y2} H ${x2}`;

        newPaths.push({
          id: `${m.id}->${m.next_match_id}`,
          d,
        });
      }
    });

    setPaths(newPaths);
  }, [matches]);

  useEffect(() => {
    calculatePaths();
    const timer = setTimeout(calculatePaths, 150);

    const handleResize = () => calculatePaths();
    window.addEventListener('resize', handleResize);

    return () => {
      clearTimeout(timer);
      window.removeEventListener('resize', handleResize);
    };
  }, [calculatePaths]);

  const getRoundTitle = (
    roundIdx: number,
    totalRounds: number,
    type: 'UB' | 'LB'
  ) => {
    const isFinal = roundIdx === totalRounds - 1;
    const isSemi = roundIdx === totalRounds - 2;

    if (type === 'UB') {
      if (isFinal) return 'Final Upper Bracket';
      if (isSemi && totalRounds >= 3) return 'Semifinal UB';
      return `Babak ${roundIdx + 1}`;
    } else {
      if (isFinal) return 'Final Lower Bracket';
      return `Lower Babak ${roundIdx + 1}`;
    }
  };

  return (
    <div className="flex flex-col w-full space-y-6">
      {/* 1. Champions Podium */}
      <ChampionsView
        tournament={tournament}
        matches={matches}
        participants={allParticipants}
        champions={champions}
      />

      {/* 2. Scrollable Bracket Canvas */}
      <div
        ref={containerRef}
        onScroll={calculatePaths}
        className="w-full overflow-x-auto p-6 md:p-10 rounded-xl bg-surface-base border border-border-default min-h-[500px] relative select-none"
      >
        {/* SVG Connectors Overlay */}
        <svg
          className="absolute inset-0 pointer-events-none z-10 w-full h-full"
          style={{ minWidth: '100%', minHeight: '100%' }}
        >
          {paths.map((p) => (
            <path
              key={p.id}
              d={p.d}
              fill="none"
              stroke="var(--color-border-default)"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
              className="transition-colors duration-200"
            />
          ))}
        </svg>

        <div className="flex gap-16 relative z-20 w-max">
          {/* Main Bracket Columns */}
          <div className="flex flex-col gap-16">
            {/* UPPER BRACKET */}
            <div>
              <div className="mb-4 inline-flex items-center gap-2 px-3 py-1 rounded-md bg-surface-elevated border border-border-subtle text-xs font-mono font-medium text-accent">
                <span>UPPER BRACKET</span>
              </div>

              <div className="flex gap-16">
                {ubRounds.map((roundMatches, rIdx) => {
                  const title = getRoundTitle(rIdx, ubRounds.length, 'UB');
                  return (
                    <div
                      key={`ub-${rIdx}`}
                      className="flex flex-col justify-around gap-8"
                    >
                      <div className="text-center pb-2">
                        <span className="text-[11px] font-mono text-text-muted uppercase tracking-wider bg-surface-elevated px-2 py-0.5 rounded border border-border-subtle">
                          {title}
                        </span>
                      </div>
                      <div className="flex flex-col justify-around flex-1 gap-6">
                        {roundMatches.map((m) => (
                          <MatchCard
                            key={m.id}
                            match={m}
                            allParticipants={allParticipants}
                            isReadOnly={isReadOnly}
                          />
                        ))}
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>

            {/* LOWER BRACKET (Jika Double Elimination) */}
            {hasLower && (
              <div className="pt-8 border-t border-border-subtle">
                <div className="mb-4 inline-flex items-center gap-2 px-3 py-1 rounded-md bg-surface-elevated border border-border-subtle text-xs font-mono font-medium text-rose-400">
                  <span>LOWER BRACKET</span>
                </div>

                <div className="flex gap-16">
                  {lbRounds.map((roundMatches, rIdx) => {
                    const title = getRoundTitle(rIdx, lbRounds.length, 'LB');
                    return (
                      <div
                        key={`lb-${rIdx}`}
                        className="flex flex-col justify-around gap-8"
                      >
                        <div className="text-center pb-2">
                          <span className="text-[11px] font-mono text-text-muted uppercase tracking-wider bg-surface-elevated px-2 py-0.5 rounded border border-border-subtle">
                            {title}
                          </span>
                        </div>
                        <div className="flex flex-col justify-around flex-1 gap-6">
                          {roundMatches.map((m) => (
                            <MatchCard
                              key={m.id}
                              match={m}
                              allParticipants={allParticipants}
                              isReadOnly={isReadOnly}
                            />
                          ))}
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            )}
          </div>

          {/* GRAND FINAL */}
          {grandFinalMatches.length > 0 && (
            <div className="flex flex-col justify-center border-l border-border-subtle pl-16">
              <div className="text-center mb-6">
                <span className="text-xs font-mono font-bold text-amber-400 uppercase tracking-widest px-4 py-1.5 rounded-full bg-amber-500/10 border border-amber-500/20">
                  GRAND FINAL
                </span>
              </div>
              <div className="space-y-6">
                {grandFinalMatches.map((m) => (
                  <MatchCard
                    key={m.id}
                    match={m}
                    allParticipants={allParticipants}
                    isReadOnly={isReadOnly}
                  />
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}