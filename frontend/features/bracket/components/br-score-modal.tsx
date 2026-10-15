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

  useEffect(() => {
    const existing = (match.scores?.results || []) as any[];
    if (existing.length > 0) {
      const mapped = participants.map((p) => {
        const found = existing.find((r) => (r.team_id || r.teamId) === p.id);
        if (found) {
          return {
            teamId: p.id,
            rank: found.rank ?? '',
            kills: found.kills ?? '',
            total: found.total ?? 0,
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
        description: `Skor Game ${match.match_number} berhasil direkam ke leaderboard.`,
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

  return (
    <Modal
      isOpen={true}
      onClose={() => !loading && onClose()}
      title={`Input Hasil Game #${match.match_number}`}
      description="Masukkan peringkat penempatan (Placement) dan jumlah eliminasi (Kills) untuk masing-masing tim."
      size="lg"
    >
      <form onSubmit={handleSave} className="space-y-4 pt-2">
        <div className="max-h-[60vh] overflow-y-auto border border-border-subtle rounded-lg">
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
                    <div className="relative">
                      <input
                        type="number"
                        min="1"
                        value={row.rank}
                        onChange={(e) =>
                          handleChange(idx, 'rank', e.target.value)
                        }
                        className="w-full bg-surface-base border border-border-default rounded-md px-2 py-1 text-center font-mono text-text-primary focus:border-accent outline-none"
                        placeholder="-"
                      />
                      {Number(row.rank) === 1 && (
                        <Trophy
                          size={12}
                          className="absolute top-2 right-1.5 text-amber-400"
                        />
                      )}
                    </div>
                  </td>
                  <td className="px-2 py-2.5">
                    <input
                      type="number"
                      min="0"
                      value={row.kills}
                      onChange={(e) =>
                        handleChange(idx, 'kills', e.target.value)
                      }
                      className="w-full bg-surface-base border border-border-default rounded-md px-2 py-1 text-center font-mono text-text-primary focus:border-accent outline-none"
                      placeholder="0"
                    />
                  </td>
                  <td className="px-4 py-2.5 text-center font-bold font-mono text-sm text-accent">
                    {row.total}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <div className="flex justify-end gap-2 pt-3 border-t border-border-subtle">
          <Button
            type="button"
            variant="secondary"
            onClick={onClose}
            disabled={loading}
          >
            Batal
          </Button>
          <Button type="submit" variant="primary" isLoading={loading}>
            <Save size={16} />
            <span>Simpan Hasil Game</span>
          </Button>
        </div>
      </form>
    </Modal>
  );
}
