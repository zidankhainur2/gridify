'use client';

import { useState } from 'react';
import { Edit2, CheckCircle2, Users } from 'lucide-react';
import BRScoreModal from './br-score-modal';
import { Match, Participant } from '@/types/api';

interface Props {
  match: Match;
  participants: Participant[];
  tournamentId: string;
  isReadOnly?: boolean;
}

export default function BRMatchInput({
  match,
  participants,
  tournamentId,
  isReadOnly = false,
}: Props) {
  const [isOpen, setIsOpen] = useState(false);
  const isCompleted = match.status === 'COMPLETED';
  const canEdit = !isReadOnly;

  const playingGroups = match.scores?.groups as string[] | undefined;

  return (
    <>
      <div
        onClick={() => canEdit && setIsOpen(true)}
        className={`relative p-4 rounded-xl border transition-colors flex flex-col justify-between h-full ${
          canEdit
            ? 'cursor-pointer hover:border-accent hover:bg-surface-elevated'
            : 'cursor-default'
        } ${
          isCompleted
            ? 'bg-surface-raised border-border-default'
            : 'bg-surface-raised/70 border-border-subtle'
        }`}
      >
        <div>
          <div className="flex justify-between items-center mb-2">
            <span className="font-bold text-sm text-text-primary">
              Game #{match.match_number}
            </span>
            {isCompleted ? (
              <span className="text-emerald-400 text-xs flex items-center gap-1 font-mono font-medium">
                <CheckCircle2 size={13} /> Selesai
              </span>
            ) : (
              <span className="text-text-muted text-xs bg-surface-elevated px-2 py-0.5 rounded border border-border-subtle font-mono">
                Belum Main
              </span>
            )}
          </div>

          {playingGroups && playingGroups.length > 0 && (
            <div className="flex items-center gap-1.5 mb-2">
              <Users size={12} className="text-text-muted" />
              <span className="text-xs text-accent font-medium">
                Group: {playingGroups.join(' & ')}
              </span>
            </div>
          )}
        </div>

        {!isReadOnly && (
          <div className="text-[11px] text-text-muted mt-3 pt-2 border-t border-border-subtle">
            {isCompleted
              ? 'Klik untuk mengedit hasil skor'
              : 'Klik untuk input Placement & Kills'}
          </div>
        )}

        {canEdit && (
          <div className="absolute top-4 right-4 opacity-0 group-hover:opacity-100 transition-opacity">
            <Edit2 size={14} className="text-accent" />
          </div>
        )}
      </div>

      {canEdit && isOpen && (
        <BRScoreModal
          match={match}
          participants={participants}
          tournamentId={tournamentId}
          onClose={() => setIsOpen(false)}
        />
      )}
    </>
  );
}
