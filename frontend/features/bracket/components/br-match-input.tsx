'use client';

import { useState } from 'react';
import { Edit2, CheckCircle2, Users, MapPin } from 'lucide-react';
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

  const gameNum = match.round_number || match.match_number;
  const playingGroups = match.playing_groups || (match.scores?.groups as string[] | undefined);

  return (
    <>
      <div
        onClick={() => canEdit && setIsOpen(true)}
        className={`relative p-4 rounded-xl border transition-all flex flex-col justify-between h-full select-none ${
          canEdit
            ? 'cursor-pointer hover:border-accent hover:bg-surface-elevated shadow-xs'
            : 'cursor-default'
        } ${
          isCompleted
            ? 'bg-surface-raised border-border-default'
            : 'bg-surface-raised/70 border-border-subtle'
        }`}
      >
        <div>
          <div className="flex justify-between items-center mb-2.5">
            <span className="font-bold text-sm text-text-primary">
              Game #{gameNum}
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

          <div className="space-y-1.5 mb-2">
            {match.map_name && (
              <div className="flex items-center gap-1.5 text-xs text-text-secondary">
                <MapPin size={12} className="text-accent shrink-0" />
                <span className="font-medium">Map: {match.map_name}</span>
              </div>
            )}

            {playingGroups && playingGroups.length > 0 && (
              <div className="flex items-center gap-1.5 text-xs text-text-muted">
                <Users size={12} className="text-secondary shrink-0" />
                <span>
                  Grup: <span className="font-semibold text-text-secondary">{playingGroups.join(' & ')}</span>
                </span>
              </div>
            )}
          </div>
        </div>

        {!isReadOnly && (
          <div className="text-[11px] text-text-muted mt-3 pt-2.5 border-t border-border-subtle flex items-center justify-between">
            <span>
              {isCompleted
                ? 'Klik untuk mengedit hasil skor'
                : 'Klik untuk input Placement & Kills'}
            </span>
            <Edit2 size={13} className="text-accent opacity-70 group-hover:opacity-100" />
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
