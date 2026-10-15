'use client';

import { useState } from 'react';
import { Edit2 } from 'lucide-react';
import ScoreModal from './score-modal';
import { Match, Participant } from '@/types/api';

interface LeagueMatchCardProps {
  match: Match;
  allParticipants?: Participant[];
  isReadOnly?: boolean;
}

export default function LeagueMatchCard({
  match,
  allParticipants = [],
  isReadOnly = false,
}: LeagueMatchCardProps) {
  const [isModalOpen, setIsModalOpen] = useState(false);

  const p1 = match.participant_a;
  const p2 = match.participant_b;
  const scoreA = match.score_a ?? match.participant_a_score ?? match.scores?.a ?? 0;
  const scoreB = match.score_b ?? match.participant_b_score ?? match.scores?.b ?? 0;

  const isCompleted = match.status === 'COMPLETED';
  const isAWin = isCompleted && scoreA > scoreB;
  const isBWin = isCompleted && scoreB > scoreA;

  const isReady = !!(p1 && p2);
  const canEdit = isReady && !isReadOnly;

  return (
    <>
      <div
        onClick={() => canEdit && setIsModalOpen(true)}
        className={`relative group bg-surface-raised border p-4 rounded-xl flex items-center justify-between transition-colors ${
          canEdit
            ? 'border-border-default cursor-pointer hover:border-accent hover:bg-surface-elevated shadow-xs'
            : 'border-border-subtle cursor-default'
        }`}
      >
        {/* Tombol Edit Indicator */}
        {canEdit && (
          <div className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity">
            <div className="p-1 rounded-md bg-accent text-white shadow-xs">
              <Edit2 size={12} />
            </div>
          </div>
        )}

        {/* Tim A */}
        <div
          className={`flex-1 text-right truncate pr-4 text-sm transition-colors ${
            isAWin
              ? 'text-emerald-400 font-bold'
              : p1
              ? 'text-text-primary font-medium'
              : 'text-text-muted italic'
          }`}
        >
          {p1?.name || 'TBD'}
        </div>

        {/* Skor Capsule */}
        <div
          className={`px-3 py-1 rounded-md font-mono font-bold text-xs min-w-16 text-center border shrink-0 transition-colors ${
            isCompleted
              ? 'bg-surface-base text-text-primary border-border-default'
              : 'bg-surface-elevated text-text-muted border-border-subtle'
          }`}
        >
          {isCompleted ? `${scoreA} - ${scoreB}` : 'VS'}
        </div>

        {/* Tim B */}
        <div
          className={`flex-1 text-left truncate pl-4 text-sm transition-colors ${
            isBWin
              ? 'text-emerald-400 font-bold'
              : p2
              ? 'text-text-primary font-medium'
              : 'text-text-muted italic'
          }`}
        >
          {p2?.name || 'TBD'}
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
