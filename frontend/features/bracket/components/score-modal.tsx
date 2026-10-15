'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';
import { apiClient } from '@/lib/api-client';
import { Match, Participant } from '@/types/api';
import { Modal } from '@/components/ui/modal';
import { Button } from '@/components/ui/button';

interface ScoreModalProps {
  isOpen: boolean;
  onClose: () => void;
  match: Match;
  allParticipants?: Participant[];
}

export default function ScoreModal({
  isOpen,
  onClose,
  match,
}: ScoreModalProps) {
  const router = useRouter();

  const getInitialA = () => match.score_a ?? match.participant_a_score ?? match.scores?.a ?? 0;
  const getInitialB = () => match.score_b ?? match.participant_b_score ?? match.scores?.b ?? 0;

  const [scoreA, setScoreA] = useState<number>(getInitialA());
  const [scoreB, setScoreB] = useState<number>(getInitialB());
  const [isLoading, setIsLoading] = useState(false);

  // Sync state whenever match or modal open state changes
  useEffect(() => {
    if (isOpen) {
      setScoreA(getInitialA());
      setScoreB(getInitialB());
    }
  }, [match.id, match.score_a, match.score_b, isOpen]);

  const p1Name = match.participant_a?.name || 'Peserta A';
  const p2Name = match.participant_b?.name || 'Peserta B';

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      await apiClient.matches.updateScore(match.tournament_id, match.id, {
        score_a: Number(scoreA) || 0,
        score_b: Number(scoreB) || 0,
        participant_a_score: Number(scoreA) || 0,
        participant_b_score: Number(scoreB) || 0,
      });

      toast.success('Skor Pertandingan Diperbarui!', {
        description: `${p1Name} (${scoreA}) vs (${scoreB}) ${p2Name}`,
      });

      router.refresh();
      onClose();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal memperbarui skor match.';
      toast.error('Gagal Mengupdate Skor', { description: msg });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={() => !isLoading && onClose()}
      title={`Input Skor Match #${match.match_number}`}
      description="Masukkan skor hasil pertandingan. Sistem akan otomatis memajukan pemenang ke babak berikutnya."
      size="md"
    >
      <form onSubmit={handleSave} className="space-y-6 pt-3">
        {/* Team A Card */}
        <div className="p-4 rounded-xl bg-surface-base border border-border-default flex items-center justify-between gap-4">
          <div className="flex-1 min-w-0">
            <span className="text-[11px] font-mono text-text-muted uppercase tracking-wider block mb-0.5">
              Peserta / Tim A
            </span>
            <p className="font-bold text-base text-text-primary truncate">
              {p1Name}
            </p>
          </div>
          <div className="w-24">
            <input
              type="number"
              min="0"
              value={scoreA}
              onChange={(e) => setScoreA(parseInt(e.target.value, 10) || 0)}
              disabled={isLoading}
              className="w-full bg-surface-elevated border border-border-default rounded-lg px-3 py-2 text-center text-lg font-mono font-bold text-text-primary focus:border-accent focus:ring-1 focus:ring-accent outline-none"
              placeholder="0"
              required
            />
          </div>
        </div>

        {/* VS Divider */}
        <div className="flex items-center gap-3">
          <div className="h-px bg-border-subtle flex-1" />
          <span className="text-xs font-mono font-bold text-text-muted px-2 py-0.5 rounded bg-surface-elevated border border-border-subtle">
            VS
          </span>
          <div className="h-px bg-border-subtle flex-1" />
        </div>

        {/* Team B Card */}
        <div className="p-4 rounded-xl bg-surface-base border border-border-default flex items-center justify-between gap-4">
          <div className="flex-1 min-w-0">
            <span className="text-[11px] font-mono text-text-muted uppercase tracking-wider block mb-0.5">
              Peserta / Tim B
            </span>
            <p className="font-bold text-base text-text-primary truncate">
              {p2Name}
            </p>
          </div>
          <div className="w-24">
            <input
              type="number"
              min="0"
              value={scoreB}
              onChange={(e) => setScoreB(parseInt(e.target.value, 10) || 0)}
              disabled={isLoading}
              className="w-full bg-surface-elevated border border-border-default rounded-lg px-3 py-2 text-center text-lg font-mono font-bold text-text-primary focus:border-accent focus:ring-1 focus:ring-accent outline-none"
              placeholder="0"
              required
            />
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex items-center justify-end gap-3 pt-2">
          <Button
            type="button"
            variant="outline"
            onClick={onClose}
            disabled={isLoading}
          >
            Batal
          </Button>
          <Button type="submit" variant="primary" isLoading={isLoading}>
            Simpan Hasil Pertandingan
          </Button>
        </div>
      </form>
    </Modal>
  );
}