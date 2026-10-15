'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { UserPlus, User, Phone } from 'lucide-react';
import { toast } from 'sonner';
import { apiClient } from '@/lib/api-client';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

interface AddParticipantFormProps {
  tournamentId: string;
  disabled?: boolean;
}

export default function AddParticipantForm({
  tournamentId,
  disabled = false,
}: AddParticipantFormProps) {
  const router = useRouter();
  const [name, setName] = useState('');
  const [contactInfo, setContactInfo] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      toast.error('Nama peserta wajib diisi');
      return;
    }

    setLoading(true);

    try {
      await apiClient.participants.create(tournamentId, {
        name: name.trim(),
        contact_info: contactInfo.trim() || undefined,
      });

      toast.success('Peserta Ditambahkan', {
        description: `"${name}" berhasil didaftarkan ke turnamen.`,
      });

      setName('');
      setContactInfo('');
      router.refresh();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal menambahkan peserta.';
      toast.error('Gagal', { description: msg });
    } finally {
      setLoading(false);
    }
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="p-5 rounded-xl bg-surface-raised border border-border-default space-y-4"
    >
      <div className="flex items-center gap-2 pb-2 border-b border-border-subtle">
        <UserPlus size={16} className="text-accent" />
        <h3 className="font-semibold text-sm text-text-primary">
          Tambah Peserta
        </h3>
      </div>

      <Input
        label="Nama Peserta / Tim"
        value={name}
        onChange={(e) => setName(e.target.value)}
        placeholder="Contoh: RRQ Hoshi / Alex"
        required
        disabled={disabled}
      />

      <Input
        label="Kontak / ID Game (Opsional)"
        value={contactInfo}
        onChange={(e) => setContactInfo(e.target.value)}
        placeholder="WhatsApp, Discord, atau Game ID"
        disabled={disabled}
      />

      <Button
        type="submit"
        variant="primary"
        className="w-full"
        isLoading={loading}
        disabled={disabled}
      >
        <UserPlus size={16} />
        <span>Daftarkan Peserta</span>
      </Button>

      {disabled && (
        <p className="text-[11px] text-text-muted text-center italic">
          Turnamen sedang aktif/selesai. Penambahan peserta dikunci.
        </p>
      )}
    </form>
  );
}