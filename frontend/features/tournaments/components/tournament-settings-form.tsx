'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Save, Clock, Activity, CheckCircle2, Globe, Shield, Check } from 'lucide-react';
import { toast } from 'sonner';
import { apiClient } from '@/lib/api-client';
import { Tournament, TournamentStatus } from '@/types/api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/select';

const STATUS_OPTIONS: Array<{
  value: TournamentStatus;
  label: string;
  desc: string;
  icon: typeof Clock;
}> = [
  {
    value: 'DRAFT',
    label: 'Draft',
    desc: 'Turnamen belum dimulai, peserta masih bisa ditambah atau diubah.',
    icon: Clock,
  },
  {
    value: 'ONGOING',
    label: 'Sedang Berlangsung',
    desc: 'Bracket aktif dan skor pertandingan dapat diinput secara realtime.',
    icon: Activity,
  },
  {
    value: 'COMPLETED',
    label: 'Selesai',
    desc: 'Kompetisi berakhir dan juara/podium telah ditentukan.',
    icon: CheckCircle2,
  },
];

export default function TournamentSettingsForm({
  tournament,
}: {
  tournament: Tournament;
}) {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [title, setTitle] = useState(tournament.title);
  const [description, setDescription] = useState(tournament.description || '');
  const [status, setStatus] = useState(tournament.status);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) {
      toast.error('Nama turnamen wajib diisi');
      return;
    }

    setLoading(true);

    try {
      await apiClient.tournaments.update(tournament.id, {
        title: title.trim(),
        description: description.trim() || undefined,
        status,
      });

      toast.success('Pengaturan Disimpan', {
        description: 'Perubahan data turnamen berhasil diperbarui.',
      });

      router.refresh();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal menyimpan pengaturan.';
      toast.error('Gagal', { description: msg });
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* 1. Status Turnamen */}
      <div className="p-6 rounded-xl bg-surface-raised border border-border-default space-y-4">
        <div className="flex items-center gap-2 pb-2 border-b border-border-subtle">
          <Globe size={16} className="text-accent" />
          <h3 className="font-semibold text-sm text-text-primary">
            Status Turnamen
          </h3>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3" role="radiogroup" aria-label="Status Turnamen">
          {STATUS_OPTIONS.map((opt) => {
            const isSelected = status === opt.value;
            const IconComp = opt.icon;

            return (
              <button
                key={opt.value}
                type="button"
                role="radio"
                aria-checked={isSelected}
                onClick={() => setStatus(opt.value)}
                className={`p-4 rounded-xl text-left transition-all cursor-pointer focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent ${
                  isSelected
                    ? 'border-2 border-accent bg-accent/12 shadow-sm ring-1 ring-accent/30'
                    : 'border border-border-default bg-surface-base hover:border-border-strong hover:bg-surface-elevated/70'
                }`}
              >
                <div className="flex items-start justify-between gap-2 mb-2">
                  <div className="flex items-center gap-2">
                    <IconComp
                      size={18}
                      className={isSelected ? 'text-accent' : 'text-text-muted'}
                    />
                    <span className="font-bold text-sm text-text-primary">
                      {opt.label}
                    </span>
                  </div>
                  {isSelected && (
                    <div className="w-4 h-4 rounded-full bg-accent text-white flex items-center justify-center shrink-0">
                      <Check size={10} strokeWidth={3} />
                    </div>
                  )}
                </div>
                <p className="text-xs text-text-secondary leading-relaxed">
                  {opt.desc}
                </p>
              </button>
            );
          })}
        </div>
      </div>

      {/* 2. Informasi Dasar */}
      <div className="p-6 rounded-xl bg-surface-raised border border-border-default space-y-4">
        <div className="flex items-center gap-2 pb-2 border-b border-border-subtle">
          <Shield size={16} className="text-accent" />
          <h3 className="font-semibold text-sm text-text-primary">
            Informasi Dasar
          </h3>
        </div>

        <Input
          label="Nama Turnamen"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Nama Turnamen"
          required
        />

        <Textarea
          label="Deskripsi Turnamen (Opsional)"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Detail aturan, hadiah, atau jadwal turnamen..."
          rows={4}
        />

        <div className="flex justify-end pt-3 border-t border-border-subtle">
          <Button type="submit" variant="primary" isLoading={loading}>
            <Save size={16} />
            <span>Simpan Perubahan</span>
          </Button>
        </div>
      </div>
    </form>
  );
}