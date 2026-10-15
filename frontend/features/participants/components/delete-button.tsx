'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Trash2 } from 'lucide-react';
import { toast } from 'sonner';
import { apiClient } from '@/lib/api-client';
import { ConfirmDialog } from '@/components/ui/confirm-dialog';

interface DeleteParticipantButtonProps {
  tournamentId: string;
  participantId: string;
  participantName: string;
  disabled?: boolean;
}

export function DeleteParticipantButton({
  tournamentId,
  participantId,
  participantName,
  disabled = false,
}: DeleteParticipantButtonProps) {
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);

  const handleDelete = async () => {
    setIsDeleting(true);

    try {
      await apiClient.participants.delete(tournamentId, participantId);
      toast.success('Peserta Dihapus', {
        description: `"${participantName}" telah dikeluarkan dari turnamen.`,
      });

      setIsOpen(false);
      router.refresh();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal menghapus peserta.';
      toast.error('Gagal', { description: msg });
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <>
      <button
        type="button"
        onClick={() => !disabled && setIsOpen(true)}
        disabled={disabled}
        className="p-2 text-text-muted hover:text-danger hover:bg-danger/10 rounded-lg transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
        title={disabled ? 'Peserta dikunci saat turnamen aktif' : 'Hapus Peserta'}
        aria-label="Hapus Peserta"
      >
        <Trash2 size={15} />
      </button>

      <ConfirmDialog
        isOpen={isOpen}
        onClose={() => setIsOpen(false)}
        onConfirm={handleDelete}
        title="Hapus Peserta?"
        description={`Apakah Anda yakin ingin mengeluarkan peserta "${participantName}" dari turnamen ini?`}
        confirmText="Hapus Peserta"
        cancelText="Batal"
        variant="danger"
        isLoading={isDeleting}
      />
    </>
  );
}