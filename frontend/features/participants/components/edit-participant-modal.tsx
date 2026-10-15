'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Edit2 } from 'lucide-react';
import { toast } from 'sonner';
import { apiClient } from '@/lib/api-client';
import { Participant } from '@/types/api';
import { Modal } from '@/components/ui/modal';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

interface EditParticipantModalProps {
  tournamentId: string;
  participant: Participant;
  disabled?: boolean;
}

export function EditParticipantModal({
  tournamentId,
  participant,
  disabled = false,
}: EditParticipantModalProps) {
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);
  const [name, setName] = useState(participant.name);
  const [contactInfo, setContactInfo] = useState(participant.contact_info || '');
  const [loading, setLoading] = useState(false);

  const handleOpen = () => {
    if (disabled) return;
    setName(participant.name);
    setContactInfo(participant.contact_info || '');
    setIsOpen(true);
  };

  const handleUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      toast.error('Nama peserta wajib diisi');
      return;
    }

    setLoading(true);

    try {
      await apiClient.participants.update(tournamentId, participant.id, {
        name: name.trim(),
        contact_info: contactInfo.trim() || undefined,
      });

      toast.success('Data Peserta Diperbarui', {
        description: `Perubahan pada "${name}" berhasil disimpan.`,
      });

      setIsOpen(false);
      router.refresh();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal memperbarui peserta.';
      toast.error('Gagal', { description: msg });
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <button
        type="button"
        onClick={handleOpen}
        disabled={disabled}
        className="p-2 rounded-lg text-text-muted hover:text-text-primary hover:bg-surface-elevated transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
        title={disabled ? 'Peserta dikunci saat turnamen aktif' : 'Edit Peserta'}
        aria-label="Edit Peserta"
      >
        <Edit2 size={15} />
      </button>

      <Modal
        isOpen={isOpen}
        onClose={() => setIsOpen(false)}
        title="Edit Data Peserta"
        description="Perbarui nama tim/pemain atau informasi kontak peserta."
        size="md"
      >
        <form onSubmit={handleUpdate} className="space-y-4 pt-2">
          <Input
            label="Nama Peserta / Tim"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Contoh: Sentinels / Alex"
            required
            autoFocus
          />

          <Input
            label="Informasi Kontak (Opsional)"
            value={contactInfo}
            onChange={(e) => setContactInfo(e.target.value)}
            placeholder="Nomor WhatsApp, email, atau Game ID"
          />

          <div className="flex justify-end gap-2 pt-3 border-t border-border-subtle">
            <Button
              type="button"
              variant="secondary"
              onClick={() => setIsOpen(false)}
              disabled={loading}
            >
              Batal
            </Button>
            <Button type="submit" variant="primary" isLoading={loading}>
              Simpan Perubahan
            </Button>
          </div>
        </form>
      </Modal>
    </>
  );
}
