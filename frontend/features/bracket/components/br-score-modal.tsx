'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Trophy, Save } from 'lucide-react';
import { toast } from 'sonner';
import { apiClient } from '@/lib/api-client';
import { Match, Participant } from '@/types/api';
import { Modal } from '@/components/ui/modal';
import { Button } from '@/components/ui/button';

interface BRScoreModalProps {
  match: Match;
  participants: Participant[];
  tournamentId: string;
  onClose: () => void;
}

interface BRRowItem {
  teamId: string;
  rank: number | '';
  kills: number | '';
  total: number;
}

const PLACEMENT_POINTS: Record<number, number> = {
  1: 10,
  2: 6,
  3: 5,
  4: 4,
  5: 3,
  6: 2,
  7: 1,
  8: 1,
};

export default function BRScoreModal({
  match,
  participants,
  tournamentId,
  onClose,
}: BRScoreModalProps) {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [rows, setRows] = useState<BRRowItem[]>([]);

  const gameNum = match.round_number || match.match_number;

  useEffect(() => {
    // Read from match.br_results (populated by backend Go) or legacy scores.results
    const existingBR = (match.br_results || []) as any[];
    const legacyResults = (match.scores?.results || []) as any[];
    const existing = existingBR.length > 0 ? existingBR : legacyResults;

    if (existing.length > 0) {
      const mapped = participants.map((p) => {
        const found = existing.find(
          (r) => (r.participant_id || r.participant?.id || r.team_id || r.teamId) === p.id
        );
        if (found) {
          const rankVal = found.placement_rank ?? found.rank ?? '';
          const killsVal = found.kills ?? '';
          const totalVal = found.total_points ?? found.total ?? 0;
          return {
            teamId: p.id,
            rank: rankVal,
            kills: killsVal,
            total: totalVal,
          };
        }
        return { teamId: p.id, rank: '' as const, kills: '' as const, total: 0 };
      });

      setRows(
        mapped.sort((a, b) => (Number(a.rank) || 99) - (Number(b.rank) || 99))
      );
    } else {
      setRows(
        participants.map((p) => ({
          teamId: p.id,
          rank: '',
          kills: '',
          total: 0,
        }))
      );
    }
  }, [match, participants]);

  const handleChange = (
    idx: number,
    field: 'rank' | 'kills',
    valStr: string
  ) => {
    const next = [...rows];
    const val = parseInt(valStr, 10);

    if (field === 'rank') {
      next[idx].rank = valStr === '' ? '' : Math.max(1, isNaN(val) ? 0 : val);
    } else {
      next[idx].kills = valStr === '' ? '' : Math.max(0, isNaN(val) ? 0 : val);
    }

    const rankNum = Number(next[idx].rank);
    const killsNum = Number(next[idx].kills) || 0;
    const placePts = PLACEMENT_POINTS[rankNum] || 0;

    next[idx].total = placePts + killsNum;
    setRows(next);
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    const cleanResults = rows.map((r) => ({
      team_id: r.teamId,
      rank: Number(r.rank) || 0,
      kills: Number(r.kills) || 0,
      total: r.total,
    }));

    try {
      await apiClient.matches.updateBattleRoyaleMatch(
        tournamentId,
        match.id,
        cleanResults
      );

      toast.success('Hasil Putaran Disimpan!', {
        description: `Skor Game #${gameNum} berhasil direkam ke leaderboard.`,
      });

      router.refresh();
      onClose();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal menyimpan hasil Battle Royale.';
      toast.error('Gagal', { description: msg });
    } finally {
      setLoading(false);
    }
  };

  const getTeamName = (id: string) =>
    participants.find((p) => p.id === id)?.name || 'Unknown';

  const modalTitle = `Input Hasil Game #${gameNum}${match.map_name ? ` • Map ${match.map_name}` : ''}`;

  return (
    <Modal
      isOpen={true}
      onClose={() => !loading && onClose()}
      title={modalTitle}
      description="Masukkan peringkat penempatan (Placement Rank) dan jumlah eliminasi (Kills) untuk masing-masing tim."
      size="lg"
    >
      <form onSubmit={handleSave} className="space-y-4 pt-2">
        <div className="max-h-[60vh] overflow-y-auto border border-border-default rounded-lg">
          <table className="w-full text-xs text-left">
            <thead className="bg-surface-elevated text-text-muted font-mono uppercase sticky top-0 z-10 border-b border-border-subtle">
              <tr>
                <th className="px-4 py-2.5">Tim / Peserta</th>
                <th className="px-2 py-2.5 text-center w-24">Rank #</th>
                <th className="px-2 py-2.5 text-center w-24">Kills</th>
                <th className="px-4 py-2.5 text-center w-24">Total Poin</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-subtle">
              {rows.map((row, idx) => (
                <tr key={row.teamId} className="hover:bg-surface-elevated/40">
                  <td className="px-4 py-2.5 font-medium text-text-primary">
                    {getTeamName(row.teamId)}
                  </td>
                  <td className="px-2 py-2.5">
                    <input
                      type="number"
                      min="1"
                      max={participants.length}
                      value={row.rank}
                      onChange={(e) => handleChange(idx, 'rank', e.target.value)}
                      placeholder="Rank"
                      className="w-full bg-surface-base border border-border-default rounded px-2 py-1 text-center font-mono font-bold text-text-primary focus:border-accent focus:outline-none"
                    />
                  </td>
                  <td className="px-2 py-2.5">
                    <input
                      type="number"
                      min="0"
                      value={row.kills}
                      onChange={(e) => handleChange(idx, 'kills', e.target.value)}
                      placeholder="0"
                      className="w-full bg-surface-base border border-border-default rounded px-2 py-1 text-center font-mono font-bold text-rose-400 focus:border-accent focus:outline-none"
                    />
                  </td>
                  <td className="px-4 py-2.5 text-center font-mono font-bold text-accent">
                    {row.total}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <div className="flex justify-between items-center pt-2">
          <span className="text-[11px] font-mono text-text-muted">
            Format: Rank 1 = 10pts, 2 = 6pts, 3 = 5pts... + 1pt per Kill
          </span>
          <div className="flex gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={onClose}
              disabled={loading}
            >
              Batal
            </Button>
            <Button
              type="submit"
              variant="primary"
              size="sm"
              isLoading={loading}
            >
              <Save size={14} />
              Simpan Hasil
            </Button>
          </div>
        </div>
      </form>
    </Modal>
  );
}
