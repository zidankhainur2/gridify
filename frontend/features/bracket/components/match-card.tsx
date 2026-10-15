'use client';

import { useState } from 'react';
import { Edit2, Trophy } from 'lucide-react';
import ScoreModal from './score-modal';
import { Match, Participant } from '@/types/api';

interface MatchCardProps {
  match: Match;
  allParticipants: Participant[];
  isReadOnly?: boolean;
}

export default function MatchCard({
  match,
  allParticipants,
  isReadOnly = false,
}: MatchCardProps) {
  const [isModalOpen, setIsModalOpen] = useState(false);

  const p1 = match.participant_a;
  const p2 = match.participant_b;
  const scoreA = match.participant_a_score ?? match.scores?.a ?? 0;
  const scoreB = match.participant_b_score ?? match.scores?.b ?? 0;
  const winnerId = match.winner_id;

  const isCompleted = match.status === 'COMPLETED';
  const isReady = !!(p1 && p2);
  const canEdit = !isReadOnly && isReady;

  const isP1Winner = isCompleted && winnerId === p1?.id;
  const isP2Winner = isCompleted && winnerId === p2?.id;

  return (
    <>
      <div
        id={`match-card-${match.id}`}
        data-match-id={match.id}
        onClick={() => canEdit && setIsModalOpen(true)}
        className={`w-64 rounded-lg border transition-colors select-none ${
          canEdit
            ? 'border-border-default bg-surface-raised cursor-pointer hover:border-accent hover:bg-surface-elevated shadow-xs'
            : 'border-border-subtle bg-surface-raised/80 cursor-default'
        }`}
      >
        {/* Header Match */}
        <div className="bg-surface-elevated/60 px-3 py-1.5 flex justify-between items-center text-[10px] font-mono uppercase tracking-wider border-b border-border-subtle rounded-t-lg">
          <span className="text-text-muted">Match #{match.match_number}</span>
          <span
            className={
              isCompleted
                ? 'text-emerald-400 font-bold'
                : match.status === 'ONGOING'
                ? 'text-accent font-bold animate-pulse'
                : 'text-text-muted'
            }
          >
            {isCompleted ? 'FINAL' : match.status}
          </span>
        </div>

        {/* Participant 1 */}
        <div
          className={`flex items-center justify-between px-3 py-2 border-b border-border-subtle text-xs transition-colors ${
            isP1Winner ? 'bg-accent/10 font-bold' : ''
          }`}
        >
          <div className="flex items-center gap-2 min-w-0 pr-2">
            <div
              className={`w-2 h-2 rounded-full shrink-0 ${
                isP1Winner
                  ? 'bg-accent shadow-xs'
                  : p1
                  ? 'bg-zinc-600'
                  : 'bg-zinc-800'
              }`}
            />
            <span
              className={`truncate ${
                isP1Winner
                  ? 'text-text-primary'
                  : p1
                  ? 'text-text-secondary'
                  : 'text-text-muted italic'
              }`}
            >
              {p1?.name || 'Menunggu Lawan...'}
            </span>
            {isP1Winner && (
              <Trophy size={13} className="text-amber-400 shrink-0" />
            )}
          </div>
          <span
            className={`font-mono font-bold text-xs shrink-0 ${
              isP1Winner ? 'text-accent' : 'text-text-muted'
            }`}
          >
            {scoreA}
          </span>
        </div>

        {/* Participant 2 */}
        <div
          className={`flex items-center justify-between px-3 py-2 text-xs rounded-b-lg transition-colors ${
            isP2Winner ? 'bg-accent/10 font-bold' : ''
          }`}
        >
          <div className="flex items-center gap-2 min-w-0 pr-2">
            <div
              className={`w-2 h-2 rounded-full shrink-0 ${
                isP2Winner
                  ? 'bg-accent shadow-xs'
                  : p2
                  ? 'bg-zinc-600'
                  : 'bg-zinc-800'
              }`}
            />
            <span
              className={`truncate ${
                isP2Winner
                  ? 'text-text-primary'
                  : p2
                  ? 'text-text-secondary'
                  : 'text-text-muted italic'
              }`}
            >
              {p2?.name || 'Menunggu Lawan...'}
            </span>
            {isP2Winner && (
              <Trophy size={13} className="text-amber-400 shrink-0" />
            )}
          </div>
          <span
            className={`font-mono font-bold text-xs shrink-0 ${
              isP2Winner ? 'text-accent' : 'text-text-muted'
            }`}
          >
            {scoreB}
          </span>
        </div>
      </div>

      {canEdit && (
        <ScoreModal
          match={match}
          allParticipants={allParticipants}
          isOpen={isModalOpen}
          onClose={() => setIsModalOpen(false)}
        />
      )}
    </>
  );
}