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
  isCompletedWinner: boolean;
  isBye: boolean;
  targetPoint: { x: number; y: number };
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

  // 1. Pemisahan Bracket berbasis bracket_type dengan fallback ke round_number warisan
  const upperMatches = matches.filter(
    (m) =>
      m.bracket_type === 'UPPER' ||
      (!m.bracket_type && m.round_number > 0 && m.round_number < 999)
  );

  const lowerMatches = matches.filter(
    (m) =>
      m.bracket_type === 'LOWER' ||
      (!m.bracket_type && m.round_number < 0)
  );

  const grandFinalMatches = matches.filter(
    (m) =>
      m.bracket_type === 'GRAND_FINAL' ||
      (!m.bracket_type && m.round_number === 999)
  );

  const thirdPlaceMatches = matches.filter(
    (m) => m.bracket_type === 'THIRD_PLACE'
  );

  const hasLower = lowerMatches.length > 0;
  const hasThirdPlace = thirdPlaceMatches.length > 0;

  // Grouping per round (sorted)
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

  // 2. Kalkulasi Garis Konektor SVG Dinamis dengan Target Slot Entry Point
  const calculatePaths = useCallback(() => {
    if (!containerRef.current) return;
    const container = containerRef.current;
    const cRect = container.getBoundingClientRect();

    const newPaths: ConnectorPath[] = [];

    matches.forEach((m) => {
      if (!m.next_match_id) return;

      const sourceCardEl = document.getElementById(`match-card-${m.id}`);
      if (!sourceCardEl) return;

      // Cek apakah target slot khusus (match-slot-a / match-slot-b) ada di DOM
      let targetEl: HTMLElement | null = null;
      if (m.next_match_slot === 'A') {
        targetEl = document.getElementById(`match-slot-a-${m.next_match_id}`);
      } else if (m.next_match_slot === 'B') {
        targetEl = document.getElementById(`match-slot-b-${m.next_match_id}`);
      }

      // Fallback ke kartu target secara umum jika slot khusus tidak ditemukan
      if (!targetEl) {
        targetEl = document.getElementById(`match-card-${m.next_match_id}`);
      }

      if (sourceCardEl && targetEl) {
        const sRect = sourceCardEl.getBoundingClientRect();
        const tRect = targetEl.getBoundingClientRect();

        // Sumber: Jika sudah ada pemenang, keluar dari baris slot pemenang
        let y1 = sRect.top + sRect.height / 2 - cRect.top + container.scrollTop;
        if (m.status === 'COMPLETED' && m.winner_id) {
          const winnerSlotEl =
            m.winner_id === m.participant_a_id
              ? document.getElementById(`match-slot-a-${m.id}`)
              : m.winner_id === m.participant_b_id
              ? document.getElementById(`match-slot-b-${m.id}`)
              : null;
          if (winnerSlotEl) {
            const wRect = winnerSlotEl.getBoundingClientRect();
            y1 = wRect.top + wRect.height / 2 - cRect.top + container.scrollTop;
          }
        }

        const x1 = sRect.right - cRect.left + container.scrollLeft;

        // Titik masuk sisi kiri baris slot target
        const x2 = tRect.left - cRect.left + container.scrollLeft;
        const y2 = tRect.top + tRect.height / 2 - cRect.top + container.scrollTop;

        // Pastikan x2 > x1
        if (x2 <= x1) return;

        // Garis orthogonal ber-radius halus: Horizontal -> Radius -> Vertikal -> Radius -> Horizontal
        const midX = (x1 + x2) / 2;
        let d = '';

        if (Math.abs(y1 - y2) < 3) {
          d = `M ${x1} ${y1} H ${x2}`;
        } else {
          const radius = Math.min(8, Math.abs(midX - x1) * 0.75, Math.abs(y2 - y1) / 2);
          const dirY = y2 > y1 ? 1 : -1;
          d = `M ${x1} ${y1} L ${midX - radius} ${y1} Q ${midX} ${y1} ${midX} ${
            y1 + radius * dirY
          } L ${midX} ${y2 - radius * dirY} Q ${midX} ${y2} ${
            midX + radius
          } ${y2} L ${x2} ${y2}`;
        }

        const isCompletedWinner = m.status === 'COMPLETED' && Boolean(m.winner_id);
        const isBye = Boolean(m.is_bye);

        newPaths.push({
          id: `${m.id}->${m.next_match_id}-${m.next_match_slot || 'default'}`,
          d,
          isCompletedWinner,
          isBye,
          targetPoint: { x: x2, y: y2 },
        });
      }
    });

    setPaths(newPaths);
  }, [matches]);

  useEffect(() => {
    // Initial run with animation frame
    const rId = requestAnimationFrame(calculatePaths);
    const timer = setTimeout(calculatePaths, 150);

    const container = containerRef.current;
    let observer: ResizeObserver | null = null;

    if (container && typeof ResizeObserver !== 'undefined') {
      observer = new ResizeObserver(() => {
        requestAnimationFrame(calculatePaths);
      });
      observer.observe(container);
    }

    const handleResize = () => requestAnimationFrame(calculatePaths);
    window.addEventListener('resize', handleResize);

    return () => {
      cancelAnimationFrame(rId);
      clearTimeout(timer);
      window.removeEventListener('resize', handleResize);
      if (observer && container) {
        observer.unobserve(container);
      }
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
      if (isFinal) return totalRounds === 1 ? 'Final' : 'Final Upper Bracket';
      if (isSemi && totalRounds >= 3) return 'Semifinal UB';
      return `Babak ${roundIdx + 1}`;
    } else {
      if (isFinal) return 'Final Lower Bracket';
      if (isSemi && totalRounds >= 3) return 'Semifinal LB';
      return `Lower Babak ${roundIdx + 1}`;
    }
  };

  return (
    <div className="flex flex-col w-full space-y-6">
      {/* 1. Compact Champions Podium */}
      <ChampionsView
        tournament={tournament}
        matches={matches}
        participants={allParticipants}
        champions={champions}
      />

      {/* 2. Scrollable Bracket Canvas */}
      <div
        ref={containerRef}
        onScroll={() => requestAnimationFrame(calculatePaths)}
        className="w-full overflow-x-auto p-6 md:p-10 rounded-xl bg-surface-base border border-border-default min-h-[500px] relative select-none scrollbar-thin scrollbar-thumb-border-default hover:scrollbar-thumb-border-strong"
      >
        {/* SVG Connectors Overlay */}
        <svg
          className="absolute inset-0 pointer-events-none z-10 w-full h-full"
          style={{ minWidth: '100%', minHeight: '100%' }}
        >
          {/* Base Uncompleted or Bye Lines */}
          {paths
            .filter((p) => !p.isCompletedWinner)
            .map((p) => (
              <path
                key={p.id}
                d={p.d}
                fill="none"
                stroke={p.isBye ? 'var(--color-text-muted)' : 'var(--color-border-default)'}
                strokeWidth={p.isBye ? 1.5 : 2}
                strokeDasharray={p.isBye ? '4 3' : undefined}
                strokeLinecap="round"
                strokeLinejoin="round"
                className="transition-colors duration-200"
              />
            ))}

          {/* Highlighted Winner Progression Lines & Terminal Dots */}
          {paths
            .filter((p) => p.isCompletedWinner)
            .map((p) => (
              <g key={p.id}>
                <path
                  d={p.d}
                  fill="none"
                  stroke="var(--color-success)"
                  strokeWidth="2.5"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  className="transition-all duration-300 shadow-sm"
                />
                <circle
                  cx={p.targetPoint.x}
                  cy={p.targetPoint.y}
                  r="3.5"
                  fill="var(--color-success)"
                  className="transition-all duration-300"
                />
              </g>
            ))}
        </svg>

        <div className="flex gap-16 relative z-20 w-max">
          {/* Main Bracket Columns */}
          <div className="flex flex-col gap-14">
            {/* UPPER BRACKET */}
            <div>
              <div className="mb-4 inline-flex items-center gap-2 px-3 py-1 rounded-md bg-surface-elevated border border-border-subtle text-xs font-mono font-medium text-accent">
                <span>{hasLower ? 'UPPER BRACKET' : 'BAGAN PERTANDINGAN'}</span>
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
                        <span className="text-[11px] font-mono font-semibold text-text-muted uppercase tracking-wider bg-surface-elevated px-2.5 py-1 rounded-md border border-border-subtle">
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
                <div className="mb-4 inline-flex items-center gap-2 px-3 py-1 rounded-md bg-surface-elevated border border-border-subtle text-xs font-mono font-medium text-secondary">
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
                          <span className="text-[11px] font-mono font-semibold text-text-muted uppercase tracking-wider bg-surface-elevated px-2.5 py-1 rounded-md border border-border-subtle">
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

          {/* GRAND FINAL & BRONZE MATCH COLUMN */}
          {(grandFinalMatches.length > 0 || hasThirdPlace) && (
            <div className="flex flex-col justify-between border-l border-border-subtle pl-14 space-y-10">
              {/* Grand Final Section */}
              {grandFinalMatches.length > 0 && (
                <div className="flex flex-col justify-center flex-1">
                  <div className="text-center mb-6">
                    <span className="text-xs font-mono font-bold text-amber-400 uppercase tracking-widest px-4 py-1.5 rounded-full bg-amber-500/10 border border-amber-500/30 shadow-xs">
                      🏆 GRAND FINAL
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

              {/* 3rd Place Bronze Match Section */}
              {hasThirdPlace && (
                <div className="flex flex-col justify-center pt-8 border-t border-border-subtle/60">
                  <div className="text-center mb-5">
                    <span className="text-[11px] font-mono font-semibold text-amber-600 dark:text-amber-500 uppercase tracking-wider px-3 py-1 rounded-full bg-amber-700/10 border border-amber-600/25">
                      🥉 PEREBUTAN JUARA 3
                    </span>
                  </div>
                  <div className="space-y-4">
                    {thirdPlaceMatches.map((m) => (
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
          )}
        </div>
      </div>
    </div>
  );
}