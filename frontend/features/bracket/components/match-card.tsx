'use client';

import { useState } from 'react';
import { Trophy } from 'lucide-react';
import ScoreModal from './score-modal';
import { Match, Participant } from '@/types/api';

interface MatchCardProps {
  match: Match;
  allParticipants?: Participant[];
  slotAHint?: string;
  slotBHint?: string;
  isReadOnly?: boolean;
}

export default function MatchCard({
  match,
  allParticipants = [],
  slotAHint,
  slotBHint,
  isReadOnly = false,
}: MatchCardProps) {
  const [isModalOpen, setIsModalOpen] = useState(false);

  const p1 = match.participant_a;
  const p2 = match.participant_b;
  const scoreA = match.score_a ?? match.participant_a_score ?? match.scores?.a ?? 0;
  const scoreB = match.score_b ?? match.participant_b_score ?? match.scores?.b ?? 0;
  const winnerId = match.winner_id;

  const isCompleted = match.status === 'COMPLETED';
  const isBye = match.is_bye === true;
  const isReady = !!(p1 && p2);
  const canEdit = !isReadOnly && isReady && !isBye;

  const isP1Winner = isCompleted && winnerId === p1?.id;
  const isP2Winner = isCompleted && winnerId === p2?.id;

  return (
    <>
      <div
        id={`match-card-${match.id}`}
        data-match-id={match.id}
        onClick={() => canEdit && setIsModalOpen(true)}
        className={`w-64 rounded-lg border transition-all select-none ${
          isBye
            ? 'border-border-default/60 bg-surface-raised/60 opacity-80 cursor-default'
            : canEdit
            ? 'border-border-default bg-surface-raised cursor-pointer hover:border-accent hover:bg-surface-elevated shadow-xs'
            : 'border-border-subtle bg-surface-raised/80 cursor-default'
        }`}
      >
        {/* Header Match */}
        <div className="bg-surface-elevated/60 px-3 py-1.5 flex justify-between items-center text-[10px] font-mono uppercase tracking-wider border-b border-border-subtle rounded-t-lg">
          <span className="text-text-muted">
            {match.bracket_type === 'THIRD_PLACE'
              ? 'Bronze Match'
              : match.bracket_type === 'GRAND_FINAL'
              ? 'Grand Final'
              : `Match #${match.match_number}`}
          </span>
          <span
            className={
              isBye
                ? 'text-secondary font-bold'
                : isCompleted
                ? 'text-emerald-400 font-bold'
                : match.status === 'ONGOING'
                ? 'text-accent font-bold animate-pulse'
                : 'text-text-muted'
            }
          >
            {isBye ? 'BYE' : isCompleted ? 'FINAL' : match.status}
          </span>
        </div>

        {/* Participant 1 (Slot A) */}
        <div
          id={`match-slot-a-${match.id}`}
          className={`flex items-center justify-between px-3 py-2 border-b border-border-subtle text-xs transition-colors ${
            isP1Winner ? 'bg-accent/10 font-bold' : ''
          }`}
        >
          <div className="flex items-center gap-2 min-w-0 pr-2">
            {/* Seed Box */}
            {p1 && p1.seed_number > 0 && (
              <span className="px-1 py-0.2 rounded bg-surface-base border border-border-subtle text-[10px] font-mono text-text-muted shrink-0">
                #{p1.seed_number}
              </span>
            )}
            <div
              className={`w-2 h-2 rounded-full shrink-0 ${
                isP1Winner
                  ? 'bg-accent shadow-xs'
                  : p1
                  ? 'bg-slate-500'
                  : 'bg-slate-800'
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
              {p1 ? p1.name : slotAHint || (isBye ? 'AUTO ADVANCE' : 'Menunggu Lawan...')}
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
            {isBye ? '-' : scoreA}
          </span>
        </div>

        {/* Participant 2 (Slot B) */}
        <div
          id={`match-slot-b-${match.id}`}
          className={`flex items-center justify-between px-3 py-2 text-xs rounded-b-lg transition-colors ${
            isP2Winner ? 'bg-accent/10 font-bold' : ''
          }`}
        >
          <div className="flex items-center gap-2 min-w-0 pr-2">
            {/* Seed Box */}
            {p2 && p2.seed_number > 0 && (
              <span className="px-1 py-0.2 rounded bg-surface-base border border-border-subtle text-[10px] font-mono text-text-muted shrink-0">
                #{p2.seed_number}
              </span>
            )}
            <div
              className={`w-2 h-2 rounded-full shrink-0 ${
                isP2Winner
                  ? 'bg-accent shadow-xs'
                  : p2
                  ? 'bg-slate-500'
                  : 'bg-slate-800'
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
              {p2
                ? p2.name
                : isBye
                ? 'BYE (KOSONG)'
                : slotBHint || 'Menunggu Lawan...'}
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
            {isBye ? '-' : scoreB}
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